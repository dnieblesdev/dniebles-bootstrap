package config

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

func TestDetectorDetect(t *testing.T) {
	toolGit := planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}
	runtimeGo := planning.ResourceRef{Kind: planning.ResourceKindRuntime, Name: "go"}

	tests := []struct {
		name     string
		catalog  planning.Catalog
		base     string
		exists   map[string]bool
		resolver KeyPathResolver
		wantKeys []string
	}{
		{
			name: "marks required keys present when path exists",
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{
					runtimeGo: {Ref: runtimeGo, ConfigPolicy: planning.ConfigPolicy{RequiredKeys: []string{"go.env"}}},
				},
			},
			base:     "/home/test/.dotfiles/config",
			exists:   map[string]bool{"/home/test/.dotfiles/config/go/env": true},
			wantKeys: []string{"go.env"},
		},
		{
			name: "reports key absent when path is missing",
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{
					runtimeGo: {Ref: runtimeGo, ConfigPolicy: planning.ConfigPolicy{RequiredKeys: []string{"go.env"}}},
				},
			},
			base:     "/home/test/.dotfiles/config",
			exists:   map[string]bool{},
			wantKeys: nil,
		},
		{
			name: "collects keys from multiple resources without duplicates",
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{
					runtimeGo: {Ref: runtimeGo, ConfigPolicy: planning.ConfigPolicy{RequiredKeys: []string{"go.env", "shared.key"}}},
					toolGit:   {Ref: toolGit, ConfigPolicy: planning.ConfigPolicy{RequiredKeys: []string{"shared.key"}}},
				},
			},
			base:     "/home/test/.dotfiles/config",
			exists:   map[string]bool{"/home/test/.dotfiles/config/go/env": true, "/home/test/.dotfiles/config/shared/key": true},
			wantKeys: []string{"go.env", "shared.key"},
		},
		{
			name: "treats empty key as absent",
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{
					runtimeGo: {Ref: runtimeGo, ConfigPolicy: planning.ConfigPolicy{RequiredKeys: []string{""}}},
				},
			},
			base:     "/home/test/.dotfiles/config",
			exists:   map[string]bool{"/home/test/.dotfiles/config": true},
			wantKeys: nil,
		},
		{
			name: "treats absolute key as absent",
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{
					runtimeGo: {Ref: runtimeGo, ConfigPolicy: planning.ConfigPolicy{RequiredKeys: []string{"/etc/passwd"}}},
				},
			},
			base:     "/home/test/.dotfiles/config",
			exists:   map[string]bool{"/etc/passwd": true},
			wantKeys: nil,
		},
		{
			name: "treats dot-dot escaping key as absent",
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{
					runtimeGo: {Ref: runtimeGo, ConfigPolicy: planning.ConfigPolicy{RequiredKeys: []string{"go..env"}}},
				},
			},
			base:     "/home/test/.dotfiles/config",
			exists:   map[string]bool{"/home/test/.dotfiles/config/go/../env": true},
			wantKeys: nil,
		},
		{
			name: "empty catalog returns empty state",
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{},
			},
			base:     "/home/test/.dotfiles/config",
			exists:   map[string]bool{},
			wantKeys: nil,
		},
		{
			name: "custom resolver overrides default mapping",
			catalog: planning.Catalog{
				Resources: map[planning.ResourceRef]planning.Resource{
					runtimeGo: {Ref: runtimeGo, ConfigPolicy: planning.ConfigPolicy{RequiredKeys: []string{"go.env"}}},
				},
			},
			base: "/ignored",
			resolver: func(basePath, key string) (string, bool) {
				return "/custom/" + key, key == "go.env"
			},
			exists:   map[string]bool{"/custom/go.env": true},
			wantKeys: []string{"go.env"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := Detector{
				BasePath: tt.base,
				Exists: func(path string) bool {
					return tt.exists[path]
				},
				PathForKey: tt.resolver,
			}

			got := detector.Detect(tt.catalog)
			want := planning.ConfigState{PresentKeys: map[string]bool{}}
			for _, key := range tt.wantKeys {
				want.PresentKeys[key] = true
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("Detect() = %#v, want %#v", got, want)
			}
		})
	}
}

func TestDetectIsDeterministic(t *testing.T) {
	catalog := planning.Catalog{
		Resources: map[planning.ResourceRef]planning.Resource{
			{Kind: planning.ResourceKindRuntime, Name: "go"}: {
				ConfigPolicy: planning.ConfigPolicy{RequiredKeys: []string{"go.env"}},
			},
		},
	}

	detector := Detector{
		BasePath: "/home/test/.dotfiles/config",
		Exists:   func(path string) bool { return path == "/home/test/.dotfiles/config/go/env" },
	}

	first := detector.Detect(catalog)
	second := detector.Detect(catalog)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("Detect is not deterministic:\nfirst=%#v\nsecond=%#v", first, second)
	}
}

