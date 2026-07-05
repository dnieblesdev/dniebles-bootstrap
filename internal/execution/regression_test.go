package execution

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestNoApplyCommandInCLI(t *testing.T) {
	mainPath := filepath.Join("..", "..", "cmd", "dbootstrap", "main.go")
	src, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, mainPath, src, 0)
	if err != nil {
		t.Fatalf("parse main.go: %v", err)
	}

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Name.Name == "run" {
			ast.Inspect(fn, func(n ast.Node) bool {
				if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					value := strings.Trim(lit.Value, "`\"")
					if value == "apply" {
						t.Fatalf("main.go contains an apply command: %s", lit.Value)
					}
				}
				return true
			})
		}
	}
}
