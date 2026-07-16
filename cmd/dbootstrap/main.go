package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	catalogtoml "github.com/dnieblesdev/dniebles-bootstrap/internal/catalog/toml"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/config"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/dotfiles"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/environment"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/execution"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/state"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/version"
)

const (
	exitSuccess = 0
	exitFailure = 1
	exitUsage   = 2
)

// catalogPathResolver resolves the default installed catalog path from XDG
// base directories, falling back to $HOME/.local/share and, last, the
// Homebrew package-share location under HOMEBREW_PREFIX.
type catalogPathResolver struct {
	LookupEnv  func(string) (string, bool)
	HomeDir    func() (string, error)
	PathExists func(string) bool
}

// Resolve returns the default catalog path. Candidates are checked in
// precedence order: XDG_DATA_HOME, $HOME/.local/share, and
// $HOMEBREW_PREFIX/share/dbootstrap/catalog/bootstrap.toml. The first
// existing candidate wins so that higher-priority catalogs always take
// precedence. If no candidate exists, the highest-priority configured path
// is returned so missing-catalog diagnostics remain useful. An empty string
// is returned when no candidate can be built.
func (r catalogPathResolver) Resolve() string {
	lookupEnv := r.LookupEnv
	if lookupEnv == nil {
		lookupEnv = os.LookupEnv
	}
	homeDir := r.HomeDir
	if homeDir == nil {
		homeDir = os.UserHomeDir
	}
	pathExists := r.PathExists
	if pathExists == nil {
		pathExists = fileExists
	}

	var candidates []string

	if value, ok := lookupEnv("XDG_DATA_HOME"); ok && value != "" {
		candidates = append(candidates, filepath.Join(value, "dbootstrap", "catalog", "bootstrap.toml"))
	}

	if home, err := homeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".local", "share", "dbootstrap", "catalog", "bootstrap.toml"))
	}

	if prefix, ok := lookupEnv("HOMEBREW_PREFIX"); ok && prefix != "" {
		candidates = append(candidates, filepath.Join(prefix, "share", "dbootstrap", "catalog", "bootstrap.toml"))
	}

	for _, candidate := range candidates {
		if pathExists(candidate) {
			return candidate
		}
	}

	if len(candidates) > 0 {
		return candidates[0]
	}
	return ""
}

// fileExists reports whether path is an existing file or directory.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

var defaultCatalogPath = func() string {
	return catalogPathResolver{LookupEnv: os.LookupEnv, HomeDir: os.UserHomeDir, PathExists: fileExists}.Resolve()
}

// applyMode describes the safety mode selected for the apply command.
type applyMode string

const (
	applyModeDefaultNonMutating applyMode = "default-non-mutating"
	applyModeDryRun             applyMode = "dry-run"
	applyModeConfirmed          applyMode = "confirmed"
	applyModeConfirmedSudo      applyMode = "confirmed-sudo"
	applyModeConfirmedAcquire   applyMode = "confirmed-acquire-homebrew"
)

var (
	detectEnvironmentFacts  = environment.Detect
	detectInstallationState = state.Detect
	detectConfigState       = config.Detect
	detectDotfilesState     = dotfiles.Detect
	brewCommandExists       = execution.BrewCommandExists
	aptCommandExists        = aptCommandExistsOnPath
	newOSCommandRunner      = func() execution.CommandRunner { return execution.NewOSCommandRunner() }
	newHomebrewInstaller    = func(kind planning.ResourceKind, runner execution.CommandRunner, exists execution.CommandExists) execution.Installer {
		return execution.NewHomebrewInstaller(kind, runner, exists)
	}
	newAptInstaller = func(kind planning.ResourceKind, runner execution.CommandRunner, exists execution.CommandExists, sudo bool) execution.Installer {
		return execution.NewAptInstaller(kind, runner, exists, sudo)
	}
	newDotfilesInstaller = func(runner execution.CommandRunner) execution.Installer {
		provider := execution.NewLocalDotfilesProvider(runner, execution.DotfilesBaseResolver{})
		return execution.NewDotfilesInstaller(provider)
	}
	acquireHomebrew = execution.AcquireHomebrew
)

func aptCommandExistsOnPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

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
	if len(args) == 1 && args[0] == "--version" {
		fmt.Fprintln(stdout, version.Version)
		return exitSuccess
	}

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
	case "bootstrap":
		return runApplyLike("bootstrap", args[1:], stdout, stderr)
	case "setup":
		return runSetup(args[1:], stdout, stderr)
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
	return runApplyLike("apply", args, stdout, stderr)
}

func runSetup(args []string, stdout, stderr io.Writer) int {
	if isHelpRequest(args) {
		printSetupUsage(stdout)
		return exitSuccess
	}

	catalogPath, mode, ok := parseSetupFlags(args, stderr)
	if !ok {
		return exitUsage
	}

	catalog, err := catalogtoml.LoadFile(catalogPath)
	if err != nil {
		fmt.Fprintf(stderr, "error: load catalog %q: %v\n", catalogPath, err)
		return exitFailure
	}
	request, err := resolveDefaultProfile(catalog)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return exitFailure
	}
	return runApplyLikeCatalog("setup", catalog, catalogPath, request, mode, stdout, stderr)
}

func runApplyLike(command string, args []string, stdout, stderr io.Writer) int {
	if command == "bootstrap" && isHelpRequest(args) {
		printApplyLikeUsage(command, stdout)
		return exitSuccess
	}

	request, catalogPath, mode, ok := parseApplyLikeFlags(command, args, stderr)
	if !ok {
		return exitUsage
	}

	catalog, err := catalogtoml.LoadFile(catalogPath)
	if err != nil {
		fmt.Fprintf(stderr, "error: load catalog %q: %v\n", catalogPath, err)
		return exitFailure
	}
	return runApplyLikeCatalog(command, catalog, catalogPath, request, mode, stdout, stderr)
}

func runApplyLikeCatalog(command string, catalog planning.Catalog, catalogPath string, request planning.PlanRequest, mode applyMode, stdout, stderr io.Writer) int {
	result, facts := buildPlanFromCatalog(catalog, request)

	if hasPlanningError(result) {
		renderPlanResult(stdout, request.Profile, request.Resources, catalogPath, facts, result)
		renderDiagnostics(stderr, result)
		return exitFailure
	}

	executionPlan := result.Plan
	if mode == applyModeConfirmedAcquire && planHasBrewBackedInstall(result.Plan) && !brewCommandExists("brew") {
		acquisition := acquireHomebrew(context.Background(), facts)
		if acquisition.Err != nil {
			fmt.Fprintf(stderr, "error: %v\n", acquisition.Err)
			return exitFailure
		}
		if acquisition.Acquired {
			fmt.Fprintln(stdout, "Homebrew acquisition complete. Re-run dbootstrap apply to install packages.")
			return exitSuccess
		}
	}
	if isConfirmedMode(mode) && planHasEligibleBrewFormulaPackage(result.Plan) {
		var presenceRunner execution.CommandRunner
		if brewCommandExists("brew") {
			presenceRunner = newOSCommandRunner()
		}
		presence := state.BrewFormulaDetector{
			CommandExists: brewCommandExists,
			Runner:        presenceRunner,
		}.Detect(context.Background(), result.Plan)
		executionPlan = state.ApplyBrewFormulaPresence(result.Plan, presence)
	}
	if isConfirmedMode(mode) && facts.OS == "linux" && planHasEligibleAptPackage(result.Plan) {
		var presenceRunner execution.CommandRunner
		if aptCommandExists("dpkg-query") {
			presenceRunner = newOSCommandRunner()
		}
		presence := state.AptPackageDetector{
			CommandExists: aptCommandExists,
			Runner:        presenceRunner,
		}.Detect(context.Background(), result.Plan)
		executionPlan = state.ApplyAptPackagePresence(executionPlan, presence)
	}
	runner := buildApplyRunner(mode, facts, executionPlan)
	if !isConfirmedMode(mode) {
		executionPlan.Steps = append([]planning.PlanStep(nil), result.Plan.Steps...)
		for index := range executionPlan.Steps {
			executionPlan.Steps[index].Status = ""
		}
	}
	report := runner.Run(context.Background(), executionPlan)
	report = appendApplyBootstrap(report, result.Plan, command, mode)
	renderExecutionReport(stdout, mode, report)
	if isConfirmedMode(mode) && hasFailedExecutionResult(report) {
		return exitFailure
	}
	return exitSuccess
}

