package dotfiles

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

func TestDetectorDetect(t *testing.T) {
	dotBash := planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"}
	dotShell := planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "shell"}
	toolGit := planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}

	tests := []struct {
		name           string
		basePath       string
		createBaseDirs bool
		exists         PathExists
		readDir        ReadDir
		catalog        planning.Catalog
		want           planning.InstallationState
	}{
		{
			name: "missing repo returns empty state",
			exists: func(path string) bool {
				return false
			},
			readDir: func(path string) ([]os.DirEntry, error) {
				t.Fatal("ReadDir must not be called when repo is missing")
				return nil, nil
			},
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{
					dotBash: {},
				},
			},
			want: planning.InstallationState{PresentResources: map[planning.ResourceRef]bool{}},
		},
		{
			name:     "present module is reported",
			basePath: "/home/user/.dotfiles",
			exists: func(path string) bool {
				return path == "/home/user/.dotfiles"
			},
			readDir: func(path string) ([]os.DirEntry, error) {
				return []os.DirEntry{
					fakeDirEntry{name: "bash"},
					fakeDirEntry{name: "readme"},
				}, nil
			},
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{
					dotBash:  {},
					dotShell: {},
				},
			},
			want: planning.InstallationState{
				PresentResources: map[planning.ResourceRef]bool{
					dotBash: true,
				},
			},
		},
		{
			name: "missing module is absent",
			exists: func(path string) bool {
				return true
			},
			readDir: func(path string) ([]os.DirEntry, error) {
				return []os.DirEntry{
					fakeDirEntry{name: "other"},
				}, nil
			},
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{
					dotBash: {},
				},
			},
			want: planning.InstallationState{PresentResources: map[planning.ResourceRef]bool{}},
		},
		{
			name: "read error returns empty state",
			exists: func(path string) bool {
				return true
			},
			readDir: func(path string) ([]os.DirEntry, error) {
				return nil, errors.New("permission denied")
			},
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{
					dotBash: {},
				},
			},
			want: planning.InstallationState{PresentResources: map[planning.ResourceRef]bool{}},
		},
		{
			name: "ignores non-dotfile resources",
			exists: func(path string) bool {
				return true
			},
			readDir: func(path string) ([]os.DirEntry, error) {
				return []os.DirEntry{
					fakeDirEntry{name: "git"},
					fakeDirEntry{name: "bash"},
				}, nil
			},
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{
					toolGit: {},
					dotBash: {},
				},
			},
			want: planning.InstallationState{
				PresentResources: map[planning.ResourceRef]bool{
					dotBash: true,
				},
			},
		},
		{
			name:           "nil seams use real filesystem with configured base path",
			basePath:       t.TempDir(),
			createBaseDirs: true,
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{
					dotBash: {},
				},
			},
			want: planning.InstallationState{
				PresentResources: map[planning.ResourceRef]bool{
					dotBash: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basePath := tt.basePath
			if tt.createBaseDirs {
				if err := os.Mkdir(filepath.Join(basePath, "bash"), 0o700); err != nil {
					t.Fatalf("create module dir: %v", err)
				}
			}

			detector := Detector{
				BasePath: basePath,
				Exists:   tt.exists,
				ReadDir:  tt.readDir,
			}

			got := detector.Detect(tt.catalog)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Detect() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDetectUsesDefaultSeams(t *testing.T) {
	dotBash := planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"}

	base := t.TempDir()
	if err := os.Mkdir(filepath.Join(base, "bash"), 0o700); err != nil {
		t.Fatalf("create module dir: %v", err)
	}

	got := Detector{BasePath: base}.Detect(planning.Catalog{
		Resources: map[planning.ResourceRef]planning.Resource{
			dotBash: {},
		},
	})

	want := planning.InstallationState{
		PresentResources: map[planning.ResourceRef]bool{dotBash: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Detect() = %#v, want %#v", got, want)
	}
}

func TestDetectDoesNotMutateFilesystem(t *testing.T) {
	base := t.TempDir()
	if err := os.Mkdir(filepath.Join(base, "bash"), 0o700); err != nil {
		t.Fatalf("create module dir: %v", err)
	}

	before, err := os.ReadDir(base)
	if err != nil {
		t.Fatalf("read base before detection: %v", err)
	}

	_ = Detector{BasePath: base}.Detect(planning.Catalog{
		Resources: map[planning.ResourceRef]planning.Resource{
			{Kind: planning.ResourceKindDotfile, Name: "bash"}: {},
		},
	})

	after, err := os.ReadDir(base)
	if err != nil {
		t.Fatalf("read base after detection: %v", err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("filesystem mutated: before=%#v after=%#v", before, after)
	}
}

type fakeDirEntry struct {
	name string
}

func (e fakeDirEntry) Name() string               { return e.name }
func (e fakeDirEntry) IsDir() bool                { return true }
func (e fakeDirEntry) Type() os.FileMode          { return 0 }
func (e fakeDirEntry) Info() (os.FileInfo, error) { return nil, errors.New("not implemented") }
