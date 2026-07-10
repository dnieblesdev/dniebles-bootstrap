package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/execution"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

const (
	homebrewDocumentationURL   = "https://brew.sh/"
	homebrewManualActionOutput = "- homebrew:bootstrap: Install Homebrew\n" +
		"  reason: Homebrew is required by selected resources but is not installed on this host.\n" +
		"  instruction: Review the official Homebrew installation documentation before making host changes:\n" +
		"  instruction: https://brew.sh/\n" +
		"  instruction: Install Homebrew manually only after you understand the documented steps, then re-run dbootstrap apply.\n"
)

func TestRunPlanCommand(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		installationState planning.InstallationState
		configState       planning.ConfigState
		dotfilesState     planning.InstallationState
		wantCode          int
		wantStdout        string
		wantStderr        string
	}{
		{
			name:              "success uses adapter and planner with exact output",
			args:              []string{"plan", "--profile", "dev", "--catalog", "../../catalog/bootstrap.toml"},
			installationState: planning.InstallationState{},
			configState:       planning.ConfigState{},
			wantCode:          exitSuccess,
			wantStdout: "Plan profile: dev\n" +
				"Catalog: ../../catalog/bootstrap.toml\n" +
				"Environment: os=linux arch=amd64 distro=test-distro wsl=true\n" +
				"\n" +
				"Steps:\n" +
				"1. tool:git [planned] Version control\n" +
				"   depends_on: none\n" +
				"   attention: none\n" +
				"2. package:ripgrep [planned] Fast text search\n" +
				"   depends_on: tool:git\n" +
				"   attention: none\n" +
				"3. runtime:go [attention_required] Go toolchain\n" +
				"   depends_on: tool:git\n" +
				"   attention: missing required config \"go.env\"\n" +
				"\n" +
				"Results:\n" +
				"- package:ripgrep: planned\n" +
				"- runtime:go: attention_required\n" +
				"  reason: missing required config \"go.env\"\n" +
				"- tool:git: planned\n",
			wantStderr: "",
		},
		{
			name: "present tool renders already installed",
			args: []string{"plan", "--profile", "dev", "--catalog", "../../catalog/bootstrap.toml"},
			installationState: planning.InstallationState{
				PresentResources: map[planning.ResourceRef]bool{
					{Kind: planning.ResourceKindTool, Name: "git"}: true,
				},
			},
			configState: planning.ConfigState{},
			wantCode:    exitSuccess,
			wantStdout: "Plan profile: dev\n" +
				"Catalog: ../../catalog/bootstrap.toml\n" +
				"Environment: os=linux arch=amd64 distro=test-distro wsl=true\n" +
				"\n" +
				"Steps:\n" +
				"1. tool:git [already_installed] Version control\n" +
				"   depends_on: none\n" +
				"   attention: none\n" +
				"2. package:ripgrep [planned] Fast text search\n" +
				"   depends_on: tool:git\n" +
				"   attention: none\n" +
				"3. runtime:go [attention_required] Go toolchain\n" +
				"   depends_on: tool:git\n" +
				"   attention: missing required config \"go.env\"\n" +
				"\n" +
				"Results:\n" +
				"- package:ripgrep: planned\n" +
				"- runtime:go: attention_required\n" +
				"  reason: missing required config \"go.env\"\n" +
				"- tool:git: already_installed\n",
			wantStderr: "",
		},
		{
			name:              "present config removes attention for runtime",
			args:              []string{"plan", "--profile", "dev", "--catalog", "../../catalog/bootstrap.toml"},
			installationState: planning.InstallationState{},
			configState: planning.ConfigState{
				PresentKeys: map[string]bool{"go.env": true},
			},
			wantCode: exitSuccess,
			wantStdout: "Plan profile: dev\n" +
				"Catalog: ../../catalog/bootstrap.toml\n" +
				"Environment: os=linux arch=amd64 distro=test-distro wsl=true\n" +
				"\n" +
				"Steps:\n" +
				"1. tool:git [planned] Version control\n" +
				"   depends_on: none\n" +
				"   attention: none\n" +
				"2. package:ripgrep [planned] Fast text search\n" +
				"   depends_on: tool:git\n" +
				"   attention: none\n" +
				"3. runtime:go [planned] Go toolchain\n" +
				"   depends_on: tool:git\n" +
				"   attention: none\n" +
				"\n" +
				"Results:\n" +
				"- package:ripgrep: planned\n" +
				"- runtime:go: planned\n" +
				"- tool:git: planned\n",
			wantStderr: "",
		},
		{
			name:       "missing target is a stable usage error",
			args:       []string{"plan"},
			wantCode:   exitUsage,
			wantStdout: "",
			wantStderr: "Usage: dbootstrap plan [--profile <name>] [--resource <kind:name>] [--catalog <path>]\nerror: --profile or --resource is required\n",
		},
		{
			name:              "resource only plans explicit resource",
			args:              []string{"plan", "--resource", "tool:git", "--catalog", "../../catalog/bootstrap.toml"},
			installationState: planning.InstallationState{},
			configState:       planning.ConfigState{},
			wantCode:          exitSuccess,
			wantStdout: "Plan resources: tool:git\n" +
				"Catalog: ../../catalog/bootstrap.toml\n" +
				"Environment: os=linux arch=amd64 distro=test-distro wsl=true\n" +
				"\n" +
				"Steps:\n" +
				"1. tool:git [planned] Version control\n" +
				"   depends_on: none\n" +
				"   attention: none\n" +
				"\n" +
				"Results:\n" +
				"- tool:git: planned\n",
			wantStderr: "",
		},
		{
			name:              "profile and resource union",
			args:              []string{"plan", "--profile", "dev", "--resource", "dotfile:bash", "--catalog", "../../catalog/bootstrap.toml"},
			installationState: planning.InstallationState{},
			configState:       planning.ConfigState{},
			dotfilesState:     planning.InstallationState{},
			wantCode:          exitSuccess,
			wantStdout: "Plan profile: dev\n" +
				"Catalog: ../../catalog/bootstrap.toml\n" +
				"Environment: os=linux arch=amd64 distro=test-distro wsl=true\n" +
				"\n" +
				"Steps:\n" +
				"1. dotfile:bash [planned] Bash dotfiles\n" +
				"   depends_on: none\n" +
				"   attention: none\n" +
				"2. tool:git [planned] Version control\n" +
				"   depends_on: none\n" +
				"   attention: none\n" +
				"3. package:ripgrep [planned] Fast text search\n" +
				"   depends_on: tool:git\n" +
				"   attention: none\n" +
				"4. runtime:go [attention_required] Go toolchain\n" +
				"   depends_on: tool:git\n" +
				"   attention: missing required config \"go.env\"\n" +
				"\n" +
				"Results:\n" +
				"- dotfile:bash: planned\n" +
				"- package:ripgrep: planned\n" +
				"- runtime:go: attention_required\n" +
				"  reason: missing required config \"go.env\"\n" +
				"- tool:git: planned\n",
			wantStderr: "",
		},
		{
			name:              "repeated resources are deduplicated",
			args:              []string{"plan", "--resource", "tool:git", "--resource", "tool:git", "--catalog", "../../catalog/bootstrap.toml"},
			installationState: planning.InstallationState{},
			configState:       planning.ConfigState{},
			wantCode:          exitSuccess,
			wantStdout: "Plan resources: tool:git\n" +
				"Catalog: ../../catalog/bootstrap.toml\n" +
				"Environment: os=linux arch=amd64 distro=test-distro wsl=true\n" +
				"\n" +
				"Steps:\n" +
				"1. tool:git [planned] Version control\n" +
				"   depends_on: none\n" +
				"   attention: none\n" +
				"\n" +
				"Results:\n" +
				"- tool:git: planned\n",
			wantStderr: "",
		},
		{
			name:       "malformed resource ref is rejected",
			args:       []string{"plan", "--resource", "git", "--catalog", "../../catalog/bootstrap.toml"},
			wantCode:   exitUsage,
			wantStdout: "",
			wantStderr: "Usage: dbootstrap plan [--profile <name>] [--resource <kind:name>] [--catalog <path>]\nerror: invalid resource ref \"git\": expected kind:name\n",
		},
		{
			name:       "unsupported resource kind is rejected",
			args:       []string{"plan", "--resource", "service:git", "--catalog", "../../catalog/bootstrap.toml"},
			wantCode:   exitUsage,
			wantStdout: "",
			wantStderr: "Usage: dbootstrap plan [--profile <name>] [--resource <kind:name>] [--catalog <path>]\nerror: unsupported resource kind \"service\" in ref \"service:git\"\n",
		},
		{
			name:     "unknown profile exits with diagnostics",
			args:     []string{"plan", "--profile", "missing", "--catalog", "../../catalog/bootstrap.toml"},
			wantCode: exitFailure,
			wantStdout: "Plan profile: missing\n" +
				"Catalog: ../../catalog/bootstrap.toml\n" +
				"Environment: os=linux arch=amd64 distro=test-distro wsl=true\n" +
				"\n" +
				"Steps:\n" +
				"- none\n" +
				"\n" +
				"Results:\n" +
				"- diagnostic: error\n" +
				"  reason: unknown profile \"missing\"\n",
			wantStderr: "Diagnostics:\n- diagnostic: unknown profile \"missing\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64", Distro: "test-distro", WSL: true})
			stubInstallationState(t, tt.installationState)
			stubConfigState(t, tt.configState)
			stubDotfilesState(t, tt.dotfilesState)

			gotCode := run(tt.args, &stdout, &stderr)

			if gotCode != tt.wantCode {
				t.Fatalf("run() exit code = %d, want %d", gotCode, tt.wantCode)
			}
			if got := stdout.String(); got != tt.wantStdout {
				t.Fatalf("stdout = %q, want %q", got, tt.wantStdout)
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Fatalf("stderr = %q, want %q", got, tt.wantStderr)
			}
		})
	}
}

