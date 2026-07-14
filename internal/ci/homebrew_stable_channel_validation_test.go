package ci

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

const homebrewStableChannelValidationWorkflow = "../../.github/workflows/homebrew-formula-pr-validation.yml"

func readHomebrewStableChannelValidationWorkflow(t *testing.T) string {
	t.Helper()

	data, err := os.ReadFile(homebrewStableChannelValidationWorkflow)
	if err != nil {
		t.Fatalf("read %s: %v", homebrewStableChannelValidationWorkflow, err)
	}
	return string(data)
}

func TestHomebrewStableChannelValidation_PRHeadLocalFormulaContract(t *testing.T) {
	content := readHomebrewStableChannelValidationWorkflow(t)

	for _, want := range []string{
		"pull_request:",
		"permissions:\n  contents: read",
		"ubuntu-24.04",
		"ubuntu-24.04-arm",
		"macos-14",
		"ref: ${{ github.event.pull_request.head.sha }}",
		"persist-credentials: false",
		"local/pr-candidate/dbootstrap",
		"homebrew-pr-candidate",
		"original-formula-sha256.txt",
		"staged-formula-sha256.txt",
		"cmp homebrew-receipt/original-formula-sha256.txt homebrew-receipt/staged-formula-sha256.txt",
		"brew audit --strict --formula local/pr-candidate/dbootstrap",
		"brew style local/pr-candidate/dbootstrap",
		"HOMEBREW_NO_INSTALL_FROM_API=1 brew install --build-from-source local/pr-candidate/dbootstrap",
		"brew test local/pr-candidate/dbootstrap",
		"if: always()",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("workflow must contain %q", want)
		}
	}

	for _, forbidden := range []string{
		"workflow_dispatch:", "push:", "workflow_call:", "contents: write", "GITHUB_TOKEN",
		"secrets.", "persist-credentials: true", "gh release", "release-publish.yml",
		"git tag", "git push", "upload-release-asset", "docker", "qemu", "brew tap ", "git init", "git clone",
		"dnieblesdev/dniebles-bootstrap/dbootstrap",
		"brew audit --strict --formula Formula/dbootstrap.rb",
		"brew style Formula/dbootstrap.rb",
		"brew install --build-from-source ./Formula/dbootstrap.rb",
	} {
		if strings.Contains(strings.ToLower(content), strings.ToLower(forbidden)) {
			t.Errorf("workflow must not contain %q", forbidden)
		}
	}

	actionReference := regexp.MustCompile(`(?m)^\s*-\s+uses:\s+[^\s@]+@([0-9a-f]{40})\s*$`)
	matches := actionReference.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		t.Fatal("workflow must use actions pinned to immutable 40-character SHAs")
	}
	if len(matches) != strings.Count(content, "uses:") {
		t.Errorf("every action reference must use an immutable 40-character SHA; got %d pinned of %d", len(matches), strings.Count(content, "uses:"))
	}
}

func TestHomebrewStableChannelValidation_NativeReceiptAndFailClosedContract(t *testing.T) {
	content := readHomebrewStableChannelValidationWorkflow(t)

	for _, want := range []string{
		"https://raw.githubusercontent.com/Homebrew/install/c7952e40b7957268f61643152f4db725379b292e/install.sh",
		"HOMEBREW_INSTALL",
		"curl --fail --location --silent --show-error",
		"exit 1",
		"test -f Formula/dbootstrap.rb",
		"ruby test/homebrew_stable_channel_test.rb",
		"brew --version",
		"dbootstrap --version",
		"bootstrap.toml",
		"brew uninstall dbootstrap",
		"macOS is unsupported",
		"release asset request count: 0",
		"HOMEBREW_CACHE=\"$RUNNER_TEMP/homebrew-empty-cache\"",
		"HOMEBREW_NO_AUTO_UPDATE=1",
		"actions/upload-artifact@",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("workflow must contain %q", want)
		}
	}

	if strings.Contains(content, "runs-on: ubuntu-latest") || strings.Contains(content, "matrix") {
		t.Error("workflow must use explicit native runner jobs, not a substitutable matrix runner")
	}
	if got := strings.Count(content, "ruby test/homebrew_stable_channel_test.rb"); got != 3 {
		t.Errorf("each native job must run the formula contract; got %d executions, want 3", got)
	}
	if got := strings.Count(content, "local/pr-candidate/dbootstrap"); got != 11 {
		t.Errorf("each native job must stage and use the qualified local tap formula for audit, style, install, and Linux formula test; got %d references, want 11", got)
	}
	if got := strings.Count(content, "cp Formula/dbootstrap.rb \"$candidate_tap/Formula/dbootstrap.rb\""); got != 3 {
		t.Errorf("each native job must copy the checked-out formula into its local tap; got %d copies, want 3", got)
	}
	if got := strings.Count(content, "brew test local/pr-candidate/dbootstrap"); got != 2 {
		t.Errorf("each Linux job must run the staged formula test after installation; got %d executions, want 2", got)
	}
}
