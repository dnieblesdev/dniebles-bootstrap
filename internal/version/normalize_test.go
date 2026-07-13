package version

import "testing"

func TestNormalizeGitVersion(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"normal tag", "v1.2.3", "v1.2.3"},
		{"slash branch", "feature/new-thing", "feature-new-thing"},
		{"multiple slashes", "bugfix/core/memory", "bugfix-core-memory"},
		{"spaces", "v1 2 3", "v1-2-3"},
		{"invalid chars", "v1.2.3; rm -rf /", "v1.2.3-rm--rf"},
		{"empty", "", "dev"},
		{"only separators after trim", "-/_-", "dev"},
		{"leading separator", "-v1.2.3", "v1.2.3"},
		{"trailing separator", "v1.2.3-", "v1.2.3"},
		{"plus becomes hyphen", "v1.2.3+build.123", "v1.2.3-build.123"},
		{"coalesce consecutive invalid", "a!!!b", "a-b"},
		{"underscore preserved", "v1_2_3", "v1_2_3"},
		{"unicode becomes hyphen", "v1.2.3-rc.1\u00A0", "v1.2.3-rc.1"},
		{"consecutive dots preserved", "v1..2..3", "v1..2..3"},
		{"only invalid", "!@#$%", "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeGitVersion(tt.in)
			if got != tt.want {
				t.Fatalf("NormalizeGitVersion(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
