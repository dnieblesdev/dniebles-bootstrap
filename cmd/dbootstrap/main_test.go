package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	catalogtoml "github.com/dnieblesdev/dniebles-bootstrap/internal/catalog/toml"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/execution"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/version"
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
	catalogPath := writePrimaryCatalog(t)
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
				"2. package:jq [planned] JSON processor\n" +
				"   depends_on: tool:git\n" +
				"   attention: none\n" +
				"3. package:ripgrep [planned] Fast text search\n" +
				"   depends_on: tool:git\n" +
				"   attention: none\n" +
				"4. runtime:go [attention_required] Go toolchain\n" +
				"   depends_on: tool:git\n" +
				"   attention: missing required config \"go.env\"\n" +
				"\n" +
				"Results:\n" +
				"- package:jq: planned\n" +
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
				"2. package:jq [planned] JSON processor\n" +
				"   depends_on: tool:git\n" +
				"   attention: none\n" +
				"3. package:ripgrep [planned] Fast text search\n" +
				"   depends_on: tool:git\n" +
				"   attention: none\n" +
				"4. runtime:go [attention_required] Go toolchain\n" +
				"   depends_on: tool:git\n" +
				"   attention: missing required config \"go.env\"\n" +
				"\n" +
				"Results:\n" +
				"- package:jq: planned\n" +
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
				"2. package:jq [planned] JSON processor\n" +
				"   depends_on: tool:git\n" +
				"   attention: none\n" +
				"3. package:ripgrep [planned] Fast text search\n" +
				"   depends_on: tool:git\n" +
				"   attention: none\n" +
				"4. runtime:go [planned] Go toolchain\n" +
				"   depends_on: tool:git\n" +
				"   attention: none\n" +
				"\n" +
				"Results:\n" +
				"- package:jq: planned\n" +
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
				"3. package:jq [planned] JSON processor\n" +
				"   depends_on: tool:git\n" +
				"   attention: none\n" +
				"4. package:ripgrep [planned] Fast text search\n" +
				"   depends_on: tool:git\n" +
				"   attention: none\n" +
				"5. runtime:go [attention_required] Go toolchain\n" +
				"   depends_on: tool:git\n" +
				"   attention: missing required config \"go.env\"\n" +
				"\n" +
				"Results:\n" +
				"- dotfile:bash: planned\n" +
				"- package:jq: planned\n" +
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

			gotCode := run(replaceCatalogPath(tt.args, catalogPath), &stdout, &stderr)

			if gotCode != tt.wantCode {
				t.Fatalf("run() exit code = %d, want %d", gotCode, tt.wantCode)
			}
			if got, want := stdout.String(), strings.ReplaceAll(tt.wantStdout, "../../catalog/bootstrap.toml", catalogPath); got != want {
				t.Fatalf("stdout = %q, want %q", got, want)
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

func TestRunVersionFlag(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		version    string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "default version reports dev",
			args:       []string{"--version"},
			wantCode:   exitSuccess,
			wantStdout: "dev\n",
			wantStderr: "",
		},
		{
			name:       "injected version is reported",
			args:       []string{"--version"},
			version:    "v1.2.3",
			wantCode:   exitSuccess,
			wantStdout: "v1.2.3\n",
			wantStderr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.version != "" {
				originalVersion := version.Version
				version.Version = tt.version
				t.Cleanup(func() { version.Version = originalVersion })
			}

			var stdout bytes.Buffer
			var stderr bytes.Buffer
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
				"  apply   Execute safely; --yes may run eligible brew-backed installs, eligible Linux APT installs, and selected dotfiles\n" +
				"          APT uses apt-get directly with --yes, or sudo apt-get only with --yes --sudo\n" +
				"  bootstrap  Execute an explicit selection through the safe apply workflow\n" +
				"error: command is required\n",
		},
		{
			name: "unknown command",
			args: []string{"deploy"},
			wantStderr: "Usage: dbootstrap <command> [options]\n" +
				"\n" +
				"Commands:\n" +
				"  plan    Build a deterministic plan for a profile\n" +
				"  apply   Execute safely; --yes may run eligible brew-backed installs, eligible Linux APT installs, and selected dotfiles\n" +
				"          APT uses apt-get directly with --yes, or sudo apt-get only with --yes --sudo\n" +
				"  bootstrap  Execute an explicit selection through the safe apply workflow\n" +
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
	catalogPath := writePrimaryCatalog(t)
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
				"- not supported yet: 4\n" +
				"- failed: 0\n" +
				"\n" +
				"Steps:\n" +
				"1. tool:git [not supported yet] noop installer does not perform real installation\n" +
				"2. package:jq [not supported yet] noop installer does not perform real installation\n" +
				"3. package:ripgrep [not supported yet] noop installer does not perform real installation\n" +
				"4. runtime:go [not supported yet] noop installer does not perform real installation\n" +
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
				"- not supported yet: 4\n" +
				"- failed: 0\n" +
				"\n" +
				"Steps:\n" +
				"1. tool:git [not supported yet] noop installer does not perform real installation\n" +
				"2. package:jq [not supported yet] noop installer does not perform real installation\n" +
				"3. package:ripgrep [not supported yet] noop installer does not perform real installation\n" +
				"4. runtime:go [not supported yet] noop installer does not perform real installation\n" +
				"\n" +
				"Manual Actions:\n" +
				homebrewManualActionOutput,
			wantStderr: "",
		},
		{
			name:              "yes flag reports unknown package presence when brew is missing",
			args:              []string{"apply", "--profile", "dev", "--catalog", "../../catalog/bootstrap.toml", "--yes"},
			installationState: planning.InstallationState{},
			configState:       planning.ConfigState{},
			wantCode:          exitFailure,
			wantStdout: "Execution Report\n" +
				"Mode: confirmed\n" +
				"Confirmed mode: brew-backed tool/package steps, eligible Linux APT-backed tool/package steps, and selected dotfile resources may have changed this machine; unsupported, non-provider-backed, and unselected steps remain non-mutating or not supported yet.\n" +
				"\n" +
				"Summary:\n" +
				"- changed: 0\n" +
				"- unchanged: 1\n" +
				"- not supported yet: 1\n" +
				"- failed: 2\n" +
				"\n" +
				"Steps:\n" +
				"1. tool:git [unchanged] skipped because Homebrew must be installed manually before brew-backed resources can be applied\n" +
				"2. package:jq [failed] Homebrew formula presence could not be determined; no mutation attempted\n" +
				"3. package:ripgrep [failed] Homebrew formula presence could not be determined; no mutation attempted\n" +
				"4. runtime:go [not supported yet] noop installer does not perform real installation\n" +
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
			wantStderr: "Usage: dbootstrap apply [--profile <name>] [--resource <kind:name>] [--catalog <path>] [--dry-run] [--yes [--sudo]]\nerror: --profile or --resource is required\n",
		},
		{
			name:       "dry run and yes cannot be combined",
			args:       []string{"apply", "--profile", "dev", "--catalog", "../../catalog/bootstrap.toml", "--dry-run", "--yes"},
			wantCode:   exitUsage,
			wantStdout: "",
			wantStderr: "Usage: dbootstrap apply [--profile <name>] [--resource <kind:name>] [--catalog <path>] [--dry-run] [--yes [--sudo]]\nerror: --dry-run and --yes cannot be combined\n",
		},
		{
			name:       "malformed resource ref is rejected",
			args:       []string{"apply", "--resource", "git", "--catalog", "../../catalog/bootstrap.toml"},
			wantCode:   exitUsage,
			wantStdout: "",
			wantStderr: "Usage: dbootstrap apply [--profile <name>] [--resource <kind:name>] [--catalog <path>] [--dry-run] [--yes [--sudo]]\nerror: invalid resource ref \"git\": expected kind:name\n",
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

			gotCode := run(replaceCatalogPath(tt.args, catalogPath), &stdout, &stderr)

			if gotCode != tt.wantCode {
				t.Fatalf("run() exit code = %d, want %d", gotCode, tt.wantCode)
			}
			if got, want := stdout.String(), strings.ReplaceAll(tt.wantStdout, "../../catalog/bootstrap.toml", catalogPath); got != want {
				t.Fatalf("stdout = %q, want %q", got, want)
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Fatalf("stderr = %q, want %q", got, tt.wantStderr)
			}
		})
	}
}

