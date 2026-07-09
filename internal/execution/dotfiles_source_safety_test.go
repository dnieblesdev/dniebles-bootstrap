package execution

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDotfilesExecutionCoreSourceSafety(t *testing.T) {
	files := []string{
		"dotfiles_base.go",
		"dotfiles_provider.go",
		"dotfiles_installer.go",
	}
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}
			src := string(content)
			for _, forbidden := range []string{"exec.Command", "exec.CommandContext", "clone", "pull", "submodule", "fetch", "remote"} {
				if strings.Contains(src, forbidden) {
					t.Fatalf("%s contains forbidden token %q", file, forbidden)
				}
			}
		})
	}
}

func TestInternalDotfilesStaysReadOnly(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("..", "dotfiles"))
	if err != nil {
		t.Fatalf("read internal/dotfiles: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		path := filepath.Join("..", "dotfiles", entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		src := uncommentedSource(string(content))
		for _, forbidden := range []string{"RunCommand", "CommandRequest", "dotlink", "exec.Command", "clone", "pull", "submodule", "fetch"} {
			if strings.Contains(src, forbidden) {
				t.Fatalf("%s contains execution or acquisition token %q", path, forbidden)
			}
		}
	}
}

func uncommentedSource(src string) string {
	var b strings.Builder
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}