func buildApplyRunner(mode applyMode, facts planning.EnvironmentFacts, plan planning.Plan) *execution.Runner {
	if !isConfirmedMode(mode) {
		return newNoopApplyRunner()
	}

	hasBrew := planHasBrewBackedInstall(plan)
	hasApt := planHasAptBackedInstall(plan)
	hasDotfiles := planHasDotfileSteps(plan)
	if !hasBrew && !hasApt && !hasDotfiles {
		return newProviderAwareNoopRunner()
	}

	var runner execution.CommandRunner
	commandRunner := func() execution.CommandRunner {
		if runner == nil {
			runner = newOSCommandRunner()
		}
		return runner
	}

	var brewTool, brewPackage execution.Installer
	if hasBrew {
		if brewCommandExists("brew") {
			brewTool = newHomebrewInstaller(planning.ResourceKindTool, commandRunner(), brewCommandExists)
			brewPackage = newHomebrewInstaller(planning.ResourceKindPackage, commandRunner(), brewCommandExists)
		} else {
			brewTool = missingHomebrewInstaller{kind: planning.ResourceKindTool}
			brewPackage = missingHomebrewInstaller{kind: planning.ResourceKindPackage}
		}
	}
	runtimeInstaller := execution.NoopForKind(planning.ResourceKindRuntime)
	if planHasEligibleGoRuntime(plan) {
		runtimeInstaller = goRuntimeHomebrewInstaller{
			delegate: newHomebrewInstaller(planning.ResourceKindRuntime, commandRunner(), brewCommandExists),
		}
	}
	var toolApt, packageApt execution.Installer
	if hasApt {
		if facts.OS == "linux" {
			toolApt = newAptInstaller(planning.ResourceKindTool, commandRunner(), aptCommandExists, mode == applyModeConfirmedSudo)
			packageApt = newAptInstaller(planning.ResourceKindPackage, commandRunner(), aptCommandExists, mode == applyModeConfirmedSudo)
		} else {
			toolApt = execution.NewNonLinuxAptInstaller(planning.ResourceKindTool, facts.OS)
			packageApt = execution.NewNonLinuxAptInstaller(planning.ResourceKindPackage, facts.OS)
		}
	}
	toolInstaller := execution.BrewOrAptInstaller(planning.ResourceKindTool, brewTool, toolApt)
	packageInstaller := execution.BrewOrAptInstaller(planning.ResourceKindPackage, brewPackage, packageApt)
	dotfilesInstaller := execution.NoopForKind(planning.ResourceKindDotfile)
	if hasDotfiles {
		dotfilesInstaller = newDotfilesInstaller(commandRunner())
	}
	return execution.NewRunner(
		toolInstaller,
		runtimeInstaller,
		packageInstaller,
		dotfilesInstaller,
	)
}

func newProviderAwareNoopRunner() *execution.Runner {
	return execution.NewRunner(
		execution.BrewOnlyInstaller(planning.ResourceKindTool, nil),
		execution.NoopForKind(planning.ResourceKindRuntime),
		execution.BrewOnlyInstaller(planning.ResourceKindPackage, nil),
		execution.NoopForKind(planning.ResourceKindDotfile),
	)
}

func planHasAptBackedInstall(plan planning.Plan) bool {
	for _, step := range plan.Steps {
		if step.Resource.Install != nil && step.Resource.Install.Provider == "apt" {
			return true
		}
	}
	return false
}

func isConfirmedMode(mode applyMode) bool {
	return mode == applyModeConfirmed || mode == applyModeConfirmedSudo || mode == applyModeConfirmedAcquire
}