func TestParseResourceRef(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantRef   planning.ResourceRef
		wantError string
	}{
		{
			name:    "tool ref",
			value:   "tool:git",
			wantRef: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"},
		},
		{
			name:    "runtime ref",
			value:   "runtime:go",
			wantRef: planning.ResourceRef{Kind: planning.ResourceKindRuntime, Name: "go"},
		},
		{
			name:    "package ref",
			value:   "package:ripgrep",
			wantRef: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "ripgrep"},
		},
		{
			name:    "dotfile ref",
			value:   "dotfile:bash",
			wantRef: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"},
		},
		{
			name:      "missing separator",
			value:     "git",
			wantError: `invalid resource ref "git": expected kind:name`,
		},
		{
			name:      "missing kind",
			value:     ":git",
			wantError: `invalid resource ref ":git": expected kind:name`,
		},
		{
			name:      "missing name",
			value:     "tool:",
			wantError: `invalid resource ref "tool:": expected kind:name`,
		},
		{
			name:      "too many separators",
			value:     "tool:git:extra",
			wantError: `invalid resource ref "tool:git:extra": expected kind:name`,
		},
		{
			name:      "unsupported kind",
			value:     "service:git",
			wantError: `unsupported resource kind "service" in ref "service:git"`,
		},
		{
			name:      "empty value",
			value:     "",
			wantError: "resource ref is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseResourceRef(tt.value)
			if tt.wantError != "" {
				if err == nil {
					t.Fatalf("parseResourceRef(%q) = %v, want error", tt.value, got)
				}
				if err.Error() != tt.wantError {
					t.Fatalf("parseResourceRef(%q) error = %q, want %q", tt.value, err.Error(), tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseResourceRef(%q) error = %v", tt.value, err)
			}
			if got != tt.wantRef {
				t.Fatalf("parseResourceRef(%q) = %v, want %v", tt.value, got, tt.wantRef)
			}
		})
	}
}

