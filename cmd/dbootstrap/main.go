package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	catalogtoml "github.com/dnieblesdev/dniebles-bootstrap/internal/catalog/toml"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/config"
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
)

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
	if *profile == "" {
		printPlanUsage(stderr)
		fmt.Fprintln(stderr, "error: --profile is required")
		return exitUsage
	}

	catalog, err := catalogtoml.LoadFile(*catalogPath)
	if err != nil {
		fmt.Fprintf(stderr, "error: load catalog %q: %v\n", *catalogPath, err)
		return exitFailure
	}

	facts := detectEnvironmentFacts()
	installation := detectInstallationState(catalog)
	configState := detectConfigState(catalog)
	result := planning.BuildPlan(
		catalog,
		planning.PlanRequest{Profile: *profile},
		facts,
		configState,
		installation,
	)

	renderPlanResult(stdout, *profile, *catalogPath, facts, result)
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

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: dbootstrap <command> [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  plan    Build a deterministic plan for a profile")
}

func printPlanUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: dbootstrap plan --profile <name> [--catalog <path>]")
}
