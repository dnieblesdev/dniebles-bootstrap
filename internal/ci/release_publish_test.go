package ci

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

const releasePublishWorkflow = "../../.github/workflows/release-publish.yml"

// extractBlock returns the contiguous indented YAML block that starts with the
// given header line. It stops when a non-empty line at the same or lower
// indentation is found.
func extractBlock(content, header string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if line != header {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		end := len(lines)
		for j := i + 1; j < len(lines); j++ {
			next := lines[j]
			if next == "" {
				continue
			}
			nextIndent := len(next) - len(strings.TrimLeft(next, " "))
			if nextIndent <= indent {
				end = j
				break
			}
		}
		return strings.Join(lines[i:end], "\n")
	}
	return ""
}

func readWorkflow(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(releasePublishWorkflow)
	if err != nil {
		t.Fatalf("read %s: %v", releasePublishWorkflow, err)
	}
	return string(data)
}

func TestReleasePublish_GlobalPermissions(t *testing.T) {
	content := readWorkflow(t)

	block := extractBlock(content, "permissions:")
	if block == "" {
		t.Fatalf("workflow must have a global permissions block")
	}

	if !strings.Contains(block, "contents: read") {
		t.Errorf("global permissions must grant contents: read")
	}
	if strings.Contains(block, "actions: write") {
		t.Errorf("global permissions must not grant actions: write")
	}
}

func TestReleasePublish_ValidateJobPermissions(t *testing.T) {
	content := readWorkflow(t)

	job := extractBlock(content, "  validate:")
	if job == "" {
		t.Fatalf("workflow must have a validate job")
	}

	perm := extractBlock(job, "    permissions:")
	if perm == "" {
		t.Fatalf("validate job must declare permissions")
	}
	if !strings.Contains(perm, "contents: read") {
		t.Errorf("validate job permissions must grant contents: read")
	}
	if strings.Contains(perm, "actions: write") {
		t.Errorf("validate job permissions must not grant actions: write")
	}
}

func TestReleasePublish_PublishJobPermissions(t *testing.T) {
	content := readWorkflow(t)

	job := extractBlock(content, "  publish:")
	if job == "" {
		t.Fatalf("workflow must have a publish job")
	}

	perm := extractBlock(job, "    permissions:")
	if perm == "" {
		t.Fatalf("publish job must declare permissions")
	}
	if !strings.Contains(perm, "contents: write") {
		t.Errorf("publish job permissions must grant contents: write")
	}
	if !strings.Contains(perm, "actions: read") {
		t.Errorf("publish job permissions must grant actions: read")
	}
}

func TestReleasePublish_NeedsValidationBarrier(t *testing.T) {
	content := readWorkflow(t)

	if !strings.Contains(content, "\n  build:\n    needs: validate\n") {
		t.Errorf("build job must depend on validate")
	}
	if !strings.Contains(content, "\n  publish:\n    needs: [validate, build]\n") {
		t.Errorf("publish job must depend on validate and build")
	}
}

func runValidateCommand(t *testing.T, version string) (output string, err error) {
	t.Helper()
	cmd := exec.Command("go", "run", "../version/cmd/validate", "--release", "--version", version)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func TestReleasePublish_ValidateCommandAcceptsValidVersions(t *testing.T) {
	validVersions := []struct {
		version    string
		prerelease string
	}{
		{"v1.2.3", "prerelease=false"},
		{"v1.2.3-rc.1", "prerelease=true"},
	}
	for _, tc := range validVersions {
		t.Run(tc.version, func(t *testing.T) {
			got, err := runValidateCommand(t, tc.version)
			if err != nil {
				t.Fatalf("expected validation to pass for %q: %v", tc.version, err)
			}
			if got != tc.prerelease {
				t.Fatalf("validation output for %q = %q, want %q", tc.version, got, tc.prerelease)
			}
		})
	}
}

func TestReleasePublish_ValidateCommandRejectsInvalidVersions(t *testing.T) {
	invalidVersions := []string{"1.2.3", "v1", "v1.2", "not-a-version"}
	for _, v := range invalidVersions {
		t.Run(v, func(t *testing.T) {
			_, err := runValidateCommand(t, v)
			if err == nil {
				t.Fatalf("expected validation to fail for %q", v)
			}
		})
	}
}

func TestReleasePublish_PublishJobPreservesReleaseBehavior(t *testing.T) {
	content := readWorkflow(t)

	job := extractBlock(content, "  publish:")
	if job == "" {
		t.Fatalf("workflow must have a publish job")
	}

	if !strings.Contains(job, "Guard existing tag and release") {
		t.Errorf("publish job must guard against duplicate tags and releases")
	}
	if !strings.Contains(job, "Create GitHub Release") {
		t.Errorf("publish job must contain the release creation step")
	}
}
