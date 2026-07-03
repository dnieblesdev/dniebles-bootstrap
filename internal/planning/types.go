package planning

// Catalog is the decoded, format-agnostic planning input.
type Catalog struct {
	Profiles  map[string]Profile
	Bundles   map[string]Bundle
	Resources map[ResourceRef]Resource
}

// Profile describes a named installation scope.
type Profile struct {
	Name      string
	Bundles   []string
	Resources []ResourceRef
}

// Bundle describes a reusable group of resources.
type Bundle struct {
	Name      string
	Resources []ResourceRef
}

// ResourceKind classifies installable resources without coupling to adapters.
type ResourceKind string

const (
	ResourceKindTool    ResourceKind = "tool"
	ResourceKindRuntime ResourceKind = "runtime"
	ResourceKindPackage ResourceKind = "package"
	ResourceKindDotfile ResourceKind = "dotfile"
)

// ResourceRef is a stable typed reference to an installable resource.
type ResourceRef struct {
	Kind ResourceKind
	Name string
}

// Resource is desired-state data only. It never performs installation.
type Resource struct {
	Ref          ResourceRef
	Description  string
	DependsOn    []ResourceRef
	ConfigPolicy ConfigPolicy
	Conditions   EnvironmentConditions
}

// ConfigPolicy declares configuration that must be visible in planning results.
type ConfigPolicy struct {
	RequiredKeys []string
}

// ConfigState is caller-supplied configuration presence data.
type ConfigState struct {
	PresentKeys map[string]bool
}

// EnvironmentFacts are caller-supplied facts. The planning core never probes the OS.
type EnvironmentFacts struct {
	OS     string
	Arch   string
	Distro string
	WSL    bool
}

// EnvironmentConditions optionally restrict a resource to matching facts.
type EnvironmentConditions struct {
	OS     []string
	Arch   []string
	Distro []string
	WSL    *bool
}

// PlanRequest selects a profile and/or point resources to plan.
type PlanRequest struct {
	Profile   string
	Resources []ResourceRef
}

// Plan contains ordered desired work data only.
type Plan struct {
	Steps []PlanStep
}

// PlanStep describes intended planning work for one resource.
type PlanStep struct {
	Ref              ResourceRef
	Resource         Resource
	DependsOn        []ResourceRef
	AttentionReasons []string
}

// PlanResult returns the plan plus structured planning-time statuses.
type PlanResult struct {
	Plan    Plan
	Results []PlanStepResult
}

// PlanStepStatus is a structured planning-time outcome.
type PlanStepStatus string

const (
	PlanStepStatusPlanned           PlanStepStatus = "planned"
	PlanStepStatusSkipped           PlanStepStatus = "skipped"
	PlanStepStatusAttentionRequired PlanStepStatus = "attention_required"
	PlanStepStatusError             PlanStepStatus = "error"
)

// PlanStepResult describes planning outcomes without requiring text parsing.
type PlanStepResult struct {
	Ref     ResourceRef
	Status  PlanStepStatus
	Reasons []string
}
