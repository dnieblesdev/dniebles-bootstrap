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
					toolGit:    {Ref: toolGit, Presence: &planning.PresenceMetadata{Kind: "command_exists", Name: "git"}},
					runtimeGo:  {Ref: runtimeGo, Presence: &planning.PresenceMetadata{Kind: "command_exists", Name: "go"}},
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

func TestDetectorDetectUsesOnlyConfiguredEligibleCommandPresence(t *testing.T) {
	tool := planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "editor"}
	runtime := planning.ResourceRef{Kind: planning.ResourceKindRuntime, Name: "go"}
	packageRef := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "ripgrep"}
	dotfile := planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "shell"}

	catalog := planning.Catalog{Resources: map[planning.ResourceRef]planning.Resource{
		tool:       {Ref: tool, Presence: &planning.PresenceMetadata{Kind: "command_exists", Name: "vim"}},
		runtime:    {Ref: runtime, Presence: &planning.PresenceMetadata{Kind: "command_exists", Name: "go"}},
		packageRef: {Ref: packageRef, Presence: &planning.PresenceMetadata{Kind: "command_exists", Name: "rg"}},
		dotfile:    {Ref: dotfile, Presence: &planning.PresenceMetadata{Kind: "command_exists", Name: "shell"}},
		{Kind: planning.ResourceKindTool, Name: "missing"}:  {Presence: nil},
		{Kind: planning.ResourceKindRuntime, Name: "empty"}: {Presence: &planning.PresenceMetadata{Kind: "command_exists"}},
		{Kind: planning.ResourceKindTool, Name: "other"}:    {Presence: &planning.PresenceMetadata{Kind: "path", Name: "other"}},
	}}
	var lookups []string
	got := Detector{LookPath: func(name string) (string, error) {
		lookups = append(lookups, name)
		if name == "vim" || name == "go" {
			return "/usr/bin/" + name, nil
		}
		return "", errors.New("not found")
	}}.Detect(catalog)

	want := map[planning.ResourceRef]bool{tool: true, runtime: true}
	if !reflect.DeepEqual(got.PresentResources, want) {
		t.Fatalf("present resources = %#v, want %#v", got.PresentResources, want)
	}
	lookupSet := make(map[string]bool, len(lookups))
	for _, lookup := range lookups {
		lookupSet[lookup] = true
	}
	if len(lookups) != 2 || len(lookupSet) != 2 || !lookupSet["vim"] || !lookupSet["go"] {
		t.Fatalf("lookups = %#v, want each configured eligible name exactly once", lookups)
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
