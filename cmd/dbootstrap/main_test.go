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
			name:       "missing profile is a stable usage error",
			args:       []string{"plan"},
			wantCode:   exitUsage,
			wantStdout: "",
			wantStderr: "Usage: dbootstrap plan --profile <name> [--catalog <path>]\nerror: --profile is required\n",
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
				"error: command is required\n",
		},
		{
			name: "unknown command",
			args: []string{"apply"},
			wantStderr: "Usage: dbootstrap <command> [options]\n" +
				"\n" +
				"Commands:\n" +
				"  plan    Build a deterministic plan for a profile\n" +
				"error: unknown command \"apply\"\n",
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

func TestRunPlanCatalogLoadErrorsSkipConfigDetection(t *testing.T) {
	stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
	stubInstallationState(t, planning.InstallationState{})
	stubConfigState(t, planning.ConfigState{}) // safe default if somehow reached

	original := detectConfigState
	detectConfigState = func(planning.Catalog) planning.ConfigState {
		t.Fatal("config detection must not run when catalog loading fails")
		return planning.ConfigState{}
	}
	t.Cleanup(func() { detectConfigState = original })

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