func newNoopApplyRunner() *execution.Runner {
	return execution.NewRunner(
		execution.NoopForKind(planning.ResourceKindTool),
		execution.NoopForKind(planning.ResourceKindRuntime),
		execution.NoopForKind(planning.ResourceKindPackage),
		execution.NoopForKind(planning.ResourceKindDotfile),
	)
}

func appendApplyBootstrap(report execution.ExecutionReport, plan planning.Plan, command string, mode applyMode) execution.ExecutionReport {
	if !planHasBrewBackedInstall(plan) {
		return report
	}
	if command == "bootstrap" && !isConfirmedMode(mode) {
		return execution.AppendHomebrewBootstrap(report, plan, func(string) bool { return false })
	}
	brewExists := brewCommandExists("brew")
	return execution.AppendHomebrewBootstrap(report, plan, func(name string) bool {
		return name == "brew" && brewExists
	})
}

func planHasEligibleBrewFormulaPackage(plan planning.Plan) bool {
	for _, step := range plan.Steps {
		if step.Ref.Kind == planning.ResourceKindPackage && step.Resource.Install != nil &&
			step.Resource.Install.Provider == "brew" && strings.TrimSpace(step.Resource.Install.Package) != "" {
			return true
		}
	}
	return false
}

func planHasEligibleAptPackage(plan planning.Plan) bool {
	for _, step := range plan.Steps {
		if step.Ref.Kind == planning.ResourceKindPackage && step.Resource.Install != nil &&
			step.Resource.Install.Provider == "apt" && strings.TrimSpace(step.Resource.Install.Package) != "" {
			return true
		}
	}
	return false
}

func planHasBrewBackedInstall(plan planning.Plan) bool {
	for _, step := range plan.Steps {
		if step.Resource.Install != nil && step.Resource.Install.Provider == "brew" {
			return true
		}
	}
	return false
}

func planHasEligibleGoRuntime(plan planning.Plan) bool {
	for _, step := range plan.Steps {
		if step.Ref.Kind == planning.ResourceKindRuntime && step.Ref.Name == "go" && step.Resource.Install != nil &&
			step.Resource.Install.Provider == "brew" && step.Resource.Install.Package == "go" {
			return true
		}
	}
	return false
}

func planHasDotfileSteps(plan planning.Plan) bool {
	for _, step := range plan.Steps {
		if step.Ref.Kind == planning.ResourceKindDotfile {
			return true
		}
	}
	return false
}

func hasFailedExecutionResult(report execution.ExecutionReport) bool {
	for _, result := range report.Results {
		if result.Status == execution.StepStatusFailed {
			return true
		}
	}
	return false
}

type missingHomebrewInstaller struct {
	kind planning.ResourceKind
}

type goRuntimeHomebrewInstaller struct {
	delegate execution.Installer
}

func (i goRuntimeHomebrewInstaller) SupportedKind() planning.ResourceKind {
	return planning.ResourceKindRuntime
}

func (i goRuntimeHomebrewInstaller) Install(ctx context.Context, step planning.PlanStep) execution.StepResult {
	if step.Ref.Kind != planning.ResourceKindRuntime || step.Ref.Name != "go" || step.Resource.Install == nil ||
		step.Resource.Install.Provider != "brew" || step.Resource.Install.Package != "go" || i.delegate == nil {
		return execution.NoopForKind(planning.ResourceKindRuntime).Install(ctx, step)
	}
	return i.delegate.Install(ctx, step)
}

func (i missingHomebrewInstaller) SupportedKind() planning.ResourceKind { return i.kind }

func (i missingHomebrewInstaller) Install(_ context.Context, step planning.PlanStep) execution.StepResult {
	return execution.StepResult{
		Ref:     step.Ref,
		Status:  execution.StepStatusSkipped,
		Message: "skipped because Homebrew must be installed manually before brew-backed resources can be applied",
		Err:     execution.ErrMissingHomebrew,
	}
}

