package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	catalogtoml "github.com/dnieblesdev/dniebles-bootstrap/internal/catalog/toml"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/config"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/dotfiles"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/environment"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/state"
)

const (
	defaultCatalogPath = "catalog/bootstrap.toml"

	exitSuccess = 0
	exitFailure = 1
	exitUsage   = 2
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
	flags := flag.NewFlagSet("plan", flag.ContinueOnError)
	flags.SetOutput(stderr)

	profile := flags.String("profile", "", "profile name to plan")
	var resources resourceFlag
	flags.Var(&resources, "resource", "resource target as kind:name (may be repeated)")
	catalogPath := flags.String("catalog", defaultCatalogPath, "catalog TOML file path")

	if err := flags.Parse(args); err != nil {
		printPlanUsage(stderr)
		return exitUsage
	}
	if flags.NArg() > 0 {
		printPlanUsage(stderr)
		fmt.Fprintf(stderr, "error: unexpected argument %q\n", flags.Arg(0))
		return exitUsage
	}

	resourceRefs, err := parseResourceRefs(resources.values)
	if err != nil {
		printPlanUsage(stderr)
		fmt.Fprintf(stderr, "error: %v\n", err)
		return exitUsage
	}
	resourceRefs = dedupeResourceRefs(resourceRefs)

	if *profile == "" && len(resourceRefs) == 0 {
		printPlanUsage(stderr)
		fmt.Fprintln(stderr, "error: --profile or --resource is required")
		return exitUsage
	}

	catalog, err := catalogtoml.LoadFile(*catalogPath)
	if err != nil {
		fmt.Fprintf(stderr, "error: load catalog %q: %v\n", *catalogPath, err)
		return exitFailure
	}

	facts := detectEnvironmentFacts()
	installation := detectInstallationState(catalog)
	installation = mergeInstallationState(installation, detectDotfilesState(catalog))
	configState := detectConfigState(catalog)
	result := planning.BuildPlan(
		catalog,
		planning.PlanRequest{Profile: *profile, Resources: resourceRefs},
		facts,
		configState,
		installation,
	)

	renderPlanResult(stdout, *profile, resourceRefs, *catalogPath, facts, result)
	renderDiagnostics(stderr, result)
	if hasPlanningError(result) {
		return exitFailure
	}
	return exitSuccess
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
}

func printPlanUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: dbootstrap plan [--profile <name>] [--resource <kind:name>] [--catalog <path>]")
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
