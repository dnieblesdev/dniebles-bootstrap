package config

import "testing"

// TestTemporaryCIFailureEvidence is intentionally failing and temporary.
// It exists only to prove that the GitHub Actions pull_request workflow
// stops at the failed go test ./... step and skips go vet ./... and
// go build ./... . This file and the branch that contains it must NOT
// be merged into main. See issue #1 and the ci-build-validation
// verification report for the collected evidence.
func TestTemporaryCIFailureEvidence(t *testing.T) {
	t.Fatalf("temporary CI failure evidence: this test intentionally fails to validate workflow step propagation")
}
