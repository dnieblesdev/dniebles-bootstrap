package version

import (
	"strings"
	"testing"
)

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
