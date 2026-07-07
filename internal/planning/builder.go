package planning

import (
	"fmt"
	"sort"
)

// BuildPlan expands decoded domain inputs into a deterministic dependency-aware plan.
func BuildPlan(catalog Catalog, request PlanRequest, facts EnvironmentFacts, state ConfigState, installation InstallationState) PlanResult {
	b := planBuilder{
		catalog:      catalog,
		facts:        facts,
		state:        state,
		installation: installation,
		selected:     map[ResourceRef]Resource{},
		visiting:     map[string]bool{},
		visited:      map[string]bool{},
		resultFor:    map[ResourceRef]PlanStepResult{},
	}

	b.expandRequest(request)
	b.appendOrderedSteps()

	return PlanResult{Plan: Plan{Steps: b.steps}, Results: b.results()}
}

type planBuilder struct {
	catalog      Catalog
	facts        EnvironmentFacts
	state        ConfigState
	installation InstallationState

	selected  map[ResourceRef]Resource
	visiting  map[string]bool
	visited   map[string]bool
	resultFor map[ResourceRef]PlanStepResult
	steps     []PlanStep
}

func (b *planBuilder) expandRequest(request PlanRequest) {
	if request.Profile != "" {
		profile, ok := b.catalog.Profiles[request.Profile]
		if !ok {
			b.record(PlanStepResult{Status: PlanStepStatusError, Reasons: []string{fmt.Sprintf("unknown profile %q", request.Profile)}})
		} else {
			for _, bundleName := range sortedStrings(profile.Bundles) {
				b.expandBundle(bundleName)
			}
			for _, ref := range sortedRefs(profile.Resources) {
				b.includeResource(ref)
			}
		}
	}

	for _, ref := range sortedRefs(request.Resources) {
		b.includeResource(ref)
	}
}

func (b *planBuilder) expandBundle(name string) {
	if b.visiting[name] {
		b.record(PlanStepResult{Status: PlanStepStatusError, Reasons: []string{fmt.Sprintf("bundle cycle at %q", name)}})
		return
	}
	if b.visited[name] {
		return
	}

	b.visiting[name] = true
	defer delete(b.visiting, name)

	bundle, ok := b.catalog.Bundles[name]
	if !ok {
		b.record(PlanStepResult{Status: PlanStepStatusError, Reasons: []string{fmt.Sprintf("unknown bundle %q", name)}})
		return
	}

	for _, ref := range sortedRefs(bundle.Resources) {
		b.includeResource(ref)
	}
	b.visited[name] = true
}

func (b *planBuilder) includeResource(ref ResourceRef) {
	if _, ok := b.selected[ref]; ok {
		return
	}

	resource, ok := b.catalog.Resources[ref]
	if !ok {
		b.record(PlanStepResult{Ref: ref, Status: PlanStepStatusError, Reasons: []string{fmt.Sprintf("unknown resource %s", refKey(ref))}})
		return
	}
	if resource.Ref == (ResourceRef{}) {
		resource.Ref = ref
	}
	if !matchesFacts(resource.Conditions, b.facts) {
		b.record(PlanStepResult{Ref: ref, Status: PlanStepStatusSkipped, Reasons: []string{"environment facts do not match resource conditions"}})
		return
	}

	b.selected[ref] = resource
	for _, dep := range sortedRefs(resource.DependsOn) {
		b.includeResource(dep)
	}
}

func (b *planBuilder) appendOrderedSteps() {
	for _, ref := range topoOrder(b.selected) {
		resource := b.selected[ref]
		reasons := missingConfigReasons(resource.ConfigPolicy, b.state)
		status := PlanStepStatusPlanned
		if b.installation.PresentResources[ref] {
			status = PlanStepStatusAlreadyInstalled
		} else if len(reasons) > 0 {
			status = PlanStepStatusAttentionRequired
		}

		// The resource value is copied as-is so optional metadata is preserved.
		step := PlanStep{
			Ref:              ref,
			Resource:         resource,
			DependsOn:        sortedRefs(resource.DependsOn),
			AttentionReasons: reasons,
		}
		b.steps = append(b.steps, step)
		b.record(PlanStepResult{Ref: ref, Status: status, Reasons: reasons})
	}
}

func (b *planBuilder) record(result PlanStepResult) {
	if result.Ref == (ResourceRef{}) {
		b.resultFor[ResourceRef{Kind: ResourceKind("diagnostic"), Name: fmt.Sprint(len(b.resultFor))}] = result
		return
	}
	if existing, ok := b.resultFor[result.Ref]; ok && existing.Status == PlanStepStatusError {
		return
	}
	b.resultFor[result.Ref] = result
}

func (b *planBuilder) results() []PlanStepResult {
	keys := make([]ResourceRef, 0, len(b.resultFor))
	for ref := range b.resultFor {
		keys = append(keys, ref)
	}
	keys = sortedRefs(keys)

	results := make([]PlanStepResult, 0, len(keys))
	for _, ref := range keys {
		results = append(results, b.resultFor[ref])
	}
	return results
}

func topoOrder(resources map[ResourceRef]Resource) []ResourceRef {
	refs := make([]ResourceRef, 0, len(resources))
	for ref := range resources {
		refs = append(refs, ref)
	}
	refs = sortedRefs(refs)

	visited := map[ResourceRef]bool{}
	visiting := map[ResourceRef]bool{}
	ordered := make([]ResourceRef, 0, len(refs))
	var visit func(ResourceRef)
	visit = func(ref ResourceRef) {
		if visited[ref] || visiting[ref] {
			return
		}
		visiting[ref] = true
		for _, dep := range sortedRefs(resources[ref].DependsOn) {
			if _, ok := resources[dep]; ok {
				visit(dep)
			}
		}
		delete(visiting, ref)
		visited[ref] = true
		ordered = append(ordered, ref)
	}

	for _, ref := range refs {
		visit(ref)
	}
	return ordered
}

func matchesFacts(conditions EnvironmentConditions, facts EnvironmentFacts) bool {
	if len(conditions.OS) > 0 && !contains(conditions.OS, facts.OS) {
		return false
	}
	if len(conditions.Arch) > 0 && !contains(conditions.Arch, facts.Arch) {
		return false
	}
	if len(conditions.Distro) > 0 && !contains(conditions.Distro, facts.Distro) {
		return false
	}
	if conditions.WSL != nil && *conditions.WSL != facts.WSL {
		return false
	}
	return true
}

func missingConfigReasons(policy ConfigPolicy, state ConfigState) []string {
	keys := sortedStrings(policy.RequiredKeys)
	reasons := make([]string, 0, len(keys))
	for _, key := range keys {
		if !state.PresentKeys[key] {
			reasons = append(reasons, fmt.Sprintf("missing required config %q", key))
		}
	}
	return reasons
}

func sortedRefs(refs []ResourceRef) []ResourceRef {
	copyRefs := append([]ResourceRef(nil), refs...)
	sort.Slice(copyRefs, func(i, j int) bool { return refKey(copyRefs[i]) < refKey(copyRefs[j]) })
	return copyRefs
}

func sortedStrings(values []string) []string {
	copyValues := append([]string(nil), values...)
	sort.Strings(copyValues)
	return copyValues
}

func refKey(ref ResourceRef) string {
	return string(ref.Kind) + ":" + ref.Name
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