// parseApplyFlags parses the apply-specific flags and the shared plan target
// surface. It validates that conflicting safety flags are not combined and
// returns the selected apply mode along with the plan request.
func parseApplyFlags(args []string, stderr io.Writer) (planning.PlanRequest, string, applyMode, bool) {
	return parseApplyLikeFlags("apply", args, stderr)
}

func parseApplyLikeFlags(command string, args []string, stderr io.Writer) (planning.PlanRequest, string, applyMode, bool) {
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(stderr)

	profile := flags.String("profile", "", "profile name to plan")
	var resources resourceFlag
	flags.Var(&resources, "resource", "resource target as kind:name (may be repeated)")
	catalogPath := flags.String("catalog", "", "catalog TOML file path")
	dryRun := flags.Bool("dry-run", false, "run in non-mutating dry-run mode")
	yes := flags.Bool("yes", false, "confirmed mode; may run eligible Homebrew, Linux APT, and selected dotfile work")
	acquireHomebrew := flags.Bool("acquire-homebrew", false, "with --yes, acquire missing Homebrew on Linux/WSL and stop before package installation")
	sudo := flags.Bool("sudo", false, "use sudo for confirmed APT installation")

	if err := flags.Parse(args); err != nil {
		printApplyLikeUsage(command, stderr)
		return planning.PlanRequest{}, "", "", false
	}
	if flags.NArg() > 0 {
		printApplyLikeUsage(command, stderr)
		fmt.Fprintf(stderr, "error: unexpected argument %q\n", flags.Arg(0))
		return planning.PlanRequest{}, "", "", false
	}

	if *dryRun && *yes {
		printApplyLikeUsage(command, stderr)
		fmt.Fprintln(stderr, "error: --dry-run and --yes cannot be combined")
		return planning.PlanRequest{}, "", "", false
	}
	if *sudo && !*yes {
		printApplyLikeUsage(command, stderr)
		fmt.Fprintln(stderr, "error: --sudo requires --yes")
		return planning.PlanRequest{}, "", "", false
	}

	resourceRefs, err := parseResourceRefs(resources.values)
	if err != nil {
		printApplyLikeUsage(command, stderr)
		fmt.Fprintf(stderr, "error: %v\n", err)
		return planning.PlanRequest{}, "", "", false
	}
	resourceRefs = dedupeResourceRefs(resourceRefs)

	if *profile == "" && len(resourceRefs) == 0 {
		printApplyLikeUsage(command, stderr)
		fmt.Fprintln(stderr, "error: --profile or --resource is required")
		return planning.PlanRequest{}, "", "", false
	}

	mode := applyModeDefaultNonMutating
	if *dryRun {
		mode = applyModeDryRun
	} else if *yes {
		mode = applyModeConfirmed
		if *sudo {
			mode = applyModeConfirmedSudo
		}
		if *acquireHomebrew {
			mode = applyModeConfirmedAcquire
		}
	}

	if *catalogPath == "" {
		*catalogPath = defaultCatalogPath()
	}

	return planning.PlanRequest{Profile: *profile, Resources: resourceRefs}, *catalogPath, mode, true
}

