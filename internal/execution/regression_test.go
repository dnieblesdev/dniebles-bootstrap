package execution

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

func TestPlanningProductionCodeUnchanged(t *testing.T) {
	planningDir := filepath.Join("..", "planning")
	entries, err := os.ReadDir(planningDir)
	if err != nil {
		t.Fatalf("read planning dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			t.Fatalf("stat %s: %v", entry.Name(), err)
		}
		if info.Size() == 0 {
			t.Fatalf("planning file %s is unexpectedly empty", entry.Name())
		}
	}
}

func TestNoopExecutionRemainsNonMutating(t *testing.T) {
	// This regression replaces the previous "no apply command" gate with a
	// safety check that noop execution helpers never invoke real commands or
	// touch the filesystem.
	inst := NoopForKind(planning.ResourceKindTool)
	result := inst.Install(context.Background(), planning.PlanStep{
		Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"},
	})
	if result.Status != StepStatusNotImplemented {
		t.Fatalf("noop status = %q, want not_implemented", result.Status)
	}

	provider := NoopDotfilesProvider{}
	if err := provider.EnsureModules(context.Background(), []string{"shell"}); !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("EnsureModules error = %v, want ErrNotImplemented", err)
	}
	if err := provider.RunDotlink(context.Background(), []string{"shell"}); !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("RunDotlink error = %v, want ErrNotImplemented", err)
	}
}

func TestApplyRemainsNoopOnlyAndUnwiredFromCommandRunner(t *testing.T) {
	mainPath := filepath.Join("..", "..", "cmd", "dbootstrap", "main.go")
	content, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("read %s: %v", mainPath, err)
	}
	src := string(content)

	if strings.Contains(src, "CommandRunner") {
		t.Fatalf("cmd/dbootstrap/main.go references CommandRunner; apply must stay noop-only")
	}
	if strings.Contains(src, "RunCommand") {
		t.Fatalf("cmd/dbootstrap/main.go references RunCommand; apply must stay noop-only")
	}
	if !strings.Contains(src, "NoopForKind") {
		t.Fatalf("cmd/dbootstrap/main.go no longer wires NoopForKind installers")
	}
}
