package version

import (
	"strings"
	"testing"
)

func TestValidateReleaseTag(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		wantPrerelease bool
		wantErr        bool
	}{
		{"stable", "v1.2.3", false, false},
		{"zero core", "v0.0.0", false, false},
		{"prerelease", "v1.2.3-rc.1", true, false},
		{"prerelease numeric identifier", "v1.2.3-0", true, false},
		{"multiple prerelease identifiers", "v1.2.3-alpha.1.beta.2", true, false},
		{"build metadata", "v1.2.3+build.123", false, false},
		{"prerelease and build metadata", "v1.2.3-rc.1+build.123", true, false},
		{"unprefixed", "1.2.3", false, true},
		{"partial one component", "v1", false, true},
		{"partial two components", "v1.2", false, true},
		{"leading zero major", "v01.2.3", false, true},
		{"leading zero minor", "v1.02.3", false, true},
		{"leading zero patch", "v1.2.03", false, true},
		{"leading zero prerelease numeric", "v1.2.3-rc.01", false, true},
		{"empty", "", false, true},
		{"too long", strings.Repeat("a", 65), false, true},
		{"invalid characters", "v1.2.3; rm -rf /", false, true},
		{"empty prerelease", "v1.2.3-", false, true},
		{"empty build metadata", "v1.2.3+", false, true},
		{"underscore separator", "v1.2.3_rc1", false, true},
		{"space", "v1 2 3", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrerelease, err := ValidateReleaseTag(tt.version)
			if tt.wantErr && err == nil {
				t.Fatalf("ValidateReleaseTag(%q) = (%v, nil), want error", tt.version, gotPrerelease)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ValidateReleaseTag(%q) = (%v, %v), want nil", tt.version, gotPrerelease, err)
			}
			if !tt.wantErr && gotPrerelease != tt.wantPrerelease {
				t.Fatalf("ValidateReleaseTag(%q) prerelease = %v, want %v", tt.version, gotPrerelease, tt.wantPrerelease)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"empty is valid", "", false},
		{"dev default", "dev", false},
		{"semver with v", "v1.2.3", false},
		{"semver without v", "1.2.3", false},
		{"semver prerelease", "v1.2.3-alpha.1", false},
		{"semver build metadata", "v1.2.3+build.123", false},
		{"semver prerelease and build", "v1.2.3-alpha.1+build.123", false},
		{"git describe", "v0.1.2-3-gabcdef", false},
		{"git describe dirty", "v0.1.2-dirty", false},
		{"short commit hash", "abc1234", false},
		{"underscore", "v1.2.3_rc1", false},
		{"slash is invalid", "v1.2.3/extra", true},
		{"space is invalid", "v1 2 3", true},
		{"shell injection semicolon", "v1.2.3; rm -rf /", true},
		{"command substitution dollar", "v1.2.3$(whoami)", true},
		{"command substitution backtick", "v1.2.3`whoami`", true},
		{"leading dot", ".v1.2.3", true},
		{"leading hyphen", "-v1.2.3", true},
		{"too long", strings.Repeat("a", 65), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.version)
			if tt.wantErr && err == nil {
				t.Fatalf("Validate(%q) = nil, want error", tt.version)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("Validate(%q) = %v, want nil", tt.version, err)
			}
		})
	}
}