func TestResolveDefaultCatalogPath(t *testing.T) {
	tests := []struct {
		name           string
		xdgDataHome    string
		home           string
		homeErr        error
		homebrewPrefix string
		existing       map[string]bool
		want           string
	}{
		{
			name:        "XDG_DATA_HOME takes precedence when existing",
			xdgDataHome: "/custom/data",
			home:        "/home/user",
			existing: map[string]bool{
				filepath.Join("/custom/data", "dbootstrap", "catalog", "bootstrap.toml"): true,
			},
			want: filepath.Join("/custom/data", "dbootstrap", "catalog", "bootstrap.toml"),
		},
		{
			name:           "HOME wins when XDG unset and Homebrew exists",
			xdgDataHome:    "",
			home:           "/home/user",
			homebrewPrefix: "/opt/homebrew",
			existing: map[string]bool{
				filepath.Join("/home/user", ".local", "share", "dbootstrap", "catalog", "bootstrap.toml"): true,
				filepath.Join("/opt/homebrew", "share", "dbootstrap", "catalog", "bootstrap.toml"):        true,
			},
			want: filepath.Join("/home/user", ".local", "share", "dbootstrap", "catalog", "bootstrap.toml"),
		},
		{
			name:           "Homebrew wins when higher candidates missing",
			xdgDataHome:    "",
			home:           "/home/user",
			homebrewPrefix: "/opt/homebrew",
			existing: map[string]bool{
				filepath.Join("/opt/homebrew", "share", "dbootstrap", "catalog", "bootstrap.toml"): true,
			},
			want: filepath.Join("/opt/homebrew", "share", "dbootstrap", "catalog", "bootstrap.toml"),
		},
		{
			name:           "higher priority wins over Homebrew",
			xdgDataHome:    "/custom/data",
			home:           "/home/user",
			homebrewPrefix: "/opt/homebrew",
			existing: map[string]bool{
				filepath.Join("/custom/data", "dbootstrap", "catalog", "bootstrap.toml"):           true,
				filepath.Join("/opt/homebrew", "share", "dbootstrap", "catalog", "bootstrap.toml"): true,
			},
			want: filepath.Join("/custom/data", "dbootstrap", "catalog", "bootstrap.toml"),
		},
		{
			name:        "XDG_DATA_HOME empty falls back to HOME",
			xdgDataHome: "",
			home:        "/home/user",
			existing: map[string]bool{
				filepath.Join("/home/user", ".local", "share", "dbootstrap", "catalog", "bootstrap.toml"): true,
			},
			want: filepath.Join("/home/user", ".local", "share", "dbootstrap", "catalog", "bootstrap.toml"),
		},
		{
			name:    "home resolution error returns empty",
			homeErr: errors.New("no home"),
			want:    "",
		},
		{
			name:           "absent HOMEBREW_PREFIX omits Homebrew candidate",
			xdgDataHome:    "/custom/data",
			home:           "/home/user",
			homebrewPrefix: "",
			existing: map[string]bool{
				filepath.Join("/custom/data", "dbootstrap", "catalog", "bootstrap.toml"): true,
			},
			want: filepath.Join("/custom/data", "dbootstrap", "catalog", "bootstrap.toml"),
		},
		{
			name:           "no existing candidates returns highest priority",
			xdgDataHome:    "/custom/data",
			home:           "/home/user",
			homebrewPrefix: "/opt/homebrew",
			existing:       map[string]bool{},
			want:           filepath.Join("/custom/data", "dbootstrap", "catalog", "bootstrap.toml"),
		},
		{
			name:           "no existing candidates without XDG returns home local",
			xdgDataHome:    "",
			home:           "/home/user",
			homebrewPrefix: "/opt/homebrew",
			existing:       map[string]bool{},
			want:           filepath.Join("/home/user", ".local", "share", "dbootstrap", "catalog", "bootstrap.toml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := catalogPathResolver{
				LookupEnv: func(key string) (string, bool) {
					switch key {
					case "XDG_DATA_HOME":
						return tt.xdgDataHome, tt.xdgDataHome != ""
					case "HOMEBREW_PREFIX":
						return tt.homebrewPrefix, tt.homebrewPrefix != ""
					default:
						t.Fatalf("unexpected env lookup %q", key)
						return "", false
					}
				},
				HomeDir: func() (string, error) {
					return tt.home, tt.homeErr
				},
				PathExists: func(path string) bool {
					return tt.existing[path]
				},
			}

			got := resolver.Resolve()
			if got != tt.want {
				t.Fatalf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunPlanDefaultCatalogFromXDGDataHome(t *testing.T) {
	installedCatalog := writePrimaryCatalog(t)
	original := defaultCatalogPath
	defaultCatalogPath = func() string { return installedCatalog }
	t.Cleanup(func() { defaultCatalogPath = original })

	var stdout, stderr bytes.Buffer
	stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
	stubInstallationState(t, planning.InstallationState{})
	stubConfigState(t, planning.ConfigState{})
	stubDotfilesState(t, planning.InstallationState{})

	gotCode := run([]string{"plan", "--profile", "dev"}, &stdout, &stderr)
	if gotCode != exitSuccess {
		t.Fatalf("run() exit code = %d, want %d; stderr=%q", gotCode, exitSuccess, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Plan profile: dev") {
		t.Fatalf("stdout missing plan header: %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Catalog: "+installedCatalog) {
		t.Fatalf("stdout did not use installed catalog path: %q", stdout.String())
	}
}

func TestRunPlanDefaultCatalogFromOutsideRepository(t *testing.T) {
	installedCatalog := writePrimaryCatalog(t)
	original := defaultCatalogPath
	defaultCatalogPath = func() string { return installedCatalog }
	t.Cleanup(func() { defaultCatalogPath = original })

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	workDir := t.TempDir()
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	var stdout, stderr bytes.Buffer
	stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
	stubInstallationState(t, planning.InstallationState{})
	stubConfigState(t, planning.ConfigState{})
	stubDotfilesState(t, planning.InstallationState{})

	gotCode := run([]string{"plan", "--profile", "dev"}, &stdout, &stderr)
	if gotCode != exitSuccess {
		t.Fatalf("run() exit code = %d, want %d; stderr=%q", gotCode, exitSuccess, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Catalog: "+installedCatalog) {
		t.Fatalf("stdout did not use installed catalog path outside repo: %q", stdout.String())
	}
}

func TestRunPlanDefaultCatalogSmokeIsDerived(t *testing.T) {
	assertDefaultCatalogPlanSmoke(t)
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
				"Confirmed mode: brew-backed tool/package steps, eligible Linux APT-backed tool/package steps, and selected dotfile resources may have changed this machine",
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
			brewLookups := 0
			originalBrew := brewCommandExists
			brewCommandExists = func(name string) bool {
				brewLookups++
				if name != "brew" {
					t.Fatalf("expected lookup for brew, got %q", name)
				}
				return false
			}
			t.Cleanup(func() { brewCommandExists = originalBrew })
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
			if args[0] == "apply" && brewLookups != 1 {
				t.Fatalf("Homebrew lookups = %d, want only bootstrap availability check", brewLookups)
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

func TestRunBootstrapDefaultAndDryRunDoNotProbeBrew(t *testing.T) {
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

	for _, flags := range [][]string{nil, {"--dry-run"}} {
		t.Run(strings.Join(append([]string{"default"}, flags...), " "), func(t *testing.T) {
			stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
			stubInstallationState(t, planning.InstallationState{})
			stubConfigState(t, planning.ConfigState{})
			stubDotfilesState(t, planning.InstallationState{})
			originalBrew := brewCommandExists
			brewCommandExists = func(name string) bool {
				t.Fatalf("bootstrap default and dry-run modes must not probe %q", name)
				return false
			}
			t.Cleanup(func() { brewCommandExists = originalBrew })
			stubExecutionFactories(t,
				func() execution.CommandRunner {
					t.Fatal("bootstrap default and dry-run modes must not instantiate OS command runners")
					return nil
				},
				func(planning.ResourceKind, execution.CommandRunner, execution.CommandExists) execution.Installer {
					t.Fatal("bootstrap default and dry-run modes must not instantiate Homebrew installers")
					return nil
				},
				func(execution.CommandRunner) execution.Installer {
					t.Fatal("bootstrap default and dry-run modes must not instantiate dotfiles installers")
					return nil
				},
			)

			var stdout, stderr bytes.Buffer
			args := append([]string{"bootstrap", "--profile", "dev", "--catalog", catalogPath}, flags...)
			if code := run(args, &stdout, &stderr); code != exitSuccess {
				t.Fatalf("exit code = %d, want %d; stdout=%q stderr=%q", code, exitSuccess, stdout.String(), stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
	}
}

func TestRunApplyAndBootstrapSkipDetectedCommandPresence(t *testing.T) {
	catalogPath := writeFile(t, t.TempDir(), "catalog.toml", `schema = "dniebles.catalog"
version = 1

[[tools]]
id = "editor"
description = "Editor"
[tools.install]
provider = "brew"
package = "vim"
[tools.presence]
kind = "command_exists"
name = "vim"

[[profiles]]
id = "dev"
resources = ["tool:editor"]
`)
	for _, command := range []string{"apply", "bootstrap"} {
		t.Run(command, func(t *testing.T) {
			stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
			stubInstallationState(t, planning.InstallationState{PresentResources: map[planning.ResourceRef]bool{{Kind: planning.ResourceKindTool, Name: "editor"}: true}})
			stubConfigState(t, planning.ConfigState{})
			stubDotfilesState(t, planning.InstallationState{})
			stubBrewCommandExists(t, true)
			runner := &recordingCommandRunner{result: execution.CommandResult{Status: execution.CommandStatusSucceeded}}
			stubExecutionFactories(t, func() execution.CommandRunner { return runner }, newHomebrewInstaller, newDotfilesInstaller)

			var stdout, stderr bytes.Buffer
			if code := run([]string{command, "--profile", "dev", "--catalog", catalogPath, "--yes"}, &stdout, &stderr); code != exitSuccess {
				t.Fatalf("exit code = %d, want %d; stdout=%q stderr=%q", code, exitSuccess, stdout.String(), stderr.String())
			}
			if len(runner.calls) != 0 {
				t.Fatalf("command calls = %#v, want none", runner.calls)
			}
			if !strings.Contains(stdout.String(), "tool:editor [unchanged] already installed; no mutation attempted") {
				t.Fatalf("stdout missing idempotent result: %q", stdout.String())
			}
		})
	}
}

func TestRunApplySafeModesDoNotReportConfirmedIdempotencySkip(t *testing.T) {
	catalogPath := writeFile(t, t.TempDir(), "catalog.toml", `schema = "dniebles.catalog"
version = 1

[[tools]]
id = "editor"
description = "Editor"
[tools.presence]
kind = "command_exists"
name = "vim"

[[profiles]]
id = "dev"
resources = ["tool:editor"]
`)
	for _, args := range [][]string{{"apply", "--profile", "dev", "--catalog", catalogPath}, {"apply", "--profile", "dev", "--catalog", catalogPath, "--dry-run"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
			stubInstallationState(t, planning.InstallationState{PresentResources: map[planning.ResourceRef]bool{{Kind: planning.ResourceKindTool, Name: "editor"}: true}})
			stubConfigState(t, planning.ConfigState{})
			stubDotfilesState(t, planning.InstallationState{})
			var stdout, stderr bytes.Buffer
			if code := run(args, &stdout, &stderr); code != exitSuccess {
				t.Fatalf("exit code = %d, want %d", code, exitSuccess)
			}
			if strings.Contains(stdout.String(), "already installed; no mutation attempted") || !strings.Contains(stdout.String(), "tool:editor [not supported yet]") {
				t.Fatalf("safe-mode output = %q, want existing non-mutating report", stdout.String())
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
	originalExists := aptCommandExists
	aptCommandExists = func(name string) bool { return name == "dpkg-query" || name == "apt-get" }
	t.Cleanup(func() { aptCommandExists = originalExists })
	runner := &sequenceCommandRunner{results: []execution.CommandResult{
		{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "dpkg-query: no packages found matching ripgrep", Err: errors.New("exit 1")},
		{Status: execution.CommandStatusSucceeded, ExitCode: 0},
		{Status: execution.CommandStatusSucceeded, ExitCode: 0},
	}}
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
	if len(runner.calls) != 3 {
		t.Fatalf("command runner calls = %d, want 3", len(runner.calls))
	}
	if got := runner.calls[0]; got.Executable != "dpkg-query" || strings.Join(got.Args, " ") != "--show --showformat=${Status} ripgrep" {
		t.Fatalf("CommandRequest = %#v, want dpkg-query --show --showformat=${Status} ripgrep", got)
	}
	if got := runner.calls[1]; got.Executable != "apt-get" || strings.Join(got.Args, " ") != "install -y -- ripgrep" {
		t.Fatalf("CommandRequest = %#v, want apt-get install -y -- ripgrep", got)
	}
	if got := runner.calls[2]; got.Executable != "brew" || strings.Join(got.Args, " ") != "install fd" {
		t.Fatalf("CommandRequest = %#v, want brew install fd", got)
	}
	out := stdout.String()
	for _, want := range []string{
		"Mode: confirmed",
		"tool:fd [changed] installed fd with Homebrew",
		"package:ripgrep [changed] installed ripgrep with APT",
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

func TestRunApplyLikeConfirmedMixedBrewAptPreservesBrewPresence(t *testing.T) {
	catalogPath := writeFile(t, t.TempDir(), "catalog.toml", `
schema = "dniebles.catalog"
version = 1

[[packages]]
id = "jq"
description = "JSON processor"
[packages.install]
provider = "brew"
package = "jq"

[[packages]]
id = "ripgrep"
description = "Fast text search"
[packages.install]
provider = "apt"
package = "ripgrep"

[[profiles]]
id = "dev"
resources = ["package:jq", "package:ripgrep"]
`)
	for _, command := range []string{"apply", "bootstrap"} {
		t.Run(command, func(t *testing.T) {
			stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
			stubInstallationState(t, planning.InstallationState{})
			stubConfigState(t, planning.ConfigState{})
			stubDotfilesState(t, planning.InstallationState{})
			stubBrewCommandExists(t, true)
			originalExists := aptCommandExists
			aptCommandExists = func(name string) bool { return name == "dpkg-query" || name == "apt-get" }
			t.Cleanup(func() { aptCommandExists = originalExists })
			runner := &sequenceCommandRunner{results: []execution.CommandResult{
				{Status: execution.CommandStatusSucceeded, ExitCode: 0},
				{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "dpkg-query: no packages found matching ripgrep", Err: errors.New("exit 1")},
				{Status: execution.CommandStatusSucceeded, ExitCode: 0},
			}}
			stubExecutionFactories(t,
				func() execution.CommandRunner { return runner },
				func(kind planning.ResourceKind, commandRunner execution.CommandRunner, exists execution.CommandExists) execution.Installer {
					return execution.NewHomebrewInstaller(kind, commandRunner, exists)
				},
				func(execution.CommandRunner) execution.Installer {
					t.Fatal("mixed brew+apt apply must not instantiate dotfiles installer")
					return nil
				},
			)

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			gotCode := run([]string{command, "--profile", "dev", "--catalog", catalogPath, "--yes"}, &stdout, &stderr)

			if gotCode != exitSuccess {
				t.Fatalf("run() exit code = %d, want %d", gotCode, exitSuccess)
			}
			if len(runner.calls) != 3 {
				t.Fatalf("command runner calls = %d, want 3", len(runner.calls))
			}
			if got := runner.calls[0]; got.Executable != "brew" || strings.Join(got.Args, " ") != "list --formula jq" {
				t.Fatalf("CommandRequest = %#v, want brew list --formula jq", got)
			}
			if got := runner.calls[1]; got.Executable != "dpkg-query" || strings.Join(got.Args, " ") != "--show --showformat=${Status} ripgrep" {
				t.Fatalf("CommandRequest = %#v, want dpkg-query --show --showformat=${Status} ripgrep", got)
			}
			if got := runner.calls[2]; got.Executable != "apt-get" || strings.Join(got.Args, " ") != "install -y -- ripgrep" {
				t.Fatalf("CommandRequest = %#v, want apt-get install -y -- ripgrep", got)
			}
			out := stdout.String()
			for _, want := range []string{
				"Mode: confirmed",
				"package:jq [unchanged] already installed; no mutation attempted",
				"package:ripgrep [changed] installed ripgrep with APT",
				"Manual Actions:\n- none\n",
			} {
				if !strings.Contains(out, want) {
					t.Fatalf("stdout missing %q; got %q", want, out)
				}
			}
			if stderr.String() != "" {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
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
	runner := &recordingCommandRunner{result: execution.CommandResult{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: `{"schema_version":1,"modules":["bash"],"status":"success","entries":[{"module":"bash","source":"bashrc","target":"/home/ada/.bashrc","outcome":"changed"}],"failure":null,"rollback":{"attempted":false,"completed":false,"removed":[]}}`}}
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
	if call.Executable != filepath.Join(base, "bin", "dotlink") || strings.Join(call.Args, " ") != "link --report=json bash" || call.Dir != base {
		t.Fatalf("CommandRequest = %#v, want dotlink for bash only", call)
	}
	out := stdout.String()
	for _, want := range []string{
		"Mode: confirmed",
		"Confirmed mode: brew-backed tool/package steps, eligible Linux APT-backed tool/package steps, and selected dotfile resources may have changed this machine",
		"dotfile:bash [changed] installed dotfile module bash",
		"dotfiles base: canonical base=" + base,
		"source=env",
		"modules=bash",
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
		name            string
		module          string
		baseSetup       func(t *testing.T) string
		runner          *recordingCommandRunner
		wantMessage     string
		wantDetails     []string
		wantDetailOrder []string
	}{
		{
			name:        "missing base",
			module:      "bash",
			baseSetup:   func(t *testing.T) string { return filepath.Join(t.TempDir(), "missing-dotfiles") },
			runner:      &recordingCommandRunner{result: execution.CommandResult{Status: execution.CommandStatusSucceeded, ExitCode: 0}},
			wantMessage: "dotfile module bash failed",
			wantDetails: []string{
				"dotfiles base: source=env",
				"attempted candidate=", // The full candidate is asserted from the temp path below.
				"modules=bash",
				"cause=resolve dotfiles base",
			},
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
			wantMessage: "dotfile module bash failed",
		},
		{
			name:        "missing module",
			module:      "zsh",
			baseSetup:   func(t *testing.T) string { return makeDotfilesBase(t, "bash") },
			runner:      &recordingCommandRunner{result: execution.CommandResult{Status: execution.CommandStatusSucceeded, ExitCode: 0}},
			wantMessage: "dotfile module zsh failed",
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
			name:      "failed report retains detail before nonzero exit",
			module:    "bash",
			baseSetup: func(t *testing.T) string { return makeDotfilesBase(t, "bash") },
			runner: &recordingCommandRunner{result: execution.CommandResult{
				Status: execution.CommandStatusFailed,
				Stdout: string(readDotlinkReportFixture(t, "failed.json")),
			}},
			wantMessage: "dotfile module bash failed",
			wantDetails: []string{
				"link: failed source=bashrc target=/home/ada/.bashrc",
				"aggregate failure: module=bash cause=link_failed: target exists",
			},
		},
		{
			name:      "rolled back report renders ordered detail before nonzero exit",
			module:    "bash",
			baseSetup: func(t *testing.T) string { return makeDotfilesBase(t, "bash") },
			runner: &recordingCommandRunner{result: execution.CommandResult{
				Status: execution.CommandStatusFailed,
				Stdout: string(readDotlinkReportFixture(t, "rolled-back.json")),
			}},
			wantMessage: "dotfile module bash failed",
			wantDetails: []string{
				"link: rolled_back source=bashrc target=/home/ada/.bashrc",
				"cause: rollback: link reverted",
				"aggregate failure: module=bash cause=link_failed: target exists",
				"rollback: attempted=true completed=true",
				"rollback removed: /home/ada/.bashrc",
			},
			wantDetailOrder: []string{
				"link: rolled_back source=bashrc target=/home/ada/.bashrc",
				"aggregate failure: module=bash cause=link_failed: target exists",
				"rollback: attempted=true completed=true",
				"rollback removed: /home/ada/.bashrc",
			},
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
			for _, want := range append([]string{"[failed]", tt.wantMessage, "- failed: 1"}, tt.wantDetails...) {
				if !strings.Contains(out, want) {
					t.Fatalf("stdout missing %q; got %q", want, out)
				}
			}
			if strings.Contains(out, "[changed]") {
				t.Fatalf("failed result reported changed: %q", out)
			}
			if tt.name == "missing base" {
				if !strings.Contains(out, "attempted candidate="+base) {
					t.Fatalf("stdout missing attempted candidate %q: %q", base, out)
				}
				if strings.Contains(out, "canonical base="+base) {
					t.Fatalf("stdout mislabeled unresolved candidate as canonical: %q", out)
				}
			}
			previousIndex := -1
			for _, detail := range tt.wantDetailOrder {
				index := strings.Index(out, detail)
				if index == -1 || index <= previousIndex {
					t.Fatalf("detail %q was not rendered in order: %q", detail, out)
				}
				previousIndex = index
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

func TestRunApplyConfirmedMissingDotlinkRendersPrerequisiteDiagnostics(t *testing.T) {
	base := makeDotfilesBase(t, "bash")
	missingRunner := filepath.Join(base, "bin", "dotlink")
	if err := os.Remove(missingRunner); err != nil {
		t.Fatalf("remove dotlink: %v", err)
	}
	catalogPath := writeDotfilesCatalog(t, "bash")
	stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
	stubInstallationState(t, planning.InstallationState{})
	stubConfigState(t, planning.ConfigState{})
	stubDotfilesState(t, planning.InstallationState{})
	stubBrewCommandExists(t, false)
	runner := &recordingCommandRunner{}
	stubExecutionFactories(t,
		func() execution.CommandRunner { return runner },
		func(planning.ResourceKind, execution.CommandRunner, execution.CommandExists) execution.Installer {
			return nil
		},
		func(commandRunner execution.CommandRunner) execution.Installer {
			return execution.NewDotfilesInstaller(execution.NewLocalDotfilesProvider(commandRunner, execution.DotfilesBaseResolver{
				LookupEnv: func(string) (string, bool) { return base, true },
				HomeDir:   func() (string, error) { return filepath.Dir(base), nil },
			}))
		},
	)

	var stdout, stderr bytes.Buffer
	if got := run([]string{"apply", "--yes", "--resource", "dotfile:bash", "--catalog", catalogPath}, &stdout, &stderr); got != exitFailure {
		t.Fatalf("exit code = %d, want %d", got, exitFailure)
	}
	for _, want := range []string{"dotfile:bash [failed]", "modules=bash", "phase: prerequisite validation", "attempted runner candidate: " + missingRunner, "cause: path does not exist"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout missing %q: %q", want, stdout.String())
		}
	}
	if len(runner.calls) != 0 || stderr.Len() != 0 {
		t.Fatalf("runner calls=%d stderr=%q, want no command and no stderr", len(runner.calls), stderr.String())
	}
}

func TestRunApplyConfirmedMissingBrewReportsUnknownWithoutInstantiatingHomebrewInstaller(t *testing.T) {
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

	if gotCode != exitFailure {
		t.Fatalf("run() exit code = %d, want %d", gotCode, exitFailure)
	}
	out := stdout.String()
	if !strings.Contains(out, "package:ripgrep [failed] Homebrew formula presence could not be determined; no mutation attempted") || !strings.Contains(out, "homebrew:bootstrap: Install Homebrew") {
		t.Fatalf("stdout missing unknown-presence result or bootstrap guidance: %q", out)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunBootstrapConfirmedMissingBrewReportsUnknownWithoutInstantiatingHomebrewInstaller(t *testing.T) {
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
			t.Fatal("brew-only bootstrap must not instantiate dotfiles installer")
			return nil
		},
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	gotCode := run([]string{"bootstrap", "--profile", "dev", "--catalog", catalogPath, "--yes"}, &stdout, &stderr)

	if gotCode != exitFailure {
		t.Fatalf("run() exit code = %d, want %d", gotCode, exitFailure)
	}
	out := stdout.String()
	if !strings.Contains(out, "package:ripgrep [failed] Homebrew formula presence could not be determined; no mutation attempted") || !strings.Contains(out, "homebrew:bootstrap: Install Homebrew") {
		t.Fatalf("stdout missing unknown-presence result or bootstrap guidance: %q", out)
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

func TestParseApplyFlagsSudoRequiresConfirmedMode(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		ok   bool
		mode applyMode
	}{
		{"sudo with yes", []string{"--resource", "package:ripgrep", "--yes", "--sudo"}, true, applyModeConfirmedSudo},
		{"sudo without yes", []string{"--resource", "package:ripgrep", "--sudo"}, false, ""},
		{"sudo with dry run", []string{"--resource", "package:ripgrep", "--dry-run", "--sudo"}, false, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			_, _, mode, ok := parseApplyFlags(tt.args, &stderr)
			if ok != tt.ok || mode != tt.mode {
				t.Fatalf("ok=%t mode=%q, want ok=%t mode=%q", ok, mode, tt.ok, tt.mode)
			}
		})
	}
}

func TestRunBootstrapAptFixtureContracts(t *testing.T) {
	for _, tt := range []struct {
		name       string
		args       []string
		facts      planning.EnvironmentFacts
		available  map[string]bool
		results    []execution.CommandResult
		wantCode   int
		wantCalls  []execution.CommandRequest
		wantOutput string
	}{
		{
			name:      "apt present and brew absent",
			args:      []string{"--yes"},
			facts:     planning.EnvironmentFacts{OS: "linux"},
			available: map[string]bool{"dpkg-query": true, "apt-get": true},
			results: []execution.CommandResult{
				{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "dpkg-query: no packages found matching ripgrep", Err: errors.New("exit 1")},
				{Status: execution.CommandStatusSucceeded},
			},
			wantCode: exitSuccess,
			wantCalls: []execution.CommandRequest{
				{Executable: "dpkg-query", Args: []string{"--show", "--showformat=${Status}", "ripgrep"}, Timeout: 30 * time.Second},
				{Executable: "apt-get", Args: []string{"install", "-y", "--", "ripgrep"}, Timeout: 10 * time.Minute},
			},
			wantOutput: "package:ripgrep [changed] installed ripgrep with APT",
		},
		{
			name:      "explicit sudo linux",
			args:      []string{"--yes", "--sudo"},
			facts:     planning.EnvironmentFacts{OS: "linux"},
			available: map[string]bool{"dpkg-query": true, "apt-get": true, "sudo": true},
			results: []execution.CommandResult{
				{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "dpkg-query: no packages found matching ripgrep", Err: errors.New("exit 1")},
				{Status: execution.CommandStatusSucceeded},
			},
			wantCode: exitSuccess,
			wantCalls: []execution.CommandRequest{
				{Executable: "dpkg-query", Args: []string{"--show", "--showformat=${Status}", "ripgrep"}, Timeout: 30 * time.Second},
				{Executable: "sudo", Args: []string{"apt-get", "install", "-y", "--", "ripgrep"}, Timeout: 10 * time.Minute},
			},
			wantOutput: "package:ripgrep [changed] installed ripgrep with APT",
		},
		{
			name:      "missing apt-get fails without command",
			args:      []string{"--yes"},
			facts:     planning.EnvironmentFacts{OS: "linux"},
			available: map[string]bool{"dpkg-query": true},
			results: []execution.CommandResult{
				{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "dpkg-query: no packages found matching ripgrep", Err: errors.New("exit 1")},
			},
			wantCode: exitFailure,
			wantCalls: []execution.CommandRequest{
				{Executable: "dpkg-query", Args: []string{"--show", "--showformat=${Status}", "ripgrep"}, Timeout: 30 * time.Second},
			},
			wantOutput: "package:ripgrep [failed] apt-get executable is not available on PATH",
		},
		{
			name:      "missing sudo fails without command",
			args:      []string{"--yes", "--sudo"},
			facts:     planning.EnvironmentFacts{OS: "linux"},
			available: map[string]bool{"dpkg-query": true, "apt-get": true},
			results: []execution.CommandResult{
				{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "dpkg-query: no packages found matching ripgrep", Err: errors.New("exit 1")},
			},
			wantCode: exitFailure,
			wantCalls: []execution.CommandRequest{
				{Executable: "dpkg-query", Args: []string{"--show", "--showformat=${Status}", "ripgrep"}, Timeout: 30 * time.Second},
			},
			wantOutput: "package:ripgrep [failed] sudo executable is not available on PATH",
		},
		{
			name:      "command failure renders and exits non-zero",
			args:      []string{"--yes"},
			facts:     planning.EnvironmentFacts{OS: "linux"},
			available: map[string]bool{"dpkg-query": true, "apt-get": true},
			results: []execution.CommandResult{
				{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "dpkg-query: no packages found matching ripgrep", Err: errors.New("exit 1")},
				{Status: execution.CommandStatusFailed},
			},
			wantCode: exitFailure,
			wantCalls: []execution.CommandRequest{
				{Executable: "dpkg-query", Args: []string{"--show", "--showformat=${Status}", "ripgrep"}, Timeout: 30 * time.Second},
				{Executable: "apt-get", Args: []string{"install", "-y", "--", "ripgrep"}, Timeout: 10 * time.Minute},
			},
			wantOutput: "package:ripgrep [failed] apt install ripgrep failed with status failed",
		},
		{
			name:      "timeout renders and exits non-zero",
			args:      []string{"--yes"},
			facts:     planning.EnvironmentFacts{OS: "linux"},
			available: map[string]bool{"dpkg-query": true, "apt-get": true},
			results: []execution.CommandResult{
				{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "dpkg-query: no packages found matching ripgrep", Err: errors.New("exit 1")},
				{Status: execution.CommandStatusTimedOut},
			},
			wantCode: exitFailure,
			wantCalls: []execution.CommandRequest{
				{Executable: "dpkg-query", Args: []string{"--show", "--showformat=${Status}", "ripgrep"}, Timeout: 30 * time.Second},
				{Executable: "apt-get", Args: []string{"install", "-y", "--", "ripgrep"}, Timeout: 10 * time.Minute},
			},
			wantOutput: "package:ripgrep [failed] apt install ripgrep failed with status timed_out",
		},
		{
			name:       "non linux fails without probe",
			args:       []string{"--yes"},
			facts:      planning.EnvironmentFacts{OS: "darwin"},
			available:  nil,
			results:    nil,
			wantCode:   exitFailure,
			wantCalls:  nil,
			wantOutput: "package:ripgrep [failed] apt execution unsupported_os on darwin (command status not_run)",
		},
		{
			name:       "default does not probe",
			facts:      planning.EnvironmentFacts{OS: "linux"},
			available:  nil,
			results:    nil,
			wantCode:   exitSuccess,
			wantCalls:  nil,
			wantOutput: "package:ripgrep [not supported yet] noop installer does not perform real installation",
		},
		{
			name:       "dry run does not probe",
			args:       []string{"--dry-run"},
			facts:      planning.EnvironmentFacts{OS: "linux"},
			available:  nil,
			results:    nil,
			wantCode:   exitSuccess,
			wantCalls:  nil,
			wantOutput: "package:ripgrep [not supported yet] noop installer does not perform real installation",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stubEnvironmentFacts(t, tt.facts)
			stubInstallationState(t, planning.InstallationState{})
			stubConfigState(t, planning.ConfigState{})
			stubDotfilesState(t, planning.InstallationState{})
			stubBrewCommandExists(t, false)
			originalExists := aptCommandExists
			aptCommandExists = func(name string) bool {
				if tt.available == nil {
					t.Fatalf("APT must not be probed")
				}
				return tt.available[name]
			}
			t.Cleanup(func() { aptCommandExists = originalExists })

			var runner execution.CommandRunner
			if tt.results != nil {
				runner = &sequenceCommandRunner{results: tt.results}
			} else {
				runner = &recordingCommandRunner{result: execution.CommandResult{Status: execution.CommandStatusSucceeded}}
			}
			stubExecutionFactories(t, func() execution.CommandRunner { return runner }, newHomebrewInstaller, newDotfilesInstaller)

			args := append([]string{"bootstrap", "--profile", "apt-fixture", "--catalog", writeAptCatalog(t)}, tt.args...)
			var stdout, stderr bytes.Buffer
			if code := run(args, &stdout, &stderr); code != tt.wantCode {
				t.Fatalf("exit code = %d, want %d; stdout=%q stderr=%q", code, tt.wantCode, stdout.String(), stderr.String())
			}

			var calls []execution.CommandRequest
			switch r := runner.(type) {
			case *sequenceCommandRunner:
				calls = r.calls
			case *recordingCommandRunner:
				calls = r.calls
			}
			if !reflect.DeepEqual(calls, tt.wantCalls) {
				t.Fatalf("command calls = %#v, want %#v", calls, tt.wantCalls)
			}
			if !strings.Contains(stdout.String(), tt.wantOutput) {
				t.Fatalf("stdout = %q, want it to contain %q", stdout.String(), tt.wantOutput)
			}
		})
	}
}

func writeAptCatalog(t *testing.T) string {
	t.Helper()
	return writeFile(t, t.TempDir(), "apt-provider.toml", `schema = "dniebles.catalog"
version = 1

[[packages]]
id = "ripgrep"
description = "Opt-in APT fixture"
[packages.install]
provider = "apt"
package = "ripgrep"

[[profiles]]
id = "apt-fixture"
resources = ["package:ripgrep"]
`)
}

func writePrimaryCatalog(t *testing.T) string {
	t.Helper()
	return writeFile(t, t.TempDir(), "catalog.toml", `
schema = "dniebles.catalog"
version = 1

[[tools]]
id = "git"
description = "Version control"
os = ["linux", "darwin"]
[tools.install]
provider = "brew"
package = "git"
[tools.presence]
kind = "command_exists"
name = "git"

[[runtimes]]
id = "go"
description = "Go toolchain"
depends_on = ["tool:git"]
config_required = ["go.env"]
os = ["linux", "darwin"]
arch = ["amd64", "arm64"]
[runtimes.presence]
kind = "command_exists"
name = "go"

[[packages]]
id = "ripgrep"
description = "Fast text search"
depends_on = ["tool:git"]
[packages.install]
provider = "brew"
package = "ripgrep"
[packages.presence]
kind = "command_exists"
name = "rg"

[[packages]]
id = "jq"
description = "JSON processor"
depends_on = ["tool:git"]
[packages.install]
provider = "brew"
package = "jq"
[packages.presence]
kind = "command_exists"
name = "jq"

[[dotfiles]]
id = "bash"
description = "Bash dotfiles"

[[bundles]]
id = "cli"
resources = ["tool:git", "package:ripgrep", "package:jq"]

[[profiles]]
id = "dev"
bundles = ["cli"]
resources = ["runtime:go"]
`)
}

func replaceCatalogPath(args []string, path string) []string {
	replaced := append([]string(nil), args...)
	for index, value := range replaced {
		if value == "../../catalog/bootstrap.toml" {
			replaced[index] = path
		}
	}
	return replaced
}

func assertDefaultCatalogPlanSmoke(t *testing.T) {
	t.Helper()
	const catalogPath = "../../catalog/bootstrap.toml"
	catalog, err := catalogtoml.LoadFile(catalogPath)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	profiles := make([]string, 0, len(catalog.Profiles))
	for profile := range catalog.Profiles {
		profiles = append(profiles, profile)
	}
	sort.Strings(profiles)
	if len(profiles) == 0 {
		t.Fatal("default catalog has no declared profiles")
	}
	for _, profile := range profiles {
		stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"})
		stubInstallationState(t, planning.InstallationState{})
		stubConfigState(t, planning.ConfigState{})
		stubDotfilesState(t, planning.InstallationState{})
		wantPlan, _, err := buildPlan(catalogPath, planning.PlanRequest{Profile: profile})
		if err != nil {
			t.Fatalf("buildPlan(%q) error = %v", profile, err)
		}
		var stdout, stderr bytes.Buffer
		if code := run([]string{"plan", "--profile", profile, "--catalog", catalogPath}, &stdout, &stderr); code != exitSuccess {
			t.Fatalf("run() exit code = %d, want %d; stderr=%q", code, exitSuccess, stderr.String())
		}
		out := stdout.String()
		if !strings.Contains(out, "Plan profile: "+profile) || !strings.Contains(out, "Steps:\n") || !strings.Contains(out, "Results:\n") {
			t.Fatalf("default plan smoke output = %q", out)
		}
		statusByRef := make(map[planning.ResourceRef]planning.PlanStepStatus, len(wantPlan.Results))
		for _, result := range wantPlan.Results {
			statusByRef[result.Ref] = result.Status
		}
		for index, step := range wantPlan.Plan.Steps {
			status := statusByRef[step.Ref]
			if status == "" {
				status = planning.PlanStepStatusPlanned
			}
			wantRenderedStep := fmt.Sprintf("%d. %s:%s [%s] %s\n", index+1, step.Ref.Kind, step.Ref.Name, status, step.Resource.Description)
			if count := strings.Count(out, wantRenderedStep); count != 1 {
				t.Fatalf("profile %q rendered step %q %d times, want exactly once; stdout=%q", profile, wantRenderedStep, count, out)
			}
		}
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

func readDotlinkReportFixture(t *testing.T, name string) []byte {
	t.Helper()
	fixture, err := os.ReadFile(filepath.Join("..", "..", "internal", "execution", "testdata", "dotlink-report", name))
	if err != nil {
		t.Fatalf("read dotlink report fixture %q: %v", name, err)
	}
	return fixture
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

func TestConfirmedCommandsCheckBrewFormulaBeforeInstall(t *testing.T) {
	catalogPath := writeFile(t, t.TempDir(), "catalog.toml", `
schema = "dniebles.catalog"
version = 1

[[packages]]
id = "json-tool"
description = "JSON processor"
[packages.install]
provider = "brew"
package = "jq"

[[profiles]]
id = "dev"
resources = ["package:json-tool"]
`)
	stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "darwin", Arch: "arm64"})
	stubInstallationState(t, planning.InstallationState{})
	stubConfigState(t, planning.ConfigState{})
	stubDotfilesState(t, planning.InstallationState{})
	stubBrewCommandExists(t, true)
	tests := []struct {
		name    string
		command string
		results []execution.CommandResult
		code    int
		calls   []execution.CommandRequest
		output  string
	}{
		{"apply installed", "apply", []execution.CommandResult{{Status: execution.CommandStatusSucceeded}}, exitSuccess, []execution.CommandRequest{{Executable: "brew", Args: []string{"list", "--formula", "jq"}, Timeout: 30 * time.Second}}, "already installed; no mutation attempted"},
		{"bootstrap installed", "bootstrap", []execution.CommandResult{{Status: execution.CommandStatusSucceeded}}, exitSuccess, []execution.CommandRequest{{Executable: "brew", Args: []string{"list", "--formula", "jq"}, Timeout: 30 * time.Second}}, "already installed; no mutation attempted"},
		{"apply explicitly absent", "apply", []execution.CommandResult{{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "Error: No such keg: jq", Err: errors.New("exit 1")}, {Status: execution.CommandStatusSucceeded}}, exitSuccess, []execution.CommandRequest{{Executable: "brew", Args: []string{"list", "--formula", "jq"}, Timeout: 30 * time.Second}, {Executable: "brew", Args: []string{"install", "jq"}}}, "installed jq"},
		{"bootstrap explicitly absent", "bootstrap", []execution.CommandResult{{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "Error: No such keg: jq", Err: errors.New("exit 1")}, {Status: execution.CommandStatusSucceeded}}, exitSuccess, []execution.CommandRequest{{Executable: "brew", Args: []string{"list", "--formula", "jq"}, Timeout: 30 * time.Second}, {Executable: "brew", Args: []string{"install", "jq"}}}, "installed jq"},
		{"apply timed out", "apply", []execution.CommandResult{{Status: execution.CommandStatusTimedOut, Err: context.DeadlineExceeded}}, exitFailure, []execution.CommandRequest{{Executable: "brew", Args: []string{"list", "--formula", "jq"}, Timeout: 30 * time.Second}}, "presence could not be determined; no mutation attempted"},
		{"bootstrap timed out", "bootstrap", []execution.CommandResult{{Status: execution.CommandStatusTimedOut, Err: context.DeadlineExceeded}}, exitFailure, []execution.CommandRequest{{Executable: "brew", Args: []string{"list", "--formula", "jq"}, Timeout: 30 * time.Second}}, "presence could not be determined; no mutation attempted"},
		{"bootstrap runner error", "bootstrap", []execution.CommandResult{{Status: execution.CommandStatusFailed, Err: errors.New("runner failed")}}, exitFailure, []execution.CommandRequest{{Executable: "brew", Args: []string{"list", "--formula", "jq"}, Timeout: 30 * time.Second}}, "presence could not be determined; no mutation attempted"},
		{"bootstrap unclassified non-zero", "bootstrap", []execution.CommandResult{{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "Error: unknown brew failure", Err: errors.New("exit 1")}}, exitFailure, []execution.CommandRequest{{Executable: "brew", Args: []string{"list", "--formula", "jq"}, Timeout: 30 * time.Second}}, "presence could not be determined; no mutation attempted"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &sequenceCommandRunner{results: tt.results}
			stubExecutionFactories(t, func() execution.CommandRunner { return runner }, func(kind planning.ResourceKind, commandRunner execution.CommandRunner, exists execution.CommandExists) execution.Installer {
				return execution.NewHomebrewInstaller(kind, commandRunner, exists)
			}, func(commandRunner execution.CommandRunner) execution.Installer {
				return execution.NewDotfilesInstaller(execution.NewLocalDotfilesProvider(commandRunner, execution.DotfilesBaseResolver{}))
			})
			var stdout, stderr bytes.Buffer
			if got := run([]string{tt.command, "--profile", "dev", "--catalog", catalogPath, "--yes"}, &stdout, &stderr); got != tt.code {
				t.Fatalf("exit code = %d, want %d; stderr=%q", got, tt.code, stderr.String())
			}
			if !reflect.DeepEqual(runner.calls, tt.calls) {
				t.Fatalf("command calls = %#v, want %#v", runner.calls, tt.calls)
			}
			if !strings.Contains(stdout.String(), tt.output) {
				t.Fatalf("stdout = %q, want %q", stdout.String(), tt.output)
			}
		})
	}
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

type sequenceCommandRunner struct {
	results []execution.CommandResult
	calls   []execution.CommandRequest
}

func (r *sequenceCommandRunner) RunCommand(_ context.Context, req execution.CommandRequest) execution.CommandResult {
	r.calls = append(r.calls, req)
	result := r.results[len(r.calls)-1]
	result.Request = req
	return result
}

func TestRunApplyAndBootstrapAptPackageDetection(t *testing.T) {
	catalogPath := writeFile(t, t.TempDir(), "catalog.toml", `
schema = "dniebles.catalog"
version = 1

[[packages]]
id = "ripgrep"
description = "Fast text search"
[packages.install]
provider = "apt"
package = "ripgrep"

[[profiles]]
id = "dev"
resources = ["package:ripgrep"]
`)
	tests := []struct {
		name       string
		command    string
		args       []string
		facts      planning.EnvironmentFacts
		available  map[string]bool
		results    []execution.CommandResult
		wantCode   int
		wantCalls  []execution.CommandRequest
		wantOutput string
	}{
		{
			name:      "apply linux installed hold skips apt-get",
			command:   "apply",
			args:      []string{"--profile", "dev", "--catalog", catalogPath, "--yes"},
			facts:     planning.EnvironmentFacts{OS: "linux"},
			available: map[string]bool{"dpkg-query": true, "apt-get": true},
			results: []execution.CommandResult{
				{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: "hold ok installed"},
			},
			wantCode: exitSuccess,
			wantCalls: []execution.CommandRequest{
				{Executable: "dpkg-query", Args: []string{"--show", "--showformat=${Status}", "ripgrep"}, Timeout: 30 * time.Second},
			},
			wantOutput: "package:ripgrep [unchanged] already installed; no mutation attempted",
		},
		{
			name:      "bootstrap linux partial dispatches apt-get",
			command:   "bootstrap",
			args:      []string{"--profile", "dev", "--catalog", catalogPath, "--yes"},
			facts:     planning.EnvironmentFacts{OS: "linux"},
			available: map[string]bool{"dpkg-query": true, "apt-get": true},
			results: []execution.CommandResult{
				{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: "install ok unpacked"},
				{Status: execution.CommandStatusSucceeded, ExitCode: 0},
			},
			wantCode: exitSuccess,
			wantCalls: []execution.CommandRequest{
				{Executable: "dpkg-query", Args: []string{"--show", "--showformat=${Status}", "ripgrep"}, Timeout: 30 * time.Second},
				{Executable: "apt-get", Args: []string{"install", "-y", "--", "ripgrep"}, Timeout: 10 * time.Minute},
			},
			wantOutput: "package:ripgrep [changed] installed ripgrep with APT",
		},
		{
			name:      "apply linux not found dispatches apt-get",
			command:   "apply",
			args:      []string{"--profile", "dev", "--catalog", catalogPath, "--yes"},
			facts:     planning.EnvironmentFacts{OS: "linux"},
			available: map[string]bool{"dpkg-query": true, "apt-get": true},
			results: []execution.CommandResult{
				{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "dpkg-query: no packages found matching ripgrep", Err: errors.New("exit 1")},
				{Status: execution.CommandStatusSucceeded, ExitCode: 0},
			},
			wantCode: exitSuccess,
			wantCalls: []execution.CommandRequest{
				{Executable: "dpkg-query", Args: []string{"--show", "--showformat=${Status}", "ripgrep"}, Timeout: 30 * time.Second},
				{Executable: "apt-get", Args: []string{"install", "-y", "--", "ripgrep"}, Timeout: 10 * time.Minute},
			},
			wantOutput: "package:ripgrep [changed] installed ripgrep with APT",
		},
		{
			name:      "apply linux unknown does not dispatch apt-get",
			command:   "apply",
			args:      []string{"--profile", "dev", "--catalog", catalogPath, "--yes"},
			facts:     planning.EnvironmentFacts{OS: "linux"},
			available: map[string]bool{"dpkg-query": true, "apt-get": true},
			results: []execution.CommandResult{
				{Status: execution.CommandStatusTimedOut, Err: context.DeadlineExceeded},
			},
			wantCode: exitFailure,
			wantCalls: []execution.CommandRequest{
				{Executable: "dpkg-query", Args: []string{"--show", "--showformat=${Status}", "ripgrep"}, Timeout: 30 * time.Second},
			},
			wantOutput: "package:ripgrep [failed] APT package presence could not be determined; no mutation attempted",
		},
		{
			name:       "default does not probe dpkg-query",
			command:    "apply",
			args:       []string{"--profile", "dev", "--catalog", catalogPath},
			facts:      planning.EnvironmentFacts{OS: "linux"},
			available:  map[string]bool{},
			results:    nil,
			wantCode:   exitSuccess,
			wantCalls:  nil,
			wantOutput: "package:ripgrep [not supported yet] noop installer does not perform real installation",
		},
		{
			name:       "dry run does not probe dpkg-query",
			command:    "apply",
			args:       []string{"--profile", "dev", "--catalog", catalogPath, "--dry-run"},
			facts:      planning.EnvironmentFacts{OS: "linux"},
			available:  map[string]bool{},
			results:    nil,
			wantCode:   exitSuccess,
			wantCalls:  nil,
			wantOutput: "package:ripgrep [not supported yet] noop installer does not perform real installation",
		},
		{
			name:       "plan does not probe dpkg-query",
			command:    "plan",
			args:       []string{"--profile", "dev", "--catalog", catalogPath},
			facts:      planning.EnvironmentFacts{OS: "linux"},
			available:  map[string]bool{},
			results:    nil,
			wantCode:   exitSuccess,
			wantCalls:  nil,
			wantOutput: "package:ripgrep [planned]",
		},
		{
			name:       "non linux confirmed does not probe dpkg-query",
			command:    "apply",
			args:       []string{"--profile", "dev", "--catalog", catalogPath, "--yes"},
			facts:      planning.EnvironmentFacts{OS: "darwin"},
			available:  map[string]bool{"dpkg-query": true},
			results:    nil,
			wantCode:   exitFailure,
			wantCalls:  nil,
			wantOutput: "package:ripgrep [failed] apt execution unsupported_os on darwin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubEnvironmentFacts(t, tt.facts)
			stubInstallationState(t, planning.InstallationState{})
			stubConfigState(t, planning.ConfigState{})
			stubDotfilesState(t, planning.InstallationState{})
			stubBrewCommandExists(t, false)
			originalExists := aptCommandExists
			aptCommandExists = func(name string) bool {
				if tt.available == nil {
					t.Fatalf("command %q must not be probed", name)
				}
				return tt.available[name]
			}
			t.Cleanup(func() { aptCommandExists = originalExists })

			var runner execution.CommandRunner
			if tt.results != nil {
				runner = &sequenceCommandRunner{results: tt.results}
			} else {
				runner = &recordingCommandRunner{result: execution.CommandResult{Status: execution.CommandStatusSucceeded}}
			}
			stubExecutionFactories(t, func() execution.CommandRunner { return runner }, newHomebrewInstaller, newDotfilesInstaller)

			args := append([]string{tt.command}, tt.args...)
			var stdout, stderr bytes.Buffer
			if code := run(args, &stdout, &stderr); code != tt.wantCode {
				t.Fatalf("exit code = %d, want %d; stdout=%q stderr=%q", code, tt.wantCode, stdout.String(), stderr.String())
			}

			var calls []execution.CommandRequest
			switch r := runner.(type) {
			case *sequenceCommandRunner:
				calls = r.calls
			case *recordingCommandRunner:
				calls = r.calls
			}
			if !reflect.DeepEqual(calls, tt.wantCalls) {
				t.Fatalf("command calls = %#v, want %#v", calls, tt.wantCalls)
			}
			if !strings.Contains(stdout.String(), tt.wantOutput) {
				t.Fatalf("stdout = %q, want it to contain %q", stdout.String(), tt.wantOutput)
			}
			if stderr.String() != "" {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
	}
}

func TestRunBootstrapHelp(t *testing.T) {
	for _, tt := range []struct {
		name       string
		args       []string
		wantOutput string
	}{
		{
			name:       "root help lists bootstrap",
			args:       []string{"--help"},
			wantOutput: "bootstrap  Execute an explicit selection through the safe apply workflow",
		},
		{
			name:       "long command help explains explicit targets",
			args:       []string{"bootstrap", "--help"},
			wantOutput: "Usage: dbootstrap bootstrap [--profile <name>] [--resource <kind:name>] [--catalog <path>] [--dry-run] [--yes [--sudo]]",
		},
		{
			name:       "short command help explains explicit targets",
			args:       []string{"bootstrap", "-h"},
			wantOutput: "Usage: dbootstrap bootstrap [--profile <name>] [--resource <kind:name>] [--catalog <path>] [--dry-run] [--yes [--sudo]]",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			originalDetect := detectEnvironmentFacts
			detectEnvironmentFacts = func() planning.EnvironmentFacts {
				t.Fatal("help must not detect the environment")
				return planning.EnvironmentFacts{}
			}
			t.Cleanup(func() { detectEnvironmentFacts = originalDetect })

			var stdout, stderr bytes.Buffer
			if code := run(tt.args, &stdout, &stderr); code != exitSuccess {
				t.Fatalf("exit code = %d, want %d; stdout=%q stderr=%q", code, exitSuccess, stdout.String(), stderr.String())
			}
			if !strings.Contains(stdout.String(), tt.wantOutput) {
				t.Fatalf("stdout = %q, want it to contain %q", stdout.String(), tt.wantOutput)
			}
		})
	}
}

func TestRunApplyHelpRetainsParserUsageFailure(t *testing.T) {
	for _, alias := range []string{"-h", "--help"} {
		t.Run(alias, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := run([]string{"apply", alias}, &stdout, &stderr); code != exitUsage {
				t.Fatalf("exit code = %d, want %d; stdout=%q stderr=%q", code, exitUsage, stdout.String(), stderr.String())
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if !strings.Contains(stderr.String(), "Usage of apply:") || !strings.Contains(stderr.String(), "Usage: dbootstrap apply") {
				t.Fatalf("stderr = %q, want parser and command usage", stderr.String())
			}
		})
	}
}

func TestRunBootstrapMatchesApplyAcrossSafetyModes(t *testing.T) {
	for _, tt := range []struct {
		name           string
		flags          []string
		results        []execution.CommandResult
		wantExecutions int
		wantExecutable string
	}{
		{name: "default", wantExecutions: 0},
		{name: "dry run", flags: []string{"--dry-run"}, wantExecutions: 0},
		{
			name:  "confirmed",
			flags: []string{"--yes"},
			results: []execution.CommandResult{
				{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "dpkg-query: no packages found matching ripgrep", Err: errors.New("exit 1")},
				{Status: execution.CommandStatusSucceeded},
			},
			wantExecutions: 2,
			wantExecutable: "apt-get",
		},
		{
			name:  "confirmed sudo",
			flags: []string{"--yes", "--sudo"},
			results: []execution.CommandResult{
				{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "dpkg-query: no packages found matching ripgrep", Err: errors.New("exit 1")},
				{Status: execution.CommandStatusSucceeded},
			},
			wantExecutions: 2,
			wantExecutable: "sudo",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			catalogPath := writeAptCatalog(t)
			stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux"})
			stubInstallationState(t, planning.InstallationState{})
			stubConfigState(t, planning.ConfigState{})
			stubDotfilesState(t, planning.InstallationState{})
			originalExists := aptCommandExists
			aptCommandExists = func(name string) bool { return name == "dpkg-query" || name == "apt-get" || name == "sudo" }
			t.Cleanup(func() { aptCommandExists = originalExists })

			outputs := make([]string, 0, 2)
			codes := make([]int, 0, 2)
			for _, command := range []string{"apply", "bootstrap"} {
				var runner execution.CommandRunner
				if tt.results != nil {
					runner = &sequenceCommandRunner{results: tt.results}
				} else {
					runner = &recordingCommandRunner{result: execution.CommandResult{Status: execution.CommandStatusSucceeded}}
				}
				stubExecutionFactories(t, func() execution.CommandRunner { return runner }, newHomebrewInstaller, newDotfilesInstaller)
				args := append([]string{command, "--profile", "apt-fixture", "--catalog", catalogPath}, tt.flags...)
				var stdout, stderr bytes.Buffer
				codes = append(codes, run(args, &stdout, &stderr))
				outputs = append(outputs, stdout.String()+stderr.String())

				var calls []execution.CommandRequest
				switch r := runner.(type) {
				case *sequenceCommandRunner:
					calls = r.calls
				case *recordingCommandRunner:
					calls = r.calls
				}
				if len(calls) != tt.wantExecutions {
					t.Fatalf("%s command calls = %#v, want %d", command, calls, tt.wantExecutions)
				}
				if tt.wantExecutable != "" && calls[len(calls)-1].Executable != tt.wantExecutable {
					t.Fatalf("%s last executable = %q, want %q", command, calls[len(calls)-1].Executable, tt.wantExecutable)
				}
			}
			if codes[0] != codes[1] {
				t.Fatalf("apply exit = %d, bootstrap exit = %d", codes[0], codes[1])
			}
			if outputs[0] != outputs[1] {
				t.Fatalf("apply output = %q, bootstrap output = %q", outputs[0], outputs[1])
			}
		})
	}
}

func TestRunApplyLikeRejectsSyntacticInputBeforeProbing(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		want string
	}{
		{name: "missing target", args: nil, want: "--profile or --resource is required"},
		{name: "malformed resource", args: []string{"--resource", "package"}, want: "expected kind:name"},
		{name: "positional", args: []string{"--profile", "dev", "extra"}, want: "unexpected argument \"extra\""},
		{name: "conflicting modes", args: []string{"--profile", "dev", "--dry-run", "--yes"}, want: "--dry-run and --yes cannot be combined"},
		{name: "sudo requires confirmation", args: []string{"--profile", "dev", "--sudo"}, want: "--sudo requires --yes"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for _, command := range []string{"apply", "bootstrap"} {
				t.Run(command, func(t *testing.T) {
					originalDetect := detectEnvironmentFacts
					detectEnvironmentFacts = func() planning.EnvironmentFacts {
						t.Fatal("syntactic validation must not detect the environment")
						return planning.EnvironmentFacts{}
					}
					t.Cleanup(func() { detectEnvironmentFacts = originalDetect })

					var stdout, stderr bytes.Buffer
					if code := run(append([]string{command}, tt.args...), &stdout, &stderr); code != exitUsage {
						t.Fatalf("exit code = %d, want %d; stdout=%q stderr=%q", code, exitUsage, stdout.String(), stderr.String())
					}
					if !strings.Contains(stderr.String(), tt.want) {
						t.Fatalf("stderr = %q, want it to contain %q", stderr.String(), tt.want)
					}
				})
			}
		})
	}
}

func TestRunBootstrapMatchesApplyForUnknownProfile(t *testing.T) {
	catalogPath := writeAptCatalog(t)
	for _, command := range []string{"apply", "bootstrap"} {
		t.Run(command, func(t *testing.T) {
			calls := 0
			originalDetect := detectEnvironmentFacts
			detectEnvironmentFacts = func() planning.EnvironmentFacts {
				calls++
				return planning.EnvironmentFacts{OS: "linux"}
			}
			t.Cleanup(func() { detectEnvironmentFacts = originalDetect })
			stubInstallationState(t, planning.InstallationState{})
			stubConfigState(t, planning.ConfigState{})
			stubDotfilesState(t, planning.InstallationState{})

			var stdout, stderr bytes.Buffer
			if code := run([]string{command, "--profile", "unknown", "--catalog", catalogPath}, &stdout, &stderr); code != exitFailure {
				t.Fatalf("exit code = %d, want %d; stdout=%q stderr=%q", code, exitFailure, stdout.String(), stderr.String())
			}
			if calls != 1 {
				t.Fatalf("environment detection calls = %d, want 1", calls)
			}
			if !strings.Contains(stderr.String(), "unknown profile \"unknown\"") {
				t.Fatalf("stderr = %q, want unknown profile diagnostic", stderr.String())
			}
		})
	}
}

func TestRunBootstrapMatchesApplyForUnknownResource(t *testing.T) {
	catalogPath := writeAptCatalog(t)
	outputs := make([]string, 0, 2)
	codes := make([]int, 0, 2)
	probes := make([]int, 0, 2)
	for _, command := range []string{"apply", "bootstrap"} {
		calls := 0
		originalDetect := detectEnvironmentFacts
		detectEnvironmentFacts = func() planning.EnvironmentFacts {
			calls++
			return planning.EnvironmentFacts{OS: "linux"}
		}
		t.Cleanup(func() { detectEnvironmentFacts = originalDetect })
		stubInstallationState(t, planning.InstallationState{})
		stubConfigState(t, planning.ConfigState{})
		stubDotfilesState(t, planning.InstallationState{})

		var stdout, stderr bytes.Buffer
		codes = append(codes, run([]string{command, "--resource", "package:unknown", "--catalog", catalogPath}, &stdout, &stderr))
		outputs = append(outputs, stdout.String()+stderr.String())
		probes = append(probes, calls)
	}
	if codes[0] != exitFailure || codes[1] != exitFailure {
		t.Fatalf("exit codes = %#v, want both %d", codes, exitFailure)
	}
	if outputs[0] != outputs[1] {
		t.Fatalf("apply output = %q, bootstrap output = %q", outputs[0], outputs[1])
	}
	if probes[0] != 1 || probes[1] != 1 {
		t.Fatalf("environment probes = %#v, want [1 1]", probes)
	}
	if !strings.Contains(outputs[1], "unknown resource package:unknown") {
		t.Fatalf("stderr = %q, want unknown resource diagnostic", outputs[1])
	}
}

func TestRunBootstrapMatchesApplyForPrerequisites(t *testing.T) {
	tests := []struct {
		name         string
		catalogPath  func(t *testing.T) string
		profile      string
		facts        planning.EnvironmentFacts
		configState  planning.ConfigState
		wantCode     int
		wantOutput   string
		wantProbes   int
		wantCommands int
	}{
		{
			name:         "missing catalog",
			catalogPath:  func(t *testing.T) string { return filepath.Join(t.TempDir(), "missing.toml") },
			profile:      "dev",
			wantCode:     exitFailure,
			wantOutput:   "error: load catalog",
			wantProbes:   0,
			wantCommands: 0,
		},
		{
			name:        "missing required config",
			catalogPath: writePrimaryCatalog,
			profile:     "dev",
			facts:       planning.EnvironmentFacts{OS: "linux", Arch: "amd64"},
			configState: planning.ConfigState{},
			wantCode:    exitSuccess,
			wantOutput:  "Execution Report",
			wantProbes:  1,
		},
		{
			name:        "environment mismatch",
			catalogPath: writeLinuxOnlyCatalog,
			profile:     "linux-only",
			facts:       planning.EnvironmentFacts{OS: "darwin"},
			configState: planning.ConfigState{},
			wantCode:    exitSuccess,
			wantOutput:  "No actionable steps were selected; nothing to apply.",
			wantProbes:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			catalogPath := tt.catalogPath(t)
			outputs := make([]string, 0, 2)
			codes := make([]int, 0, 2)
			probes := make([]int, 0, 2)
			commands := make([]int, 0, 2)
			for _, command := range []string{"apply", "bootstrap"} {
				calls := 0
				originalDetect := detectEnvironmentFacts
				detectEnvironmentFacts = func() planning.EnvironmentFacts {
					calls++
					return tt.facts
				}
				t.Cleanup(func() { detectEnvironmentFacts = originalDetect })
				stubInstallationState(t, planning.InstallationState{})
				stubConfigState(t, tt.configState)
				stubDotfilesState(t, planning.InstallationState{})
				runner := &recordingCommandRunner{result: execution.CommandResult{Status: execution.CommandStatusSucceeded}}
				stubExecutionFactories(t, func() execution.CommandRunner { return runner }, newHomebrewInstaller, newDotfilesInstaller)

				var stdout, stderr bytes.Buffer
				codes = append(codes, run([]string{command, "--profile", tt.profile, "--catalog", catalogPath}, &stdout, &stderr))
				outputs = append(outputs, stdout.String()+stderr.String())
				probes = append(probes, calls)
				commands = append(commands, len(runner.calls))
			}
			if codes[0] != tt.wantCode || codes[1] != tt.wantCode {
				t.Fatalf("exit codes = %#v, want both %d", codes, tt.wantCode)
			}
			if outputs[0] != outputs[1] {
				t.Fatalf("apply output = %q, bootstrap output = %q", outputs[0], outputs[1])
			}
			if probes[0] != tt.wantProbes || probes[1] != tt.wantProbes {
				t.Fatalf("environment probes = %#v, want both %d", probes, tt.wantProbes)
			}
			if commands[0] != tt.wantCommands || commands[1] != tt.wantCommands {
				t.Fatalf("command calls = %#v, want both %d", commands, tt.wantCommands)
			}
			if !strings.Contains(outputs[1], tt.wantOutput) {
				t.Fatalf("bootstrap output = %q, want %q", outputs[1], tt.wantOutput)
			}
		})
	}
}

func TestRunBootstrapMatchesApplyForPartialFailure(t *testing.T) {
	outputs := make([]string, 0, 2)
	codes := make([]int, 0, 2)
	for _, command := range []string{"apply", "bootstrap"} {
		stubEnvironmentFacts(t, planning.EnvironmentFacts{OS: "linux"})
		stubInstallationState(t, planning.InstallationState{})
		stubConfigState(t, planning.ConfigState{})
		stubDotfilesState(t, planning.InstallationState{})
		originalExists := aptCommandExists
		aptCommandExists = func(name string) bool { return name == "dpkg-query" || name == "apt-get" }
		t.Cleanup(func() { aptCommandExists = originalExists })
		runner := &sequenceCommandRunner{results: []execution.CommandResult{
			{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "dpkg-query: no packages found matching first", Err: errors.New("exit 1")},
			{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "dpkg-query: no packages found matching second", Err: errors.New("exit 1")},
			{Status: execution.CommandStatusSucceeded},
			{Status: execution.CommandStatusFailed},
		}}
		stubExecutionFactories(t, func() execution.CommandRunner { return runner }, newHomebrewInstaller, newDotfilesInstaller)

		var stdout, stderr bytes.Buffer
		codes = append(codes, run([]string{command, "--profile", "two-apt", "--catalog", writeTwoAptCatalog(t), "--yes"}, &stdout, &stderr))
		outputs = append(outputs, stdout.String()+stderr.String())
	}
	if codes[0] != exitFailure || codes[1] != exitFailure {
		t.Fatalf("exit codes = %#v, want both %d", codes, exitFailure)
	}
	if outputs[0] != outputs[1] {
		t.Fatalf("apply output = %q, bootstrap output = %q", outputs[0], outputs[1])
	}
	for _, output := range outputs {
		first := strings.Index(output, "package:first [changed]")
		second := strings.Index(output, "package:second [failed]")
		if first < 0 || second < 0 || first >= second {
			t.Fatalf("partial report = %q, want package:first [changed] before package:second [failed]", output)
		}
	}
}

func writeTwoAptCatalog(t *testing.T) string {
	t.Helper()
	return writeFile(t, t.TempDir(), "two-apt.toml", `schema = "dniebles.catalog"
version = 1

[[packages]]
id = "first"
description = "First APT fixture"
[packages.install]
provider = "apt"
package = "first"

[[packages]]
id = "second"
description = "Second APT fixture"
[packages.install]
provider = "apt"
package = "second"

[[profiles]]
id = "two-apt"
resources = ["package:first", "package:second"]
`)
}

func writeLinuxOnlyCatalog(t *testing.T) string {
	t.Helper()
	return writeFile(t, t.TempDir(), "linux-only.toml", `schema = "dniebles.catalog"
version = 1

[[packages]]
id = "linux-only"
description = "Linux-only fixture"
os = ["linux"]

[[profiles]]
id = "linux-only"
resources = ["package:linux-only"]
`)
}
