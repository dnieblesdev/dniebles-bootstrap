package execution

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseDotlinkLinkReport(t *testing.T) {
	tests := []struct {
		name     string
		fixture  string
		selected []string
		wantErr  bool
	}{
		{name: "all changed", fixture: "all-changed.json", selected: []string{"bash"}},
		{name: "all unchanged", fixture: "all-unchanged.json", selected: []string{"bash"}},
		{name: "mixed outcomes", fixture: "mixed.json", selected: []string{"bash"}},
		{name: "failed", fixture: "failed.json", selected: []string{"bash"}},
		{name: "rolled back", fixture: "rolled-back.json", selected: []string{"bash"}},
		{name: "duplicate top level", fixture: "duplicate-top-level.json", selected: []string{"bash"}, wantErr: true},
		{name: "duplicate entry", fixture: "duplicate-entry.json", selected: []string{"bash"}, wantErr: true},
		{name: "duplicate cause", fixture: "duplicate-cause.json", selected: []string{"bash"}, wantErr: true},
		{name: "duplicate failure", fixture: "duplicate-failure.json", selected: []string{"bash"}, wantErr: true},
		{name: "duplicate rollback", fixture: "duplicate-rollback.json", selected: []string{"bash"}, wantErr: true},
		{name: "unknown field", fixture: "unknown-field.json", selected: []string{"bash"}, wantErr: true},
		{name: "schema mismatch", fixture: "schema-mismatch.json", selected: []string{"bash"}, wantErr: true},
		{name: "trailing document", fixture: "trailing.json", selected: []string{"bash"}, wantErr: true},
		{name: "malformed", fixture: "malformed.json", selected: []string{"bash"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := readDotlinkReportFixture(t, tt.fixture)
			report, err := ParseDotlinkLinkReport(stdout, tt.selected)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidDotlinkReport) {
					t.Fatalf("ParseDotlinkLinkReport() error = %v, want ErrInvalidDotlinkReport", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseDotlinkLinkReport() error = %v", err)
			}
			if report.Status == "" || len(report.Entries) == 0 {
				t.Fatalf("report = %#v, want validated report entries", report)
			}
		})
	}
}

func TestParseDotlinkLinkReportRejectsSemanticContradictions(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{name: "unselected module", json: `{"schema_version":1,"modules":["other"],"status":"success","entries":[{"module":"other","source":"a","target":"b","outcome":"changed"}],"failure":null,"rollback":{"attempted":false,"completed":false,"removed":[]}}`},
		{name: "missing failed cause", json: `{"schema_version":1,"modules":["bash"],"status":"failed","entries":[{"module":"bash","source":"a","target":"b","outcome":"failed"}],"failure":{"module":"bash","cause":{"code":"x","message":"failed"}},"rollback":{"attempted":false,"completed":false,"removed":[]}}`},
		{name: "contradictory success", json: `{"schema_version":1,"modules":["bash"],"status":"success","entries":[{"module":"bash","source":"a","target":"b","outcome":"failed","cause":{"code":"x","message":"failed"}}],"failure":null,"rollback":{"attempted":false,"completed":false,"removed":[]}}`},
		{name: "rollback completed without attempt", json: `{"schema_version":1,"modules":["bash"],"status":"failed","entries":[{"module":"bash","source":"a","target":"b","outcome":"failed","cause":{"code":"x","message":"failed"}}],"failure":{"module":"bash","cause":{"code":"x","message":"failed"}},"rollback":{"attempted":false,"completed":true,"removed":[]}}`},
		{name: "incomplete entry coverage", json: `{"schema_version":1,"modules":["bash","nvim"],"status":"success","entries":[{"module":"bash","source":"a","target":"b","outcome":"changed"}],"failure":null,"rollback":{"attempted":false,"completed":false,"removed":[]}}`},
		{name: "schema type mismatch", json: `{"schema_version":"1","modules":["bash"],"status":"success","entries":[{"module":"bash","source":"a","target":"b","outcome":"changed"}],"failure":null,"rollback":{"attempted":false,"completed":false,"removed":[]}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDotlinkLinkReport([]byte(tt.json), []string{"bash"})
			if !errors.Is(err, ErrInvalidDotlinkReport) {
				t.Fatalf("ParseDotlinkLinkReport() error = %v, want ErrInvalidDotlinkReport", err)
			}
		})
	}
}

func TestParseDotlinkLinkReportRejectsDuplicateSelectedModules(t *testing.T) {
	_, err := ParseDotlinkLinkReport(readDotlinkReportFixture(t, "all-changed.json"), []string{"bash", "bash"})
	if !errors.Is(err, ErrInvalidDotlinkReport) {
		t.Fatalf("ParseDotlinkLinkReport() error = %v, want ErrInvalidDotlinkReport", err)
	}
}

func TestInvalidReportErrorsDoNotExposeOutput(t *testing.T) {
	const secret = "sensitive report body"
	_, err := ParseDotlinkLinkReport([]byte(secret), []string{"bash"})
	if !errors.Is(err, ErrInvalidDotlinkReport) {
		t.Fatalf("ParseDotlinkLinkReport() error = %v, want ErrInvalidDotlinkReport", err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("error leaked report content: %v", err)
	}
}

func readDotlinkReportFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "dotlink-report", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}
