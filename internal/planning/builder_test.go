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

func TestBuildPlanProfileClosureIsCompleteAndDependenciesPrecedeDependents(t *testing.T) {
	base := ResourceRef{Kind: ResourceKindTool, Name: "base"}
	feature := ResourceRef{Kind: ResourceKindPackage, Name: "feature"}
	extra := ResourceRef{Kind: ResourceKindRuntime, Name: "extra"}
	catalog := Catalog{
		Profiles: map[string]Profile{"dev": {Name: "dev", Bundles: []string{"workflow"}}},
		Bundles:  map[string]Bundle{"workflow": {Name: "workflow", Resources: []ResourceRef{feature, extra}}},
		Resources: map[ResourceRef]Resource{
			base:    {Ref: base},
			feature: {Ref: feature, DependsOn: []ResourceRef{base}},
			extra:   {Ref: extra, DependsOn: []ResourceRef{feature}},
		},
	}

	first := BuildPlan(catalog, PlanRequest{Profile: "dev"}, EnvironmentFacts{}, ConfigState{}, InstallationState{})
	second := BuildPlan(catalog, PlanRequest{Profile: "dev"}, EnvironmentFacts{}, ConfigState{}, InstallationState{})
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("BuildPlan is not deterministic:\nfirst=%#v\nsecond=%#v", first, second)
	}
	assertCompleteDependencyOrder(t, refsFromSteps(first.Plan.Steps), map[ResourceRef][]ResourceRef{feature: {base}, extra: {feature}})
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