func TestDetectUsesDefaultExists(t *testing.T) {
	// This test exercises the default nil-seam path without depending on the real host dotfiles layout.
	// It uses a key whose conventional path is extremely unlikely to exist.
	catalog := planning.Catalog{
		Resources: map[planning.ResourceRef]planning.Resource{
			{Kind: planning.ResourceKindRuntime, Name: "dniebles-bootstrap-test-missing-config"}: {
				ConfigPolicy: planning.ConfigPolicy{RequiredKeys: []string{"dniebles.bootstrap.test.missing.config"}},
			},
		},
	}

	got := Detect(catalog)
	if len(got.PresentKeys) != 0 {
		t.Fatalf("Detect() with default exists for missing config = %#v, want empty", got.PresentKeys)
	}
}

func TestDefaultKeyPathResolver(t *testing.T) {
	base := "/home/test/.dotfiles/config"

	tests := []struct {
		name   string
		key    string
		want   string
		wantOK bool
	}{
		{name: "splits key on dots", key: "go.env", want: filepath.Join(base, "go", "env"), wantOK: true},
		{name: "single segment", key: "git", want: filepath.Join(base, "git"), wantOK: true},
		{name: "multiple segments", key: "a.b.c", want: filepath.Join(base, "a", "b", "c"), wantOK: true},
		{name: "empty key rejected", key: "", want: "", wantOK: false},
		{name: "absolute key rejected", key: "/etc/passwd", want: "", wantOK: false},
		{name: "dot-dot segment rejected", key: "go..env", want: "", wantOK: false},
		{name: "leading dot rejected", key: ".go.env", want: "", wantOK: false},
		{name: "trailing dot rejected", key: "go.env.", want: "", wantOK: false},
		{name: "path separator in segment rejected", key: "go/env", want: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := defaultKeyPathResolver(base, tt.key)
			if ok != tt.wantOK || got != tt.want {
				t.Fatalf("defaultKeyPathResolver(%q) = (%q, %t), want (%q, %t)", tt.key, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestDetectDoesNotMutateCatalog(t *testing.T) {
	catalog := planning.Catalog{
		Resources: map[planning.ResourceRef]planning.Resource{
			{Kind: planning.ResourceKindRuntime, Name: "go"}: {
				ConfigPolicy: planning.ConfigPolicy{RequiredKeys: []string{"go.env"}},
			},
		},
	}

	before := cloneCatalog(catalog)
	_ = Detector{
		BasePath: "/home/test/.dotfiles/config",
		Exists:   func(path string) bool { return true },
	}.Detect(catalog)

	if !reflect.DeepEqual(catalog, before) {
		t.Fatalf("catalog mutated: got %#v want %#v", catalog, before)
	}
}

func TestDetectExistenceErrorTreatedAsAbsent(t *testing.T) {
	catalog := planning.Catalog{
		Resources: map[planning.ResourceRef]planning.Resource{
			{Kind: planning.ResourceKindRuntime, Name: "go"}: {
				ConfigPolicy: planning.ConfigPolicy{RequiredKeys: []string{"go.env"}},
			},
		},
	}

	detector := Detector{
		BasePath: "/home/test/.dotfiles/config",
		Exists:   func(path string) bool { return false },
	}

	got := detector.Detect(catalog)
	if len(got.PresentKeys) != 0 {
		t.Fatalf("Detect() with failing exists = %#v, want empty", got.PresentKeys)
	}
}

func cloneCatalog(catalog planning.Catalog) planning.Catalog {
	clone := planning.Catalog{}
	if catalog.Profiles != nil {
		clone.Profiles = map[string]planning.Profile{}
	}
	if catalog.Bundles != nil {
		clone.Bundles = map[string]planning.Bundle{}
	}
	if catalog.Resources != nil {
		clone.Resources = map[planning.ResourceRef]planning.Resource{}
	}
	for name, profile := range catalog.Profiles {
		profile.Bundles = append([]string(nil), profile.Bundles...)
		profile.Resources = append([]planning.ResourceRef(nil), profile.Resources...)
		clone.Profiles[name] = profile
	}
	for name, bundle := range catalog.Bundles {
		bundle.Resources = append([]planning.ResourceRef(nil), bundle.Resources...)
		clone.Bundles[name] = bundle
	}
	for ref, resource := range catalog.Resources {
		resource.DependsOn = append([]planning.ResourceRef(nil), resource.DependsOn...)
		resource.ConfigPolicy.RequiredKeys = append([]string(nil), resource.ConfigPolicy.RequiredKeys...)
		resource.Conditions.OS = append([]string(nil), resource.Conditions.OS...)
		resource.Conditions.Arch = append([]string(nil), resource.Conditions.Arch...)
		resource.Conditions.Distro = append([]string(nil), resource.Conditions.Distro...)
		clone.Resources[ref] = resource
	}
	return clone
}
