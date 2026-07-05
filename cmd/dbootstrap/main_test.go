package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
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
				"  apply   Run a dry-run execution of the plan (noop only)\n" +
				"error: command is required\n",
		},
		{
			name: "unknown command",
			args: []string{"deploy"},
			wantStderr: "Usage: dbootstrap <command> [options]\n" +
				"\n" +
				"Commands:\n" +
				"  plan    Build a deterministic plan for a profile\n" +
				"  apply   Run a dry-run execution of the plan (noop only)\n" +
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
			name:              "dry run profile renders not implemented execution report",
			args:              []string{"apply", "--profile", "dev", "--catalog", "../../catalog/bootstrap.toml"},
			installationState: planning.InstallationState{},
			configState:       planning.ConfigState{},
			wantCode:          exitSuccess,
			wantStdout: "Execution Report\n" +
				"\n" +
				"Steps:\n" +
				"1. tool:git [not_implemented] noop installer does not perform real installation\n" +
				"2. package:ripgrep [not_implemented] noop installer does not perform real installation\n" +
				"3. runtime:go [not_implemented] noop installer does not perform real installation\n",
			wantStderr: "",
		},
		{
			name:              "dry run resource only renders single step",
			args:              []string{"apply", "--resource", "tool:git", "--catalog", "../../catalog/bootstrap.toml"},
			installationState: planning.InstallationState{},
			configState:       planning.ConfigState{},
			wantCode:          exitSuccess,
			wantStdout: "Execution Report\n" +
				"\n" +
				"Steps:\n" +
				"1. tool:git [not_implemented] noop installer does not perform real installation\n",
			wantStderr: "",
		},
		{
			name:       "missing target is a stable usage error",
			args:       []string{"apply"},
			wantCode:   exitUsage,
			wantStdout: "",
			wantStderr: "Usage: dbootstrap apply [--profile <name>] [--resource <kind:name>] [--catalog <path>]\nerror: --profile or --resource is required\n",
		},
		{
			name:       "malformed resource ref is rejected",
			args:       []string{"apply", "--resource", "git", "--catalog", "../../catalog/bootstrap.toml"},
			wantCode:   exitUsage,
			wantStdout: "",
			wantStderr: "Usage: dbootstrap apply [--profile <name>] [--resource <kind:name>] [--catalog <path>]\nerror: invalid resource ref \"git\": expected kind:name\n",
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
