package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	catalogtoml "github.com/dnieblesdev/dniebles-bootstrap/internal/catalog/toml"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/config"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/dotfiles"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/environment"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/execution"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/state"
)

const (
	defaultCatalogPath = "catalog/bootstrap.toml"

	exitSuccess = 0
	exitFailure = 1
	exitUsage   = 2
)

// applyMode describes the safety mode selected for the apply command.
type applyMode string

const (
	applyModeDefaultNonMutating applyMode = "default-non-mutating"
	applyModeDryRun             applyMode = "dry-run"
	applyModeConfirmedFuture    applyMode = "confirmed-future-noop"
)

var (
	detectEnvironmentFacts  = environment.Detect
	detectInstallationState = state.Detect
	detectConfigState       = config.Detect
	detectDotfilesState     = dotfiles.Detect
)

// resourceFlag accumulates repeated --resource values.
type resourceFlag struct {
	values []string
}

func (r *resourceFlag) String() string { return strings.Join(r.values, ", ") }
func (r *resourceFlag) Set(value string) error {
	r.values = append(r.values, value)
	return nil
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		fmt.Fprintln(stderr, "error: command is required")
		return exitUsage
	}

	switch args[0] {
	case "plan":
		return runPlan(args[1:], stdout, stderr)
	case "apply":
		return runApply(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		printUsage(stdout)
		return exitSuccess
	default:
		printUsage(stderr)
		fmt.Fprintf(stderr, "error: unknown command %q\n", args[0])
		return exitUsage
	}
}

func runPlan(args []string, stdout, stderr io.Writer) int {
	request, catalogPath, ok := parsePlanFlags("plan", args, stderr)
	if !ok {
		return exitUsage
	}

	result, facts, err := buildPlan(catalogPath, request)
	if err != nil {
		fmt.Fprintf(stderr, "error: load catalog %q: %v\n", catalogPath, err)
		return exitFailure
	}

	renderPlanResult(stdout, request.Profile, request.Resources, catalogPath, facts, result)
	renderDiagnostics(stderr, result)
	if hasPlanningError(result) {
		return exitFailure
	}
	return exitSuccess
}

func runApply(args []string, stdout, stderr io.Writer) int {
	request, catalogPath, mode, ok := parseApplyFlags(args, stderr)
	if !ok {
		return exitUsage
	}

	result, facts, err := buildPlan(catalogPath, request)
	if err != nil {
		fmt.Fprintf(stderr, "error: load catalog %q: %v\n", catalogPath, err)
		return exitFailure
	}

	if hasPlanningError(result) {
		renderPlanResult(stdout, request.Profile, request.Resources, catalogPath, facts, result)
		renderDiagnostics(stderr, result)
		return exitFailure
	}

	runner := execution.NewRunner(
		execution.NoopForKind(planning.ResourceKindTool),
		execution.NoopForKind(planning.ResourceKindRuntime),
		execution.NoopForKind(planning.ResourceKindPackage),
		execution.NoopForKind(planning.ResourceKindDotfile),
	)
	report := runner.Run(context.Background(), result.Plan)
	renderExecutionReport(stdout, mode, report)
	return exitSuccess
}

// parseApplyFlags parses the apply-specific flags and the shared plan target
// surface. It validates that conflicting safety flags are not combined and
// returns the selected apply mode along with the plan request.
func parseApplyFlags(args []string, stderr io.Writer) (planning.PlanRequest, string, applyMode, bool) {
	flags := flag.NewFlagSet("apply", flag.ContinueOnError)
	flags.SetOutput(stderr)

	profile := flags.String("profile", "", "profile name to plan")
	var resources resourceFlag
	flags.Var(&resources, "resource", "resource target as kind:name (may be repeated)")
	catalogPath := flags.String("catalog", defaultCatalogPath, "catalog TOML file path")
	dryRun := flags.Bool("dry-run", false, "run in non-mutating dry-run mode")
	yes := flags.Bool("yes", false, "reserved future confirmed-mode opt-in (currently noop)")

	if err := flags.Parse(args); err != nil {
		printApplyUsage(stderr)
		return planning.PlanRequest{}, "", "", false
	}
	if flags.NArg() > 0 {
		printApplyUsage(stderr)
		fmt.Fprintf(stderr, "error: unexpected argument %q\n", flags.Arg(0))
		return planning.PlanRequest{}, "", "", false
	}

	if *dryRun && *yes {
		printApplyUsage(stderr)
		fmt.Fprintln(stderr, "error: --dry-run and --yes cannot be combined")
		return planning.PlanRequest{}, "", "", false
	}

	resourceRefs, err := parseResourceRefs(resources.values)
	if err != nil {
		printApplyUsage(stderr)
		fmt.Fprintf(stderr, "error: %v\n", err)
		return planning.PlanRequest{}, "", "", false
	}
	resourceRefs = dedupeResourceRefs(resourceRefs)

	if *profile == "" && len(resourceRefs) == 0 {
		printApplyUsage(stderr)
		fmt.Fprintln(stderr, "error: --profile or --resource is required")
		return planning.PlanRequest{}, "", "", false
	}

	mode := applyModeDefaultNonMutating
	if *dryRun {
		mode = applyModeDryRun
	} else if *yes {
		mode = applyModeConfirmedFuture
	}

	return planning.PlanRequest{Profile: *profile, Resources: resourceRefs}, *catalogPath, mode, true
}

