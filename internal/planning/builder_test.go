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
			first := BuildPlan(catalog, tt.request, EnvironmentFacts{}, ConfigState{}, InstallationState{})
			second := BuildPlan(catalog, tt.request, EnvironmentFacts{}, ConfigState{}, InstallationState{})

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

	result := BuildPlan(catalog, PlanRequest{Profile: "partial"}, EnvironmentFacts{}, ConfigState{}, InstallationState{})

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

	result := BuildPlan(catalog, PlanRequest{Resources: []ResourceRef{darwinOnly, linuxOnly}}, EnvironmentFacts{OS: "linux"}, ConfigState{}, InstallationState{})

	if got, want := refsFromSteps(result.Plan.Steps), []ResourceRef{linuxOnly}; !reflect.DeepEqual(got, want) {
		t.Fatalf("steps = %#v, want %#v", got, want)
	}
	assertStatus(t, result, linuxOnly, PlanStepStatusPlanned)
	assertStatus(t, result, darwinOnly, PlanStepStatusSkipped)
}

func TestBuildPlanInstallationStatePrecedence(t *testing.T) {
	catalog := Catalog{
		Resources: map[ResourceRef]Resource{
			toolGit:    {Ref: toolGit},
			runtimeGo:  {Ref: runtimeGo, ConfigPolicy: ConfigPolicy{RequiredKeys: []string{"go.env"}}},
			packageRip: {Ref: packageRip},
		},
	}

	// topoOrder sorts steps by refKey "kind:name": package < runtime < tool.
	wantRuntimeGit := []ResourceRef{runtimeGo, toolGit}
	wantMixed := []ResourceRef{packageRip, runtimeGo, toolGit}

	tests := []struct {
		name         string
		catalog      Catalog
		request      PlanRequest
		facts        EnvironmentFacts
		state        ConfigState
		installation InstallationState
		wantSteps    []ResourceRef
		assertions   func(t *testing.T, result PlanResult)
	}{
		{
			name:         "empty state preserves planned semantics",
			catalog:      catalog,
			request:      PlanRequest{Resources: []ResourceRef{toolGit, runtimeGo}},
			installation: InstallationState{},
			wantSteps:    wantRuntimeGit,
			assertions: func(t *testing.T, result PlanResult) {
				assertStatus(t, result, toolGit, PlanStepStatusPlanned)
				assertStatus(t, result, runtimeGo, PlanStepStatusAttentionRequired)
				assertReasonContains(t, result, runtimeGo, "go.env")
			},
		},
		{
			name:         "present resources become already installed",
			catalog:      catalog,
			request:      PlanRequest{Resources: []ResourceRef{toolGit, runtimeGo}},
			installation: InstallationState{PresentResources: map[ResourceRef]bool{toolGit: true, runtimeGo: true}},
			wantSteps:    wantRuntimeGit,
			assertions: func(t *testing.T, result PlanResult) {
				assertStatus(t, result, toolGit, PlanStepStatusAlreadyInstalled)
				assertStatus(t, result, runtimeGo, PlanStepStatusAlreadyInstalled)
			},
		},
		{
			name:         "already installed wins over attention required but keeps reasons",
			catalog:      catalog,
			request:      PlanRequest{Resources: []ResourceRef{runtimeGo}},
			installation: InstallationState{PresentResources: map[ResourceRef]bool{runtimeGo: true}},
			wantSteps:    []ResourceRef{runtimeGo},
			assertions: func(t *testing.T, result PlanResult) {
				assertStatus(t, result, runtimeGo, PlanStepStatusAlreadyInstalled)
				assertReasonContains(t, result, runtimeGo, "go.env")
			},
		},
		{
			name:         "mixed state leaves absent resources planned or attention required",
			catalog:      catalog,
			request:      PlanRequest{Resources: []ResourceRef{toolGit, runtimeGo, packageRip}},
			installation: InstallationState{PresentResources: map[ResourceRef]bool{toolGit: true}},
			wantSteps:    wantMixed,
			assertions: func(t *testing.T, result PlanResult) {
				assertStatus(t, result, toolGit, PlanStepStatusAlreadyInstalled)
				assertStatus(t, result, packageRip, PlanStepStatusPlanned)
				assertStatus(t, result, runtimeGo, PlanStepStatusAttentionRequired)
			},
		},
		{
			name:         "environment mismatch stays skipped despite present state",
			catalog:      Catalog{Resources: map[ResourceRef]Resource{toolGit: {Ref: toolGit, Conditions: EnvironmentConditions{OS: []string{"linux"}}}}},
			request:      PlanRequest{Resources: []ResourceRef{toolGit}},
			facts:        EnvironmentFacts{OS: "darwin"},
			installation: InstallationState{PresentResources: map[ResourceRef]bool{toolGit: true}},
			wantSteps:    []ResourceRef{},
			assertions: func(t *testing.T, result PlanResult) {
				assertStatus(t, result, toolGit, PlanStepStatusSkipped)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPlan(tt.catalog, tt.request, tt.facts, tt.state, tt.installation)

			if got := refsFromSteps(result.Plan.Steps); !reflect.DeepEqual(got, tt.wantSteps) {
				t.Fatalf("steps = %#v, want %#v", got, tt.wantSteps)
			}
			tt.assertions(t, result)
		})
	}
}