func TestBuildPlanCarriesResultStatusOnOrderedSteps(t *testing.T) {
	result := BuildPlan(
		Catalog{Resources: map[ResourceRef]Resource{
			toolGit:    {Ref: toolGit},
			runtimeGo:  {Ref: runtimeGo, ConfigPolicy: ConfigPolicy{RequiredKeys: []string{"go.env"}}},
			packageRip: {Ref: packageRip},
		}},
		PlanRequest{Resources: []ResourceRef{toolGit, runtimeGo, packageRip}},
		EnvironmentFacts{}, ConfigState{},
		InstallationState{PresentResources: map[ResourceRef]bool{toolGit: true}},
	)
	results := map[ResourceRef]PlanStepStatus{}
	for _, item := range result.Results {
		results[item.Ref] = item.Status
	}
	for _, step := range result.Plan.Steps {
		if step.Status != results[step.Ref] {
			t.Fatalf("step %v status = %q, want result status %q", step.Ref, step.Status, results[step.Ref])
		}
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

func TestBuildPlanPreservesResourceMetadata(t *testing.T) {
	catalog := Catalog{
		Resources: map[ResourceRef]Resource{
			packageRip: {
				Ref:       packageRip,
				DependsOn: []ResourceRef{toolGit},
				Install:   &InstallMetadata{Provider: "brew", Package: "ripgrep"},
				Presence:  &PresenceMetadata{Kind: "command_exists", Name: "rg"},
			},
			runtimeGo: {
				Ref:       runtimeGo,
				DependsOn: []ResourceRef{toolGit},
				Install:   &InstallMetadata{Provider: "asdf", Package: "golang"},
			},
			toolGit: {Ref: toolGit},
		},
	}

	result := BuildPlan(
		catalog,
		PlanRequest{Resources: []ResourceRef{packageRip, runtimeGo, toolGit}},
		EnvironmentFacts{},
		ConfigState{},
		InstallationState{},
	)

	wantSteps := []ResourceRef{toolGit, packageRip, runtimeGo}
	if got := refsFromSteps(result.Plan.Steps); !reflect.DeepEqual(got, wantSteps) {
		t.Fatalf("steps = %#v, want %#v", got, wantSteps)
	}

	for _, step := range result.Plan.Steps {
		switch step.Ref {
		case packageRip:
			if got, want := step.Resource.Install, (&InstallMetadata{Provider: "brew", Package: "ripgrep"}); !reflect.DeepEqual(got, want) {
				t.Fatalf("package install metadata = %#v, want %#v", got, want)
			}
			if got, want := step.Resource.Presence, (&PresenceMetadata{Kind: "command_exists", Name: "rg"}); !reflect.DeepEqual(got, want) {
				t.Fatalf("package presence metadata = %#v, want %#v", got, want)
			}
		case runtimeGo:
			if got, want := step.Resource.Install, (&InstallMetadata{Provider: "asdf", Package: "golang"}); !reflect.DeepEqual(got, want) {
				t.Fatalf("runtime install metadata = %#v, want %#v", got, want)
			}
			if step.Resource.Presence != nil {
				t.Fatalf("runtime presence metadata = %#v, want nil", step.Resource.Presence)
			}
		case toolGit:
			if step.Resource.Install != nil || step.Resource.Presence != nil {
				t.Fatalf("tool metadata should be nil, got install=%#v presence=%#v", step.Resource.Install, step.Resource.Presence)
			}
		}
	}
}

func TestBuildPlanMetadataDoesNotAlterPlanningOutcome(t *testing.T) {
	base := Catalog{
		Resources: map[ResourceRef]Resource{
			packageRip: {Ref: packageRip, DependsOn: []ResourceRef{toolGit}},
			runtimeGo:  {Ref: runtimeGo, DependsOn: []ResourceRef{toolGit}},
			toolGit:    {Ref: toolGit},
		},
	}
	withMetadata := cloneCatalog(base)
	packageRes := withMetadata.Resources[packageRip]
	packageRes.Install = &InstallMetadata{Provider: "brew", Package: "ripgrep"}
	packageRes.Presence = &PresenceMetadata{Kind: "command_exists", Name: "rg"}
	withMetadata.Resources[packageRip] = packageRes
	runtimeRes := withMetadata.Resources[runtimeGo]
	runtimeRes.Install = &InstallMetadata{Provider: "asdf", Package: "golang"}
	withMetadata.Resources[runtimeGo] = runtimeRes

	request := PlanRequest{Resources: []ResourceRef{packageRip, runtimeGo}}
	baseResult := BuildPlan(base, request, EnvironmentFacts{}, ConfigState{}, InstallationState{})
	metaResult := BuildPlan(withMetadata, request, EnvironmentFacts{}, ConfigState{}, InstallationState{})

	if !reflect.DeepEqual(refsFromSteps(baseResult.Plan.Steps), refsFromSteps(metaResult.Plan.Steps)) {
		t.Fatalf("metadata changed step ordering:\nbase=%#v\nmeta=%#v", baseResult.Plan.Steps, metaResult.Plan.Steps)
	}
	if !reflect.DeepEqual(baseResult.Results, metaResult.Results) {
		t.Fatalf("metadata changed planning results:\nbase=%#v\nmeta=%#v", baseResult.Results, metaResult.Results)
	}
}

func refsFromSteps(steps []PlanStep) []ResourceRef {
	refs := make([]ResourceRef, 0, len(steps))
	for _, step := range steps {
		refs = append(refs, step.Ref)
	}
	return refs
}

func assertCompleteDependencyOrder(t *testing.T, steps []ResourceRef, dependencies map[ResourceRef][]ResourceRef) {
	t.Helper()
	positions := make(map[ResourceRef]int, len(steps))
	for index, ref := range steps {
		positions[ref] = index
	}
	if len(positions) != 3 {
		t.Fatalf("steps = %#v, want complete three-resource closure", steps)
	}
	for dependent, required := range dependencies {
		dependentPosition, found := positions[dependent]
		if !found {
			t.Fatalf("steps = %#v, missing dependent %s", steps, dependent)
		}
		for _, dependency := range required {
			dependencyPosition, found := positions[dependency]
			if !found || dependencyPosition >= dependentPosition {
				t.Fatalf("steps = %#v, dependency %s must precede %s", steps, dependency, dependent)
			}
		}
	}
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
		if resource.Install != nil {
			install := *resource.Install
			resource.Install = &install
		}
		if resource.Presence != nil {
			presence := *resource.Presence
			resource.Presence = &presence
		}
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
