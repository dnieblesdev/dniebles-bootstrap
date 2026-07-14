package execution

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ErrInvalidDotlinkReport identifies malformed, unsupported, or contradictory
// dotlink report data. Its errors never include command output.
var ErrInvalidDotlinkReport = errors.New("invalid dotlink link report")

type DotlinkReportStatus string

const (
	DotlinkReportStatusSuccess DotlinkReportStatus = "success"
	DotlinkReportStatusFailed  DotlinkReportStatus = "failed"
)

type DotlinkLinkOutcome string

const (
	DotlinkLinkOutcomeChanged    DotlinkLinkOutcome = "changed"
	DotlinkLinkOutcomeUnchanged  DotlinkLinkOutcome = "unchanged"
	DotlinkLinkOutcomeFailed     DotlinkLinkOutcome = "failed"
	DotlinkLinkOutcomeRolledBack DotlinkLinkOutcome = "rolled_back"
)

// DotlinkLinkReport is a validated version-1 dotlink link report. It is kept
// separate from execution result types so translation can remain a later layer.
type DotlinkLinkReport struct {
	SchemaVersion int
	Modules       []string
	Status        DotlinkReportStatus
	Entries       []DotlinkLinkEntry
	Failure       *DotlinkFailure
	Rollback      DotlinkRollback
	CommandStatus CommandStatus
}

type DotlinkLinkEntry struct {
	Module  string
	Source  string
	Target  string
	Outcome DotlinkLinkOutcome
	Cause   *DotlinkCause
}

type DotlinkCause struct {
	Code    string
	Message string
}

type DotlinkFailure struct {
	Module string
	Cause  DotlinkCause
}

type DotlinkRollback struct {
	Attempted bool
	Completed bool
	Removed   []string
}

type dotlinkWireReport struct {
	SchemaVersion int                  `json:"schema_version"`
	Modules       []string             `json:"modules"`
	Status        DotlinkReportStatus  `json:"status"`
	Entries       []dotlinkWireEntry   `json:"entries"`
	Failure       *dotlinkWireFailure  `json:"failure"`
	Rollback      *dotlinkWireRollback `json:"rollback"`
}

type dotlinkWireEntry struct {
	Module  string             `json:"module"`
	Source  string             `json:"source"`
	Target  string             `json:"target"`
	Outcome DotlinkLinkOutcome `json:"outcome"`
	Cause   *dotlinkWireCause  `json:"cause,omitempty"`
}

type dotlinkWireCause struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type dotlinkWireFailure struct {
	Module string            `json:"module"`
	Cause  *dotlinkWireCause `json:"cause"`
}

type dotlinkWireRollback struct {
	Attempted bool     `json:"attempted"`
	Completed bool     `json:"completed"`
	Removed   []string `json:"removed"`
}

// ParseDotlinkLinkReport validates one JSON v1 report. It never accepts stderr
// and returns only safe classification errors.
func ParseDotlinkLinkReport(stdout []byte, selected []string) (DotlinkLinkReport, error) {
	if err := scanDotlinkJSON(stdout); err != nil {
		return DotlinkLinkReport{}, invalidDotlinkReport(err)
	}
	if err := validateSelectedModules(selected); err != nil {
		return DotlinkLinkReport{}, invalidDotlinkReport(err)
	}

	decoder := json.NewDecoder(bytes.NewReader(stdout))
	decoder.DisallowUnknownFields()
	var wire dotlinkWireReport
	if err := decoder.Decode(&wire); err != nil {
		return DotlinkLinkReport{}, invalidDotlinkReport(err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return DotlinkLinkReport{}, invalidDotlinkReport(err)
	}
	return validateDotlinkWireReport(wire, selected)
}

func scanDotlinkJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		return errors.New("report must be an object")
	}
	if err := scanJSONObject(decoder); err != nil {
		return err
	}
	return requireJSONEOF(decoder)
}

func scanJSONObject(decoder *json.Decoder) error {
	seen := make(map[string]struct{})
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		key, ok := token.(string)
		if !ok {
			return errors.New("object key is not a string")
		}
		if _, duplicate := seen[key]; duplicate {
			return fmt.Errorf("duplicate object key %q", key)
		}
		seen[key] = struct{}{}
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '}' {
		return errors.New("unterminated object")
	}
	return nil
}