// parsePlanFlags parses the shared target surface used by plan and apply.
// It validates --profile, repeatable --resource, and --catalog and returns the
// assembled PlanRequest, catalog path, and a flag indicating success.
func parsePlanFlags(command string, args []string, stderr io.Writer) (planning.PlanRequest, string, bool) {
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(stderr)

	profile := flags.String("profile", "", "profile name to plan")
	var resources resourceFlag
	flags.Var(&resources, "resource", "resource target as kind:name (may be repeated)")
	catalogPath := flags.String("catalog", defaultCatalogPath, "catalog TOML file path")

	if err := flags.Parse(args); err != nil {
		printCommandUsage(command, stderr)
		return planning.PlanRequest{}, "", false
	}
	if flags.NArg() > 0 {
		printCommandUsage(command, stderr)
		fmt.Fprintf(stderr, "error: unexpected argument %q\n", flags.Arg(0))
		return planning.PlanRequest{}, "", false
	}

	resourceRefs, err := parseResourceRefs(resources.values)
	if err != nil {
		printCommandUsage(command, stderr)
		fmt.Fprintf(stderr, "error: %v\n", err)
		return planning.PlanRequest{}, "", false
	}
	resourceRefs = dedupeResourceRefs(resourceRefs)

	if *profile == "" && len(resourceRefs) == 0 {
		printCommandUsage(command, stderr)
		fmt.Fprintln(stderr, "error: --profile or --resource is required")
		return planning.PlanRequest{}, "", false
	}

	return planning.PlanRequest{Profile: *profile, Resources: resourceRefs}, *catalogPath, true
}

// buildPlan loads the catalog, runs detectors, and builds the plan for the request.
func buildPlan(catalogPath string, request planning.PlanRequest) (planning.PlanResult, planning.EnvironmentFacts, error) {
	catalog, err := catalogtoml.LoadFile(catalogPath)
	if err != nil {
		return planning.PlanResult{}, planning.EnvironmentFacts{}, err
	}

	facts := detectEnvironmentFacts()
	installation := detectInstallationState(catalog)
	installation = mergeInstallationState(installation, detectDotfilesState(catalog))
	configState := detectConfigState(catalog)
	result := planning.BuildPlan(
		catalog,
		request,
		facts,
		configState,
		installation,
	)
	return result, facts, nil
}

func hasPlanningError(result planning.PlanResult) bool {
	for _, stepResult := range result.Results {
		if stepResult.Status == planning.PlanStepStatusError {
			return true
		}
	}
	return false
}

// mergeInstallationState combines two presence maps into a new state.
// It is used at the CLI composition root to fold read-only dotfile availability
// into the existing installation state without changing the planner signature.
func mergeInstallationState(base, extra planning.InstallationState) planning.InstallationState {
	merged := planning.InstallationState{
		PresentResources: make(map[planning.ResourceRef]bool, len(base.PresentResources)+len(extra.PresentResources)),
	}
	for ref, present := range base.PresentResources {
		merged.PresentResources[ref] = present
	}
	for ref, present := range extra.PresentResources {
		if present {
			merged.PresentResources[ref] = true
		}
	}
	return merged
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: dbootstrap <command> [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  plan    Build a deterministic plan for a profile")
	fmt.Fprintln(w, "  apply   Run a dry-run execution of the plan (noop only)")
}

func printCommandUsage(command string, w io.Writer) {
	switch command {
	case "apply":
		printApplyUsage(w)
	default:
		printPlanUsage(w)
	}
}

func printPlanUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: dbootstrap plan [--profile <name>] [--resource <kind:name>] [--catalog <path>]")
}

func printApplyUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: dbootstrap apply [--profile <name>] [--resource <kind:name>] [--catalog <path>] [--dry-run] [--yes]")
}

func parseResourceRef(value string) (planning.ResourceRef, error) {
	if value == "" {
		return planning.ResourceRef{}, fmt.Errorf("resource ref is empty")
	}
	parts := strings.Split(value, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return planning.ResourceRef{}, fmt.Errorf("invalid resource ref %q: expected kind:name", value)
	}
	kind := planning.ResourceKind(parts[0])
	switch kind {
	case planning.ResourceKindTool, planning.ResourceKindRuntime, planning.ResourceKindPackage, planning.ResourceKindDotfile:
		return planning.ResourceRef{Kind: kind, Name: parts[1]}, nil
	default:
		return planning.ResourceRef{}, fmt.Errorf("unsupported resource kind %q in ref %q", parts[0], value)
	}
}

func parseResourceRefs(values []string) ([]planning.ResourceRef, error) {
	refs := make([]planning.ResourceRef, 0, len(values))
	for _, value := range values {
		ref, err := parseResourceRef(value)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

func dedupeResourceRefs(refs []planning.ResourceRef) []planning.ResourceRef {
	seen := make(map[planning.ResourceRef]bool, len(refs))
	result := make([]planning.ResourceRef, 0, len(refs))
	for _, ref := range refs {
		if seen[ref] {
			continue
		}
		seen[ref] = true
		result = append(result, ref)
	}
	return result
}