func TestBuildPlanDotfilePresenceUsesInstallationState(t *testing.T) {
	catalog := Catalog{
		Resources: map[ResourceRef]Resource{
			dotShell: {Ref: dotShell, Description: "Shell config"},
		},
	}

	tests := []struct {
		name         string
		installation InstallationState
		wantStatus   PlanStepStatus
	}{
		{
			name:         "absent dotfile is planned",
			installation: InstallationState{},
			wantStatus:   PlanStepStatusPlanned,
		},
		{
			name: "present dotfile is already installed",
			installation: InstallationState{
				PresentResources: map[ResourceRef]bool{dotShell: true},
			},
			wantStatus: PlanStepStatusAlreadyInstalled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPlan(catalog, PlanRequest{Resources: []ResourceRef{dotShell}}, EnvironmentFacts{}, ConfigState{}, tt.installation)
			assertStatus(t, result, dotShell, tt.wantStatus)
		})
	}
}

func TestBuildPlanConfigState(t *testing.T) {
	catalog := Catalog{
		Resources: map[ResourceRef]Resource{
			runtimeGo: {Ref: runtimeGo, ConfigPolicy: ConfigPolicy{RequiredKeys: []string{"go.env"}}},
		},
	}

	tests := []struct {
		name       string
		state      ConfigState
		wantStatus PlanStepStatus
		wantReason string
	}{
		{
			name:       "missing config yields attention required",
			state:      ConfigState{},
			wantStatus: PlanStepStatusAttentionRequired,
			wantReason: "go.env",
		},
		{
			name:       "present config avoids attention",
			state:      ConfigState{PresentKeys: map[string]bool{"go.env": true}},
			wantStatus: PlanStepStatusPlanned,
			wantReason: "",
		},
		{
			name:       "empty present keys map preserves attention",
			state:      ConfigState{PresentKeys: map[string]bool{}},
			wantStatus: PlanStepStatusAttentionRequired,
			wantReason: "go.env",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPlan(catalog, PlanRequest{Resources: []ResourceRef{runtimeGo}}, EnvironmentFacts{}, tt.state, InstallationState{})

			assertStatus(t, result, runtimeGo, tt.wantStatus)
			if tt.wantReason != "" {
				assertReasonContains(t, result, runtimeGo, tt.wantReason)
			}
		})
	}
}

func TestBuildPlanIsPureDataOnly(t *testing.T) {
	catalog := Catalog{Resources: map[ResourceRef]Resource{toolGit: {Ref: toolGit}}}
	request := PlanRequest{Resources: []ResourceRef{toolGit}}
	facts := EnvironmentFacts{OS: "linux", Arch: "amd64"}
	state := ConfigState{PresentKeys: map[string]bool{"unused": true}}

	installation := InstallationState{PresentResources: map[ResourceRef]bool{toolGit: true}}

	beforeCatalog := cloneCatalog(catalog)
	beforeRequest := request
	beforeFacts := facts
	beforeState := cloneConfigState(state)
	beforeInstallation := cloneInstallationState(installation)

	_ = BuildPlan(catalog, request, facts, state, installation)

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
	if !reflect.DeepEqual(installation, beforeInstallation) {
		t.Fatalf("installation state mutated: got %#v want %#v", installation, beforeInstallation)
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

func cloneInstallationState(installation InstallationState) InstallationState {
	clone := InstallationState{PresentResources: map[ResourceRef]bool{}}
	for ref, present := range installation.PresentResources {
		clone.PresentResources[ref] = present
	}
	return clone
}
