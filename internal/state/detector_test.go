package state

import (
	"errors"
	"reflect"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

func TestDetectorDetect(t *testing.T) {
	toolGit := planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}
	runtimeGo := planning.ResourceRef{Kind: planning.ResourceKindRuntime, Name: "go"}
	packageRip := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "ripgrep"}
	dotShell := planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "shell"}

	tests := []struct {
		name     string
		catalog  planning.Catalog
		present  map[string]bool
		wantRefs []planning.ResourceRef
	}{
		{
			name: "marks tool and runtime refs present when lookup succeeds",
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{
					toolGit:    {Ref: toolGit},
					runtimeGo:  {Ref: runtimeGo},
					packageRip: {Ref: packageRip},
					dotShell:   {Ref: dotShell},
				},
			},
			present:  map[string]bool{"git": true, "go": true},
			wantRefs: []planning.ResourceRef{toolGit, runtimeGo},
		},
		{
			name: "ignores package and dotfile refs regardless of lookup",
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{
					packageRip: {Ref: packageRip},
					dotShell:   {Ref: dotShell},
				},
			},
			present:  map[string]bool{"ripgrep": true, "shell": true},
			wantRefs: nil,
		},
		{
			name: "empty catalog returns empty state",
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{},
			},
			present:  nil,
			wantRefs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := Detector{
				LookPath: func(name string) (string, error) {
					if tt.present[name] {
						return "/usr/bin/" + name, nil
					}
					return "", errors.New("not found")
				},
			}

			got := detector.Detect(tt.catalog)
			want := planning.InstallationState{PresentResources: map[planning.ResourceRef]bool{}}
			for _, ref := range tt.wantRefs {
				want.PresentResources[ref] = true
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("Detect() = %#v, want %#v", got, want)
			}
		})
	}
}

func TestDetectUsesDefaultLookPath(t *testing.T) {
	// This test exercises the default nil-lookup path without depending on the real host PATH.
	// It uses a non-existent executable name so exec.LookPath returns an error.
	catalog := planning.Catalog{
		Resources: map[planning.ResourceRef]planning.Resource{
			{Kind: planning.ResourceKindTool, Name: "dniebles-bootstrap-test-missing-executable"}: {},
		},
	}

	got := Detect(catalog)
	if len(got.PresentResources) != 0 {
		t.Fatalf("Detect() with default lookup for missing executable = %#v, want empty", got.PresentResources)
	}
}
