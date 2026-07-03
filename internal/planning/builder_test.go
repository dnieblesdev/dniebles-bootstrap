package planning

import (
	"reflect"
	"testing"
)

var (
	toolGit    = ResourceRef{Kind: ResourceKindTool, Name: "git"}
	runtimeGo  = ResourceRef{Kind: ResourceKindRuntime, Name: "go"}
	packageRip = ResourceRef{Kind: ResourceKindPackage, Name: "ripgrep"}
	dotShell   = ResourceRef{Kind: ResourceKindDotfile, Name: "shell"}
)

func TestBuildPlanExpansionOrderingAndStability(t *testing.T) {
	catalog := Catalog{
		Profiles: map[string]Profile{
			"dev": {
				Name:      "dev",
				Bundles:   []string{"cli"},
				Resources: []ResourceRef{dotShell},
			},
		},
		Bundles: map[string]Bundle{
			"cli": {Name: "cli", Resources: []ResourceRef{packageRip, runtimeGo, packageRip}},
		},
		Resources: map[ResourceRef]Resource{
			packageRip: {Ref: packageRip, DependsOn: []ResourceRef{toolGit}},
			runtimeGo:  {Ref: runtimeGo, DependsOn: []ResourceRef{toolGit}},
			toolGit:    {Ref: toolGit},
			dotShell:   {Ref: dotShell, DependsOn: []ResourceRef{toolGit}},
		},
	}

	tests := []struct {
		name      string
		request   PlanRequest
		wantSteps []ResourceRef
	}{
		{
			name:      "profile expands bundle resources and orders dependencies first",
			request:   PlanRequest{Profile: "dev"},
			wantSteps: []ResourceRef{toolGit, dotShell, packageRip, runtimeGo},
		},
		{
			name:      "point resources include dependencies only once",
			request:   PlanRequest{Resources: []ResourceRef{packageRip, runtimeGo}},
			wantSteps: []ResourceRef{toolGit, packageRip, runtimeGo},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first := BuildPlan(catalog, tt.request, EnvironmentFacts{}, ConfigState{})
			second := BuildPlan(catalog, tt.request, EnvironmentFacts{}, ConfigState{})

			if got := refsFromSteps(first.Plan.Steps); !reflect.DeepEqual(got, tt.wantSteps) {
				t.Fatalf("steps = %#v, want %#v", got, tt.wantSteps)
			}
			if !reflect.DeepEqual(first, second) {
				t.Fatalf("BuildPlan is not stable:\nfirst=%#v\nsecond=%#v", first, second)
			}
		})
	}
}

func TestBuildPlanInvalidReferencesAndMissingConfig(t *testing.T) {
	unknownTool := ResourceRef{Kind: ResourceKindTool, Name: "missing-tool"}
	catalog := Catalog{
		Profiles: map[string]Profile{
			"partial": {
				Name:      "partial",
				Bundles:   []string{"missing-bundle", "valid"},
				Resources: []ResourceRef{unknownTool},
			},
		},
		Bundles: map[string]Bundle{
			"valid": {Name: "valid", Resources: []ResourceRef{runtimeGo, packageRip}},
		},
		Resources: map[ResourceRef]Resource{
			runtimeGo:  {Ref: runtimeGo, ConfigPolicy: ConfigPolicy{RequiredKeys: []string{"go.env"}}},
			packageRip: {Ref: packageRip},
		},
	}

	result := BuildPlan(catalog, PlanRequest{Profile: "partial"}, EnvironmentFacts{}, ConfigState{})

	if got, want := refsFromSteps(result.Plan.Steps), []ResourceRef{packageRip, runtimeGo}; !reflect.DeepEqual(got, want) {
		t.Fatalf("steps = %#v, want valid resources planned despite invalid refs %#v", got, want)
	}
	assertStatus(t, result, runtimeGo, PlanStepStatusAttentionRequired)
	assertStatus(t, result, packageRip, PlanStepStatusPlanned)
	assertStatus(t, result, unknownTool, PlanStepStatusError)
	assertReasonContains(t, result, runtimeGo, "go.env")
	assertDiagnosticContains(t, result, "unknown bundle")
}

func TestBuildPlanEnvironmentFactsAreCallerSupplied(t *testing.T) {
	linuxOnly := ResourceRef{Kind: ResourceKindPackage, Name: "apt-package"}
	darwinOnly := ResourceRef{Kind: ResourceKindPackage, Name: "brew-package"}
	catalog := Catalog{
		Resources: map[ResourceRef]Resource{
			linuxOnly:  {Ref: linuxOnly, Conditions: EnvironmentConditions{OS: []string{"linux"}}},
			darwinOnly: {Ref: darwinOnly, Conditions: EnvironmentConditions{OS: []string{"darwin"}}},
		},
	}

	result := BuildPlan(catalog, PlanRequest{Resources: []ResourceRef{darwinOnly, linuxOnly}}, EnvironmentFacts{OS: "linux"}, ConfigState{})

	if got, want := refsFromSteps(result.Plan.Steps), []ResourceRef{linuxOnly}; !reflect.DeepEqual(got, want) {
		t.Fatalf("steps = %#v, want %#v", got, want)
	}
	assertStatus(t, result, linuxOnly, PlanStepStatusPlanned)
	assertStatus(t, result, darwinOnly, PlanStepStatusSkipped)
}