func TestRunPlanDotfilesPresenceReachesPlanning(t *testing.T) {
	catalogPath := writeFile(t, t.TempDir(), "catalog.toml", `
schema = "dniebles.catalog"
version = 1

[[dotfiles]]
id = "shell"
description = "Shell config"

[[profiles]]
id = "dev"
resources = ["dotfile:shell"]
`)
	stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
	stubInstallationState(t, planning.InstallationState{})
	stubConfigState(t, planning.ConfigState{})
	stubDotfilesState(t, planning.InstallationState{
		PresentResources: map[planning.ResourceRef]bool{
			{Kind: planning.ResourceKindDotfile, Name: "shell"}: true,
		},
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	gotCode := run([]string{"plan", "--profile", "dev", "--catalog", catalogPath}, &stdout, &stderr)

	if gotCode != exitSuccess {
		t.Fatalf("run() exit code = %d, want %d", gotCode, exitSuccess)
	}
	if !strings.Contains(stdout.String(), "dotfile:shell [already_installed]") {
		t.Fatalf("stdout missing already_installed dotfile: %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunPlanCatalogLoadErrors(t *testing.T) {
	tests := []struct {
		name       string
		catalog    string
		wantStderr string
	}{
		{
			name:       "missing catalog path",
			catalog:    filepath.Join(t.TempDir(), "missing.toml"),
			wantStderr: "no such file or directory",
		},
		{
			name:       "invalid catalog input",
			catalog:    writeFile(t, t.TempDir(), "invalid.toml", "[[tools]"),
			wantStderr: "toml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			gotCode := run([]string{"plan", "--profile", "dev", "--catalog", tt.catalog}, &stdout, &stderr)

			if gotCode != exitFailure {
				t.Fatalf("run() exit code = %d, want %d", gotCode, exitFailure)
			}
			if stdout.String() != "" {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if !strings.Contains(stderr.String(), "error: load catalog ") || !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Fatalf("stderr = %q, want load error containing %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

func TestRunUsageErrors(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStderr string
	}{
		{
			name: "missing command",
			args: nil,
			wantStderr: "Usage: dbootstrap <command> [options]\n" +
				"\n" +
				"Commands:\n" +
				"  plan    Build a deterministic plan for a profile\n" +
				"  apply   Execute the plan safely; only --yes may run brew-backed installs and selected dotfiles\n" +
				"error: command is required\n",
		},
		{
			name: "unknown command",
			args: []string{"deploy"},
			wantStderr: "Usage: dbootstrap <command> [options]\n" +
				"\n" +
				"Commands:\n" +
				"  plan    Build a deterministic plan for a profile\n" +
				"  apply   Execute the plan safely; only --yes may run brew-backed installs and selected dotfiles\n" +
				"error: unknown command \"deploy\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			gotCode := run(tt.args, &stdout, &stderr)

			if gotCode != exitUsage {
				t.Fatalf("run() exit code = %d, want %d", gotCode, exitUsage)
			}
			if stdout.String() != "" {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Fatalf("stderr = %q, want %q", got, tt.wantStderr)
			}
		})
	}
}

func TestRunPlanCatalogLoadErrorsSkipDetection(t *testing.T) {
	stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
	stubInstallationState(t, planning.InstallationState{})
	stubConfigState(t, planning.ConfigState{}) // safe default if somehow reached
	stubDotfilesState(t, planning.InstallationState{})

	originalConfig := detectConfigState
	detectConfigState = func(planning.Catalog) planning.ConfigState {
		t.Fatal("config detection must not run when catalog loading fails")
		return planning.ConfigState{}
	}
	t.Cleanup(func() { detectConfigState = originalConfig })

	originalDotfiles := detectDotfilesState
	detectDotfilesState = func(planning.Catalog) planning.InstallationState {
		t.Fatal("dotfiles detection must not run when catalog loading fails")
		return planning.InstallationState{}
	}
	t.Cleanup(func() { detectDotfilesState = originalDotfiles })

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	gotCode := run([]string{"plan", "--profile", "dev", "--catalog", filepath.Join(t.TempDir(), "missing.toml")}, &stdout, &stderr)

	if gotCode != exitFailure {
		t.Fatalf("run() exit code = %d, want %d", gotCode, exitFailure)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunApplyCommand(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		installationState planning.InstallationState
		configState       planning.ConfigState
		wantCode          int
		wantStdout        string
		wantStderr        string
	}{
		{
			name:              "default apply profile renders not implemented execution report",
			args:              []string{"apply", "--profile", "dev", "--catalog", "../../catalog/bootstrap.toml"},
			installationState: planning.InstallationState{},
			configState:       planning.ConfigState{},
			wantCode:          exitSuccess,
			wantStdout: "Execution Report\n" +
				"Mode: default-non-mutating\n" +
				"\n" +
				"Summary:\n" +
				"- changed: 0\n" +
				"- unchanged: 0\n" +
				"- not supported yet: 3\n" +
				"- failed: 0\n" +
				"\n" +
				"Steps:\n" +
				"1. tool:git [not supported yet] noop installer does not perform real installation\n" +
				"2. package:ripgrep [not supported yet] noop installer does not perform real installation\n" +
				"3. runtime:go [not supported yet] noop installer does not perform real installation\n" +
				"\n" +
				"Manual Actions:\n" +
				homebrewManualActionOutput,
			wantStderr: "",
		},
		{
			name:              "explicit dry run renders dry run mode",
			args:              []string{"apply", "--profile", "dev", "--catalog", "../../catalog/bootstrap.toml", "--dry-run"},
			installationState: planning.InstallationState{},
			configState:       planning.ConfigState{},
			wantCode:          exitSuccess,
			wantStdout: "Execution Report\n" +
				"Mode: dry-run\n" +
				"\n" +
				"Summary:\n" +
				"- changed: 0\n" +
				"- unchanged: 0\n" +
				"- not supported yet: 3\n" +
				"- failed: 0\n" +
				"\n" +
				"Steps:\n" +
				"1. tool:git [not supported yet] noop installer does not perform real installation\n" +
				"2. package:ripgrep [not supported yet] noop installer does not perform real installation\n" +
				"3. runtime:go [not supported yet] noop installer does not perform real installation\n" +
				"\n" +
				"Manual Actions:\n" +
				homebrewManualActionOutput,
			wantStderr: "",
		},
		{
			name:              "yes flag renders confirmed mode with missing brew guidance",
			args:              []string{"apply", "--profile", "dev", "--catalog", "../../catalog/bootstrap.toml", "--yes"},
			installationState: planning.InstallationState{},
			configState:       planning.ConfigState{},
			wantCode:          exitSuccess,
			wantStdout: "Execution Report\n" +
				"Mode: confirmed\n" +
				"Confirmed mode: brew-backed tool/package steps and selected dotfile resources may have changed this machine; runtime, non-brew, unselected, and unsupported steps remain non-mutating or not supported yet.\n" +
				"\n" +
				"Summary:\n" +
				"- changed: 0\n" +
				"- unchanged: 2\n" +
				"- not supported yet: 1\n" +
				"- failed: 0\n" +
				"\n" +
				"Steps:\n" +
				"1. tool:git [unchanged] skipped because Homebrew must be installed manually before brew-backed resources can be applied\n" +
				"2. package:ripgrep [unchanged] skipped because Homebrew must be installed manually before brew-backed resources can be applied\n" +
				"3. runtime:go [not supported yet] noop installer does not perform real installation\n" +
				"\n" +
				"Manual Actions:\n" +
				homebrewManualActionOutput,
			wantStderr: "",
		},
		{
			name:              "resource only renders single step",
			args:              []string{"apply", "--resource", "tool:git", "--catalog", "../../catalog/bootstrap.toml"},
			installationState: planning.InstallationState{},
			configState:       planning.ConfigState{},
			wantCode:          exitSuccess,
			wantStdout: "Execution Report\n" +
				"Mode: default-non-mutating\n" +
				"\n" +
				"Summary:\n" +
				"- changed: 0\n" +
				"- unchanged: 0\n" +
				"- not supported yet: 1\n" +
				"- failed: 0\n" +
				"\n" +
				"Steps:\n" +
				"1. tool:git [not supported yet] noop installer does not perform real installation\n" +
				"\n" +
				"Manual Actions:\n" +
				homebrewManualActionOutput,
			wantStderr: "",
		},
		{
			name:       "missing target is a stable usage error",
			args:       []string{"apply"},
			wantCode:   exitUsage,
			wantStdout: "",
			wantStderr: "Usage: dbootstrap apply [--profile <name>] [--resource <kind:name>] [--catalog <path>] [--dry-run] [--yes]\nerror: --profile or --resource is required\n",
		},
		{
			name:       "dry run and yes cannot be combined",
			args:       []string{"apply", "--profile", "dev", "--catalog", "../../catalog/bootstrap.toml", "--dry-run", "--yes"},
			wantCode:   exitUsage,
			wantStdout: "",
			wantStderr: "Usage: dbootstrap apply [--profile <name>] [--resource <kind:name>] [--catalog <path>] [--dry-run] [--yes]\nerror: --dry-run and --yes cannot be combined\n",
		},
		{
			name:       "malformed resource ref is rejected",
			args:       []string{"apply", "--resource", "git", "--catalog", "../../catalog/bootstrap.toml"},
			wantCode:   exitUsage,
			wantStdout: "",
			wantStderr: "Usage: dbootstrap apply [--profile <name>] [--resource <kind:name>] [--catalog <path>] [--dry-run] [--yes]\nerror: invalid resource ref \"git\": expected kind:name\n",
		},
		{
			name:     "unknown profile exits with plan diagnostics and no execution report",
			args:     []string{"apply", "--profile", "missing", "--catalog", "../../catalog/bootstrap.toml"},
			wantCode: exitFailure,
			wantStdout: "Plan profile: missing\n" +
				"Catalog: ../../catalog/bootstrap.toml\n" +
				"Environment: os=linux arch=amd64 distro=test-distro wsl=true\n" +
				"\n" +
				"Steps:\n" +
				"- none\n" +
				"\n" +
				"Results:\n" +
				"- diagnostic: error\n" +
				"  reason: unknown profile \"missing\"\n",
			wantStderr: "Diagnostics:\n- diagnostic: unknown profile \"missing\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64", Distro: "test-distro", WSL: true})
			stubInstallationState(t, tt.installationState)
			stubConfigState(t, tt.configState)
			stubDotfilesState(t, planning.InstallationState{})
			stubBrewCommandExists(t, false)

			gotCode := run(tt.args, &stdout, &stderr)

			if gotCode != tt.wantCode {
				t.Fatalf("run() exit code = %d, want %d", gotCode, tt.wantCode)
			}
			if got := stdout.String(); got != tt.wantStdout {
				t.Fatalf("stdout = %q, want %q", got, tt.wantStdout)
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Fatalf("stderr = %q, want %q", got, tt.wantStderr)
			}
		})
	}
}

func TestRunApplyHomebrewBootstrap(t *testing.T) {
	catalogPath := writeFile(t, t.TempDir(), "catalog.toml", `
schema = "dniebles.catalog"
version = 1

[[tools]]
id = "fd"
description = "Fast find alternative"
[tools.install]
provider = "brew"
package = "fd"
[tools.presence]
kind = "command_exists"
name = "fd"

[[profiles]]
id = "dev"
resources = ["tool:fd"]
`)

	tests := []struct {
		name           string
		args           []string
		brewExists     bool
		wantCode       int
		wantContains   []string
		wantNotContain string
	}{
		{
			name:       "default apply reports manual bootstrap when brew is missing",
			args:       []string{"apply", "--profile", "dev", "--catalog", catalogPath},
			brewExists: false,
			wantCode:   exitSuccess,
			wantContains: []string{
				"Execution Report",
				"Mode: default-non-mutating",
				"tool:fd [not supported yet]",
				"Manual Actions:",
				"homebrew:bootstrap: Install Homebrew",
				"Homebrew is required by selected resources",
				homebrewDocumentationURL,
			},
		},
		{
			name:       "dry run reports manual bootstrap when brew is missing",
			args:       []string{"apply", "--profile", "dev", "--catalog", catalogPath, "--dry-run"},
			brewExists: false,
			wantCode:   exitSuccess,
			wantContains: []string{
				"Execution Report",
				"Mode: dry-run",
				"Manual Actions:",
				"homebrew:bootstrap: Install Homebrew",
			},
		},
		{
			name:       "yes mode reports manual bootstrap when brew is missing",
			args:       []string{"apply", "--profile", "dev", "--catalog", catalogPath, "--yes"},
			brewExists: false,
			wantCode:   exitSuccess,
			wantContains: []string{
				"Execution Report",
				"Mode: confirmed",
				"Confirmed mode: brew-backed tool/package steps and selected dotfile resources may have changed this machine",
				"tool:fd [unchanged]",
				"Manual Actions:",
				"homebrew:bootstrap: Install Homebrew",
			},
		},
		{
			name:       "brew present does not trigger bootstrap",
			args:       []string{"apply", "--profile", "dev", "--catalog", catalogPath},
			brewExists: true,
			wantCode:   exitSuccess,
			wantContains: []string{
				"Execution Report",
				"Manual Actions:\n- none\n",
			},
			wantNotContain: "homebrew:bootstrap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
			stubInstallationState(t, planning.InstallationState{})
			stubConfigState(t, planning.ConfigState{})
			stubDotfilesState(t, planning.InstallationState{})
			originalBrew := brewCommandExists
			brewCommandExists = func(name string) bool {
				if name != "brew" {
					t.Fatalf("expected lookup for brew, got %q", name)
				}
				return tt.brewExists
			}
			t.Cleanup(func() { brewCommandExists = originalBrew })

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			gotCode := run(tt.args, &stdout, &stderr)

			if gotCode != tt.wantCode {
				t.Fatalf("run() exit code = %d, want %d", gotCode, tt.wantCode)
			}
			out := stdout.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(out, want) {
					t.Fatalf("stdout missing %q; got %q", want, out)
				}
			}
			if tt.wantNotContain != "" && strings.Contains(out, tt.wantNotContain) {
				t.Fatalf("stdout unexpectedly contains %q; got %q", tt.wantNotContain, out)
			}
			if stderr.String() != "" {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
	}
}

func TestRunApplySafeModesDoNotInstantiateRealExecution(t *testing.T) {
	catalogPath := writeFile(t, t.TempDir(), "catalog.toml", `
schema = "dniebles.catalog"
version = 1

[[tools]]
id = "fd"
description = "Fast find alternative"
[tools.install]
provider = "brew"
package = "fd"

[[dotfiles]]
id = "bash"
description = "Bash dotfiles"

[[profiles]]
id = "dev"
resources = ["tool:fd", "dotfile:bash"]
`)

	for _, args := range [][]string{
		{"apply", "--profile", "dev", "--catalog", catalogPath},
		{"apply", "--profile", "dev", "--catalog", catalogPath, "--dry-run"},
		{"plan", "--resource", "dotfile:bash", "--catalog", catalogPath},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
			stubInstallationState(t, planning.InstallationState{})
			stubConfigState(t, planning.ConfigState{})
			stubDotfilesState(t, planning.InstallationState{})
			stubBrewCommandExists(t, false)
			stubExecutionFactories(t,
				func() execution.CommandRunner {
					t.Fatal("safe modes must not instantiate OS command runners")
					return nil
				},
				func(planning.ResourceKind, execution.CommandRunner, execution.CommandExists) execution.Installer {
					t.Fatal("safe modes must not instantiate Homebrew installers")
					return nil
				},
				func(execution.CommandRunner) execution.Installer {
					t.Fatal("safe modes and plan must not instantiate dotfiles installers")
					return nil
				},
			)

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			gotCode := run(args, &stdout, &stderr)

			if gotCode != exitSuccess {
				t.Fatalf("run() exit code = %d, want %d", gotCode, exitSuccess)
			}
			if strings.Contains(stdout.String(), "installed fd") {
				t.Fatalf("safe mode output reported install: %q", stdout.String())
			}
			if args[0] == "apply" && !strings.Contains(stdout.String(), "dotfile:bash [not supported yet]") {
				t.Fatalf("safe apply did not keep dotfile resource not supported: %q", stdout.String())
			}
			if stderr.String() != "" {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
	}
}

func TestRunApplyConfirmedBrewPresentUsesInjectedRunnerForBrewOnly(t *testing.T) {
	catalogPath := writeFile(t, t.TempDir(), "catalog.toml", `
schema = "dniebles.catalog"
version = 1

[[tools]]
id = "fd"
description = "Fast find alternative"
[tools.install]
provider = "brew"
package = "fd"

[[packages]]
id = "ripgrep"
description = "Fast text search"
[packages.install]
provider = "apt"
package = "ripgrep"

[[runtimes]]
id = "go"
description = "Go toolchain"
[runtimes.install]
provider = "brew"
package = "go"

[[profiles]]
id = "dev"
resources = ["tool:fd", "package:ripgrep", "runtime:go"]
`)
	stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
	stubInstallationState(t, planning.InstallationState{})
	stubConfigState(t, planning.ConfigState{})
	stubDotfilesState(t, planning.InstallationState{})
	stubBrewCommandExists(t, true)
	runner := &recordingCommandRunner{result: execution.CommandResult{Status: execution.CommandStatusSucceeded, ExitCode: 0}}
	stubExecutionFactories(t,
		func() execution.CommandRunner { return runner },
		func(kind planning.ResourceKind, commandRunner execution.CommandRunner, exists execution.CommandExists) execution.Installer {
			return execution.NewHomebrewInstaller(kind, commandRunner, exists)
		},
		func(execution.CommandRunner) execution.Installer {
			t.Fatal("brew-only apply must not instantiate dotfiles installer")
			return nil
		},
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	gotCode := run([]string{"apply", "--profile", "dev", "--catalog", catalogPath, "--yes"}, &stdout, &stderr)

	if gotCode != exitSuccess {
		t.Fatalf("run() exit code = %d, want %d", gotCode, exitSuccess)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("command runner calls = %d, want 1", len(runner.calls))
	}
	if got := runner.calls[0]; got.Executable != "brew" || strings.Join(got.Args, " ") != "install fd" {
		t.Fatalf("CommandRequest = %#v, want brew install fd", got)
	}
	out := stdout.String()
	for _, want := range []string{
		"Mode: confirmed",
		"tool:fd [changed] installed fd with Homebrew",
		"package:ripgrep [not supported yet] no brew install metadata for this resource",
		"runtime:go [not supported yet] noop installer does not perform real installation",
		"Manual Actions:\n- none\n",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout missing %q; got %q", want, out)
		}
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunApplyConfirmedDotfilesUsesInjectedRunner(t *testing.T) {
	base := makeDotfilesBase(t, "bash")
	catalogPath := writeDotfilesCatalog(t, "bash", "git")
	stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
	stubInstallationState(t, planning.InstallationState{})
	stubConfigState(t, planning.ConfigState{})
	stubDotfilesState(t, planning.InstallationState{})
	stubBrewCommandExists(t, true)
	runner := &recordingCommandRunner{result: execution.CommandResult{Status: execution.CommandStatusSucceeded, ExitCode: 0}}
	stubExecutionFactories(t,
		func() execution.CommandRunner { return runner },
		func(kind planning.ResourceKind, commandRunner execution.CommandRunner, exists execution.CommandExists) execution.Installer {
			return execution.NewHomebrewInstaller(kind, commandRunner, exists)
		},
		func(commandRunner execution.CommandRunner) execution.Installer {
			provider := execution.NewLocalDotfilesProvider(commandRunner, execution.DotfilesBaseResolver{
				LookupEnv: func(string) (string, bool) { return base, true },
				HomeDir:   func() (string, error) { return filepath.Dir(base), nil },
			})
			return execution.NewDotfilesInstaller(provider)
		},
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	gotCode := run([]string{"apply", "--yes", "--resource", "dotfile:bash", "--catalog", catalogPath}, &stdout, &stderr)

	if gotCode != exitSuccess {
		t.Fatalf("run() exit code = %d, want %d", gotCode, exitSuccess)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("command runner calls = %d, want 1", len(runner.calls))
	}
	call := runner.calls[0]
	if call.Executable != filepath.Join(base, "bin", "dotlink") || strings.Join(call.Args, " ") != "link bash" || call.Dir != base {
		t.Fatalf("CommandRequest = %#v, want dotlink for bash only", call)
	}
	out := stdout.String()
	for _, want := range []string{
		"Mode: confirmed",
		"Confirmed mode: brew-backed tool/package steps and selected dotfile resources may have changed this machine",
		"dotfile:bash [changed] installed dotfile module bash",
		"dotfiles base: " + base,
		"source: env",
		"modules: bash",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout missing %q; got %q", want, out)
		}
	}
	if strings.Contains(out, "git") {
		t.Fatalf("stdout mentioned unselected module/resource: %q", out)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunApplyConfirmedDotfilesFailuresExitNonZero(t *testing.T) {
	tests := []struct {
		name        string
		module      string
		baseSetup   func(t *testing.T) string
		runner      *recordingCommandRunner
		wantMessage string
	}{
		{
			name:        "missing base",
			module:      "bash",
			baseSetup:   func(t *testing.T) string { return filepath.Join(t.TempDir(), "missing-dotfiles") },
			runner:      &recordingCommandRunner{result: execution.CommandResult{Status: execution.CommandStatusSucceeded, ExitCode: 0}},
			wantMessage: "resolve dotfiles base",
		},
		{
			name:   "missing dotlink",
			module: "bash",
			baseSetup: func(t *testing.T) string {
				base := makeDotfilesBase(t, "bash")
				if err := os.Remove(filepath.Join(base, "bin", "dotlink")); err != nil {
					t.Fatalf("remove dotlink: %v", err)
				}
				return base
			},
			runner:      &recordingCommandRunner{result: execution.CommandResult{Status: execution.CommandStatusSucceeded, ExitCode: 0}},
			wantMessage: "validate dotlink",
		},
		{
			name:        "missing module",
			module:      "zsh",
			baseSetup:   func(t *testing.T) string { return makeDotfilesBase(t, "bash") },
			runner:      &recordingCommandRunner{result: execution.CommandResult{Status: execution.CommandStatusSucceeded, ExitCode: 0}},
			wantMessage: "validate module \"zsh\"",
		},
		{
			name:      "runner failure",
			module:    "bash",
			baseSetup: func(t *testing.T) string { return makeDotfilesBase(t, "bash") },
			runner: &recordingCommandRunner{result: execution.CommandResult{
				Status:   execution.CommandStatusFailed,
				ExitCode: 42,
				Err:      errors.New("dotlink failed"),
			}},
			wantMessage: "dotfile module bash failed",
		},
		{
			name:      "runner timeout",
			module:    "bash",
			baseSetup: func(t *testing.T) string { return makeDotfilesBase(t, "bash") },
			runner: &recordingCommandRunner{result: execution.CommandResult{
				Status: execution.CommandStatusTimedOut,
				Err:    context.DeadlineExceeded,
			}},
			wantMessage: "dotfile module bash failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := tt.baseSetup(t)
			catalogPath := writeDotfilesCatalog(t, tt.module)
			stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
			stubInstallationState(t, planning.InstallationState{})
			stubConfigState(t, planning.ConfigState{})
			stubDotfilesState(t, planning.InstallationState{})
			stubBrewCommandExists(t, false)
			stubExecutionFactories(t,
				func() execution.CommandRunner { return tt.runner },
				func(planning.ResourceKind, execution.CommandRunner, execution.CommandExists) execution.Installer {
					t.Fatal("dotfiles-only apply must not instantiate Homebrew installers")
					return nil
				},
				func(commandRunner execution.CommandRunner) execution.Installer {
					provider := execution.NewLocalDotfilesProvider(commandRunner, execution.DotfilesBaseResolver{
						LookupEnv: func(string) (string, bool) { return base, true },
						HomeDir:   func() (string, error) { return filepath.Dir(base), nil },
					})
					return execution.NewDotfilesInstaller(provider)
				},
			)

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			gotCode := run([]string{"apply", "--yes", "--resource", "dotfile:" + tt.module, "--catalog", catalogPath}, &stdout, &stderr)

			if gotCode != exitFailure {
				t.Fatalf("run() exit code = %d, want %d", gotCode, exitFailure)
			}
			out := stdout.String()
			for _, want := range []string{"[failed]", tt.wantMessage, "- failed: 1"} {
				if !strings.Contains(out, want) {
					t.Fatalf("stdout missing %q; got %q", want, out)
				}
			}
			if strings.Contains(out, "[changed]") {
				t.Fatalf("failed result reported changed: %q", out)
			}
			if tt.name == "missing base" || tt.name == "missing dotlink" || tt.name == "missing module" {
				if len(tt.runner.calls) != 0 {
					t.Fatalf("command runner calls = %d, want none", len(tt.runner.calls))
				}
			} else if len(tt.runner.calls) != 1 {
				t.Fatalf("command runner calls = %d, want one dotlink attempt", len(tt.runner.calls))
			}
			for _, call := range tt.runner.calls {
				request := call.Executable + " " + strings.Join(call.Args, " ")
				for _, forbidden := range []string{"clone", "pull", "submodule", "fetch", "remote", "sparse", "apt"} {
					if strings.Contains(request, forbidden) {
						t.Fatalf("dotfiles path requested forbidden command %q in %#v", forbidden, call)
					}
				}
			}
			if stderr.String() != "" {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
	}
}

func TestRunApplyConfirmedMissingBrewDoesNotInstantiateHomebrewInstaller(t *testing.T) {
	catalogPath := writeFile(t, t.TempDir(), "catalog.toml", `
schema = "dniebles.catalog"
version = 1

[[packages]]
id = "ripgrep"
description = "Fast text search"
[packages.install]
provider = "brew"
package = "ripgrep"

[[profiles]]
id = "dev"
resources = ["package:ripgrep"]
`)
	stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
	stubInstallationState(t, planning.InstallationState{})
	stubConfigState(t, planning.ConfigState{})
	stubDotfilesState(t, planning.InstallationState{})
	stubBrewCommandExists(t, false)
	stubExecutionFactories(t,
		func() execution.CommandRunner {
			t.Fatal("missing brew must not instantiate OS command runner")
			return nil
		},
		func(planning.ResourceKind, execution.CommandRunner, execution.CommandExists) execution.Installer {
			t.Fatal("missing brew must not instantiate Homebrew installer")
			return nil
		},
		func(execution.CommandRunner) execution.Installer {
			t.Fatal("brew-only apply must not instantiate dotfiles installer")
			return nil
		},
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	gotCode := run([]string{"apply", "--profile", "dev", "--catalog", catalogPath, "--yes"}, &stdout, &stderr)

	if gotCode != exitSuccess {
		t.Fatalf("run() exit code = %d, want %d", gotCode, exitSuccess)
	}
	out := stdout.String()
	if !strings.Contains(out, "package:ripgrep [unchanged]") || !strings.Contains(out, "homebrew:bootstrap: Install Homebrew") {
		t.Fatalf("stdout missing skipped install or bootstrap guidance: %q", out)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunApplyCatalogLoadErrors(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	gotCode := run([]string{"apply", "--profile", "dev", "--catalog", filepath.Join(t.TempDir(), "missing.toml")}, &stdout, &stderr)

	if gotCode != exitFailure {
		t.Fatalf("run() exit code = %d, want %d", gotCode, exitFailure)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "error: load catalog ") || !strings.Contains(stderr.String(), "no such file or directory") {
		t.Fatalf("stderr = %q, want load error", stderr.String())
	}
}

func writeDotfilesCatalog(t *testing.T, modules ...string) string {
	t.Helper()
	var catalog strings.Builder
	catalog.WriteString("schema = \"dniebles.catalog\"\nversion = 1\n\n")
	refs := make([]string, 0, len(modules))
	for _, module := range modules {
		catalog.WriteString("[[dotfiles]]\n")
		catalog.WriteString("id = \"")
		catalog.WriteString(module)
		catalog.WriteString("\"\n")
		catalog.WriteString("description = \"")
		catalog.WriteString(module)
		catalog.WriteString(" dotfiles\"\n\n")
		refs = append(refs, "\"dotfile:"+module+"\"")
	}
	catalog.WriteString("[[profiles]]\nid = \"dev\"\nresources = [")
	catalog.WriteString(strings.Join(refs, ", "))
	catalog.WriteString("]\n")
	return writeFile(t, t.TempDir(), "catalog.toml", catalog.String())
}

func makeDotfilesBase(t *testing.T, modules ...string) string {
	t.Helper()
	base := filepath.Join(t.TempDir(), "home", ".dotfiles")
	if err := os.MkdirAll(filepath.Join(base, "bin"), 0o700); err != nil {
		t.Fatalf("create dotfiles bin: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, "bin", "dotlink"), []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatalf("write dotlink: %v", err)
	}
	for _, module := range modules {
		if err := os.MkdirAll(filepath.Join(base, module), 0o700); err != nil {
			t.Fatalf("create module %s: %v", module, err)
		}
	}
	return base
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func stubEnvironmentFacts(t *testing.T, facts planning.EnvironmentFacts) {
	t.Helper()
	original := detectEnvironmentFacts
	detectEnvironmentFacts = func() planning.EnvironmentFacts { return facts }
	t.Cleanup(func() { detectEnvironmentFacts = original })
}

func stubInstallationState(t *testing.T, installation planning.InstallationState) {
	t.Helper()
	original := detectInstallationState
	detectInstallationState = func(planning.Catalog) planning.InstallationState { return installation }
	t.Cleanup(func() { detectInstallationState = original })
}

func stubConfigState(t *testing.T, configState planning.ConfigState) {
	t.Helper()
	original := detectConfigState
	detectConfigState = func(planning.Catalog) planning.ConfigState { return configState }
	t.Cleanup(func() { detectConfigState = original })
}

func stubDotfilesState(t *testing.T, installation planning.InstallationState) {
	t.Helper()
	original := detectDotfilesState
	detectDotfilesState = func(planning.Catalog) planning.InstallationState { return installation }
	t.Cleanup(func() { detectDotfilesState = original })
}

func stubBrewCommandExists(t *testing.T, exists bool) {
	t.Helper()
	original := brewCommandExists
	brewCommandExists = func(name string) bool {
		if name != "brew" {
			t.Fatalf("expected lookup for brew, got %q", name)
		}
		return exists
	}
	t.Cleanup(func() { brewCommandExists = original })
}

func stubExecutionFactories(
	t *testing.T,
	runnerFactory func() execution.CommandRunner,
	installerFactory func(planning.ResourceKind, execution.CommandRunner, execution.CommandExists) execution.Installer,
	dotfilesInstallerFactory func(execution.CommandRunner) execution.Installer,
) {
	t.Helper()
	originalRunnerFactory := newOSCommandRunner
	originalInstallerFactory := newHomebrewInstaller
	originalDotfilesInstallerFactory := newDotfilesInstaller
	newOSCommandRunner = runnerFactory
	newHomebrewInstaller = installerFactory
	newDotfilesInstaller = dotfilesInstallerFactory
	t.Cleanup(func() {
		newOSCommandRunner = originalRunnerFactory
		newHomebrewInstaller = originalInstallerFactory
		newDotfilesInstaller = originalDotfilesInstallerFactory
	})
}

type recordingCommandRunner struct {
	result execution.CommandResult
	calls  []execution.CommandRequest
}

func (r *recordingCommandRunner) RunCommand(_ context.Context, req execution.CommandRequest) execution.CommandResult {
	r.calls = append(r.calls, req)
	r.result.Request = req
	return r.result
}