func parseSetupFlags(args []string, stderr io.Writer) (string, applyMode, bool) {
	flags := flag.NewFlagSet("setup", flag.ContinueOnError)
	flags.SetOutput(stderr)

	catalogPath := flags.String("catalog", "", "catalog TOML file path")
	dryRun := flags.Bool("dry-run", false, "run in non-mutating dry-run mode")
	yes := flags.Bool("yes", false, "confirmed mode")
	acquireHomebrew := flags.Bool("acquire-homebrew", false, "with --yes, acquire missing Homebrew on Linux/WSL and stop before package installation")
	sudo := flags.Bool("sudo", false, "use sudo for confirmed APT installation")

	if err := flags.Parse(args); err != nil {
		printSetupUsage(stderr)
		return "", "", false
	}
	if flags.NArg() > 0 {
		printSetupUsage(stderr)
		fmt.Fprintf(stderr, "error: unexpected argument %q\n", flags.Arg(0))
		return "", "", false
	}
	if *dryRun && *yes {
		printSetupUsage(stderr)
		fmt.Fprintln(stderr, "error: --dry-run and --yes cannot be combined")
		return "", "", false
	}
	if *sudo && !*yes {
		printSetupUsage(stderr)
		fmt.Fprintln(stderr, "error: --sudo requires --yes")
		return "", "", false
	}

	mode := applyModeDefaultNonMutating
	if *dryRun {
		mode = applyModeDryRun
	} else if *yes {
		mode = applyModeConfirmed
		if *sudo {
			mode = applyModeConfirmedSudo
		}
		if *acquireHomebrew {
			mode = applyModeConfirmedAcquire
		}
	}
	if *catalogPath == "" {
		*catalogPath = defaultCatalogPath()
	}
	return *catalogPath, mode, true
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
	catalogPath := flags.String("catalog", "", "catalog TOML file path")

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

	if *catalogPath == "" {
		*catalogPath = defaultCatalogPath()
	}

	return planning.PlanRequest{Profile: *profile, Resources: resourceRefs}, *catalogPath, true
}

// buildPlan loads the catalog, runs detectors, and builds the plan for the request.
func buildPlan(catalogPath string, request planning.PlanRequest) (planning.PlanResult, planning.EnvironmentFacts, error) {
	catalog, err := catalogtoml.LoadFile(catalogPath)
	if err != nil {
		return planning.PlanResult{}, planning.EnvironmentFacts{}, err
	}
	result, facts := buildPlanFromCatalog(catalog, request)
	return result, facts, nil
}

func buildPlanFromCatalog(catalog planning.Catalog, request planning.PlanRequest) (planning.PlanResult, planning.EnvironmentFacts) {
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
	return result, facts
}

func resolveDefaultProfile(catalog planning.Catalog) (planning.PlanRequest, error) {
	trimmedName := strings.TrimSpace(catalog.DefaultProfile)
	if trimmedName == "" {
		return planning.PlanRequest{}, fmt.Errorf("default profile is required for setup")
	}
	if _, ok := catalog.Profiles[trimmedName]; !ok {
		return planning.PlanRequest{}, fmt.Errorf("unknown default profile %q", trimmedName)
	}
	return planning.PlanRequest{Profile: trimmedName}, nil
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
	fmt.Fprintln(w, "  apply   Execute safely; --yes may run eligible brew-backed installs, eligible Linux APT installs, and selected dotfiles")
	fmt.Fprintln(w, "          APT uses apt-get directly with --yes, or sudo apt-get only with --yes --sudo")
	fmt.Fprintln(w, "  bootstrap  Execute an explicit selection through the safe apply workflow")
	fmt.Fprintln(w, "  setup      Execute the catalog default profile through the safe apply workflow")
}

func printCommandUsage(command string, w io.Writer) {
	switch command {
	case "apply", "bootstrap":
		printApplyLikeUsage(command, w)
	case "setup":
		printSetupUsage(w)
	default:
		printPlanUsage(w)
	}
}

func printPlanUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: dbootstrap plan [--profile <name>] [--resource <kind:name>] [--catalog <path>]")
}

func printApplyUsage(w io.Writer) {
	printApplyLikeUsage("apply", w)
}

func printApplyLikeUsage(command string, w io.Writer) {
	fmt.Fprintf(w, "Usage: dbootstrap %s [--profile <name>] [--resource <kind:name>] [--catalog <path>] [--dry-run] [--yes [--sudo] [--acquire-homebrew]]\n", command)
	if command == "bootstrap" {
		fmt.Fprintln(w, "Select at least one --profile or --resource. Default and --dry-run do not mutate; --yes confirms eligible work and --sudo requires --yes.")
	}
}

func printSetupUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: dbootstrap setup [--catalog <path>] [--dry-run] [--yes [--sudo] [--acquire-homebrew]]")
}

func isHelpRequest(args []string) bool {
	return len(args) == 1 && (args[0] == "-h" || args[0] == "--help")
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