func TestBuildPlanIsPureDataOnly(t *testing.T) {
	catalog := Catalog{Resources: map[ResourceRef]Resource{toolGit: {Ref: toolGit}}}
	request := PlanRequest{Resources: []ResourceRef{toolGit}}
	facts := EnvironmentFacts{OS: "linux", Arch: "amd64"}
	state := ConfigState{PresentKeys: map[string]bool{"unused": true}}

	beforeCatalog := cloneCatalog(catalog)
	beforeRequest := request
	beforeFacts := facts
	beforeState := cloneConfigState(state)

	_ = BuildPlan(catalog, request, facts, state)

	if !reflect.DeepEqual(catalog, beforeCatalog) {
		t.Fatalf("catalog mutated: got %#v want %#v", catalog, beforeCatalog)
	}
	if !reflect.DeepEqual(request, beforeRequest) {
		t.Fatalf("request mutated: got %#v want %#v", request, beforeRequest)
	}
	if !reflect.DeepEqual(facts, beforeFacts) {
		t.Fatalf("facts mutated: got %#v want %#v", facts, beforeFacts)
	}
	if !reflect.DeepEqual(state, beforeState) {
		t.Fatalf("state mutated: got %#v want %#v", state, beforeState)
	}
}

func refsFromSteps(steps []PlanStep) []ResourceRef {
	refs := make([]ResourceRef, 0, len(steps))
	for _, step := range steps {
		refs = append(refs, step.Ref)
	}
	return refs
}

func assertStatus(t *testing.T, result PlanResult, ref ResourceRef, status PlanStepStatus) {
	t.Helper()
	for _, got := range result.Results {
		if got.Ref == ref {
			if got.Status != status {
				t.Fatalf("status for %#v = %q, want %q", ref, got.Status, status)
			}
			return
		}
	}
	t.Fatalf("missing result for %#v in %#v", ref, result.Results)
}

func assertReasonContains(t *testing.T, result PlanResult, ref ResourceRef, want string) {
	t.Helper()
	for _, got := range result.Results {
		if got.Ref != ref {
			continue
		}
		for _, reason := range got.Reasons {
			if containsSubstring(reason, want) {
				return
			}
		}
	}
	t.Fatalf("missing reason containing %q for %#v in %#v", want, ref, result.Results)
}

func assertDiagnosticContains(t *testing.T, result PlanResult, want string) {
	t.Helper()
	for _, got := range result.Results {
		for _, reason := range got.Reasons {
			if containsSubstring(reason, want) {
				return
			}
		}
	}
	t.Fatalf("missing diagnostic containing %q in %#v", want, result.Results)
}

func containsSubstring(value, want string) bool {
	for i := 0; i+len(want) <= len(value); i++ {
		if value[i:i+len(want)] == want {
			return true
		}
	}
	return false
}

func cloneCatalog(catalog Catalog) Catalog {
	clone := Catalog{}
	if catalog.Profiles != nil {
		clone.Profiles = map[string]Profile{}
	}
	if catalog.Bundles != nil {
		clone.Bundles = map[string]Bundle{}
	}
	if catalog.Resources != nil {
		clone.Resources = map[ResourceRef]Resource{}
	}
	for name, profile := range catalog.Profiles {
		profile.Bundles = append([]string(nil), profile.Bundles...)
		profile.Resources = append([]ResourceRef(nil), profile.Resources...)
		clone.Profiles[name] = profile
	}
	for name, bundle := range catalog.Bundles {
		bundle.Resources = append([]ResourceRef(nil), bundle.Resources...)
		clone.Bundles[name] = bundle
	}
	for ref, resource := range catalog.Resources {
		resource.DependsOn = append([]ResourceRef(nil), resource.DependsOn...)
		resource.ConfigPolicy.RequiredKeys = append([]string(nil), resource.ConfigPolicy.RequiredKeys...)
		resource.Conditions.OS = append([]string(nil), resource.Conditions.OS...)
		resource.Conditions.Arch = append([]string(nil), resource.Conditions.Arch...)
		resource.Conditions.Distro = append([]string(nil), resource.Conditions.Distro...)
		clone.Resources[ref] = resource
	}
	return clone
}

func cloneConfigState(state ConfigState) ConfigState {
	clone := ConfigState{PresentKeys: map[string]bool{}}
	for key, value := range state.PresentKeys {
		clone.PresentKeys[key] = value
	}
	return clone
}