func scanJSONArray(decoder *json.Decoder) error {
	for decoder.More() {
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != ']' {
		return errors.New("unterminated array")
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		return scanJSONObject(decoder)
	case '[':
		return scanJSONArray(decoder)
	default:
		return errors.New("unexpected JSON delimiter")
	}
}

func requireJSONEOF(decoder *json.Decoder) error {
	_, err := decoder.Token()
	if err == io.EOF {
		return nil
	}
	if err == nil {
		return errors.New("trailing JSON data")
	}
	return err
}

func validateDotlinkWireReport(wire dotlinkWireReport, selected []string) (DotlinkLinkReport, error) {
	if wire.SchemaVersion != 1 || !sameModules(wire.Modules, selected) || wire.Rollback == nil {
		return DotlinkLinkReport{}, invalidDotlinkReport(errors.New("unsupported report contract"))
	}
	if wire.Status != DotlinkReportStatusSuccess && wire.Status != DotlinkReportStatusFailed {
		return DotlinkLinkReport{}, invalidDotlinkReport(errors.New("unknown report status"))
	}
	if wire.Rollback.Completed && !wire.Rollback.Attempted {
		return DotlinkLinkReport{}, invalidDotlinkReport(errors.New("rollback completed without attempt"))
	}
	if wire.Status == DotlinkReportStatusSuccess && (wire.Failure != nil || wire.Rollback.Attempted || wire.Rollback.Completed || len(wire.Rollback.Removed) != 0) {
		return DotlinkLinkReport{}, invalidDotlinkReport(errors.New("successful report has failure state"))
	}
	if wire.Status == DotlinkReportStatusFailed && !validFailure(wire.Failure, selected) {
		return DotlinkLinkReport{}, invalidDotlinkReport(errors.New("failed report lacks safe failure"))
	}

	covered := make(map[string]bool, len(selected))
	identities := make(map[string]struct{}, len(wire.Entries))
	entries := make([]DotlinkLinkEntry, 0, len(wire.Entries))
	for _, entry := range wire.Entries {
		if !containsModule(selected, entry.Module) || strings.TrimSpace(entry.Source) == "" || strings.TrimSpace(entry.Target) == "" {
			return DotlinkLinkReport{}, invalidDotlinkReport(errors.New("invalid entry identity"))
		}
		if entry.Outcome != DotlinkLinkOutcomeChanged && entry.Outcome != DotlinkLinkOutcomeUnchanged && entry.Outcome != DotlinkLinkOutcomeFailed && entry.Outcome != DotlinkLinkOutcomeRolledBack {
			return DotlinkLinkReport{}, invalidDotlinkReport(errors.New("unknown entry outcome"))
		}
		if (entry.Outcome == DotlinkLinkOutcomeFailed || entry.Outcome == DotlinkLinkOutcomeRolledBack) && !validCause(entry.Cause) {
			return DotlinkLinkReport{}, invalidDotlinkReport(errors.New("failed entry lacks safe cause"))
		}
		if wire.Status == DotlinkReportStatusSuccess && (entry.Outcome == DotlinkLinkOutcomeFailed || entry.Outcome == DotlinkLinkOutcomeRolledBack) {
			return DotlinkLinkReport{}, invalidDotlinkReport(errors.New("successful report has failed entry"))
		}
		if entry.Outcome == DotlinkLinkOutcomeRolledBack && !wire.Rollback.Attempted {
			return DotlinkLinkReport{}, invalidDotlinkReport(errors.New("rolled back entry without rollback"))
		}
		identity := entry.Module + "\x00" + entry.Source + "\x00" + entry.Target
		if _, exists := identities[identity]; exists {
			return DotlinkLinkReport{}, invalidDotlinkReport(errors.New("duplicate entry identity"))
		}
		identities[identity] = struct{}{}
		covered[entry.Module] = true
		entries = append(entries, DotlinkLinkEntry{Module: entry.Module, Source: entry.Source, Target: entry.Target, Outcome: entry.Outcome, Cause: translateCause(entry.Cause)})
	}
	for _, module := range selected {
		if !covered[module] && (wire.Failure == nil || wire.Failure.Module != module) {
			return DotlinkLinkReport{}, invalidDotlinkReport(errors.New("incomplete module coverage"))
		}
	}
	return DotlinkLinkReport{SchemaVersion: wire.SchemaVersion, Modules: append([]string(nil), wire.Modules...), Status: wire.Status, Entries: entries, Failure: translateFailure(wire.Failure), Rollback: DotlinkRollback{Attempted: wire.Rollback.Attempted, Completed: wire.Rollback.Completed, Removed: append([]string(nil), wire.Rollback.Removed...)}}, nil
}

func validateSelectedModules(selected []string) error {
	if len(selected) == 0 {
		return errors.New("no selected modules")
	}
	seen := make(map[string]struct{}, len(selected))
	for _, module := range selected {
		if strings.TrimSpace(module) == "" {
			return errors.New("empty selected module")
		}
		if _, exists := seen[module]; exists {
			return errors.New("duplicate selected module")
		}
		seen[module] = struct{}{}
	}
	return nil
}

func sameModules(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range want {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func containsModule(modules []string, module string) bool {
	for _, candidate := range modules {
		if candidate == module {
			return true
		}
	}
	return false
}

func validFailure(failure *dotlinkWireFailure, selected []string) bool {
	return failure != nil && containsModule(selected, failure.Module) && validCause(failure.Cause)
}

func validCause(cause *dotlinkWireCause) bool {
	return cause != nil && strings.TrimSpace(cause.Code) != "" && strings.TrimSpace(cause.Message) != ""
}

func translateCause(cause *dotlinkWireCause) *DotlinkCause {
	if cause == nil {
		return nil
	}
	return &DotlinkCause{Code: cause.Code, Message: cause.Message}
}

func translateFailure(failure *dotlinkWireFailure) *DotlinkFailure {
	if failure == nil {
		return nil
	}
	return &DotlinkFailure{Module: failure.Module, Cause: *translateCause(failure.Cause)}
}

func invalidDotlinkReport(err error) error {
	return errors.Join(ErrInvalidDotlinkReport, err)
}
