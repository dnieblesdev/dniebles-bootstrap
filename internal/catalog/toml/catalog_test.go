package toml

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

func TestDecodeValidCatalog(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  planning.Catalog
	}{
		{
			name:  "maps supported sections",
			input: validCatalogTOML,
			want: planning.Catalog{
				Profiles: map[string]planning.Profile{
					"dev": {Name: "dev", Bundles: []string{"cli"}, Resources: []planning.ResourceRef{runtimeGo}},
				},
				Bundles: map[string]planning.Bundle{
					"cli": {Name: "cli", Resources: []planning.ResourceRef{toolGit, packageRipgrep}},
				},
				Resources: map[planning.ResourceRef]planning.Resource{
					toolGit: {Ref: toolGit, Description: "Version control", DependsOn: []planning.ResourceRef{}, Conditions: planning.EnvironmentConditions{OS: []string{"linux", "darwin"}}},
					runtimeGo: {
						Ref:          runtimeGo,
						Description:  "Go toolchain",
						DependsOn:    []planning.ResourceRef{toolGit},
						ConfigPolicy: planning.ConfigPolicy{RequiredKeys: []string{"go.env"}},
						Conditions:   planning.EnvironmentConditions{OS: []string{"linux"}, Arch: []string{"amd64"}, WSL: boolPtr(false)},
					},
					packageRipgrep: {Ref: packageRipgrep, Description: "Fast text search", DependsOn: []planning.ResourceRef{toolGit}},
					dotBash:        {Ref: dotBash, Description: "Bash dotfiles", DependsOn: []planning.ResourceRef{}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decode(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Decode() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDecodeValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "invalid TOML syntax",
			input:   `[[tools]`,
			wantErr: "toml",
		},
		{
			name:    "missing required resource id",
			input:   "[[tools]]\ndescription = \"missing id\"",
			wantErr: "missing required id",
		},
		{
			name:    "duplicate resource id",
			input:   "[[tools]]\nid = \"git\"\n\n[[tools]]\nid = \"git\"",
			wantErr: "duplicate resource id",
		},
		{
			name:    "malformed ref",
			input:   "[[tools]]\nid = \"git\"\n\n[[bundles]]\nid = \"cli\"\nresources = [\"tool\"]",
			wantErr: "malformed ref",
		},
		{
			name:    "unsupported ref kind",
			input:   "[[tools]]\nid = \"git\"\n\n[[bundles]]\nid = \"cli\"\nresources = [\"script:shell\"]",
			wantErr: "unsupported resource kind",
		},
		{
			name:    "dotfile missing required id",
			input:   "[[dotfiles]]\ndescription = \"missing id\"",
			wantErr: "missing required id",
		},
		{
			name:    "duplicate dotfile id",
			input:   "[[dotfiles]]\nid = \"bash\"\n\n[[dotfiles]]\nid = \"bash\"",
			wantErr: "duplicate resource id",
		},
		{
			name:    "dotfile depends on unknown resource",
			input:   "[[dotfiles]]\nid = \"bash\"\ndepends_on = [\"tool:missing\"]",
			wantErr: "unknown resource",
		},
		{
			name:    "unknown resource ref",
			input:   "[[tools]]\nid = \"git\"\n\n[[bundles]]\nid = \"cli\"\nresources = [\"package:ripgrep\"]",
			wantErr: "unknown resource",
		},
		{
			name:    "unknown bundle ref",
			input:   "[[profiles]]\nid = \"dev\"\nbundles = [\"cli\"]",
			wantErr: "unknown bundle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decode(strings.NewReader(tt.input))
			if err == nil {
				t.Fatal("Decode() error = nil, want error")
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantErr)) {
				t.Fatalf("Decode() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestDecodeDotfileRefsAreValid(t *testing.T) {
	input := `
[[tools]]
id = "git"

[[dotfiles]]
id = "bash"
depends_on = ["tool:git"]

[[bundles]]
id = "shell"
resources = ["dotfile:bash"]

[[profiles]]
id = "dev"
resources = ["dotfile:bash"]
`
	catalog, err := Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	want := planning.Catalog{
		Profiles: map[string]planning.Profile{
			"dev": {Name: "dev", Resources: []planning.ResourceRef{dotBash}},
		},
		Bundles: map[string]planning.Bundle{
			"shell": {Name: "shell", Resources: []planning.ResourceRef{dotBash}},
		},
		Resources: map[planning.ResourceRef]planning.Resource{
			toolGit: {Ref: toolGit, DependsOn: []planning.ResourceRef{}},
			dotBash: {Ref: dotBash, DependsOn: []planning.ResourceRef{toolGit}},
		},
	}
	if !reflect.DeepEqual(catalog, want) {
		t.Fatalf("Decode() = %#v, want %#v", catalog, want)
	}
}

func TestLoadFileAndBuildPlanFromFixture(t *testing.T) {
	catalog, err := LoadFile("../../../catalog/bootstrap.toml")
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	result := planning.BuildPlan(
		catalog,
		planning.PlanRequest{Profile: "dev"},
		planning.EnvironmentFacts{OS: "linux", Arch: "amd64"},
		planning.ConfigState{PresentKeys: map[string]bool{"go.env": true}},
		planning.InstallationState{},
	)

	wantSteps := []planning.ResourceRef{toolGit, packageRipgrep, runtimeGo}
	if got := refsFromSteps(result.Plan.Steps); !reflect.DeepEqual(got, wantSteps) {
		t.Fatalf("planned steps = %#v, want %#v", got, wantSteps)
	}
	assertStatus(t, result, toolGit, planning.PlanStepStatusPlanned)
	assertStatus(t, result, packageRipgrep, planning.PlanStepStatusPlanned)
	assertStatus(t, result, runtimeGo, planning.PlanStepStatusPlanned)
}

func TestDecodeValidCatalogDelegatesSemanticIssuesToPlanner(t *testing.T) {
	catalog, err := Decode(strings.NewReader(`
[[runtimes]]
id = "go"
config_required = ["go.env"]

[[profiles]]
id = "dev"
resources = ["runtime:go"]
`))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	result := planning.BuildPlan(
		catalog,
		planning.PlanRequest{Profile: "dev"},
		planning.EnvironmentFacts{},
		planning.ConfigState{},
		planning.InstallationState{},
	)

	if got := refsFromSteps(result.Plan.Steps); !reflect.DeepEqual(got, []planning.ResourceRef{runtimeGo}) {
		t.Fatalf("planned steps = %#v, want %#v", got, []planning.ResourceRef{runtimeGo})
	}
	assertStatus(t, result, runtimeGo, planning.PlanStepStatusAttentionRequired)
	assertReasonContains(t, result, runtimeGo, "go.env")
}

func refsFromSteps(steps []planning.PlanStep) []planning.ResourceRef {
	refs := make([]planning.ResourceRef, 0, len(steps))
	for _, step := range steps {
		refs = append(refs, step.Ref)
	}
	return refs
}

func assertStatus(t *testing.T, result planning.PlanResult, ref planning.ResourceRef, want planning.PlanStepStatus) {
	t.Helper()
	for _, got := range result.Results {
		if got.Ref == ref {
			if got.Status != want {
				t.Fatalf("status for %#v = %q, want %q", ref, got.Status, want)
			}
			return
		}
	}
	t.Fatalf("missing result for %#v in %#v", ref, result.Results)
}

func assertReasonContains(t *testing.T, result planning.PlanResult, ref planning.ResourceRef, want string) {
	t.Helper()
	for _, got := range result.Results {
		if got.Ref != ref {
			continue
		}
		for _, reason := range got.Reasons {
			if strings.Contains(reason, want) {
				return
			}
		}
	}
	t.Fatalf("missing reason containing %q for %#v in %#v", want, ref, result.Results)
}

func boolPtr(value bool) *bool {
	return &value
}

var (
	toolGit        = planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}
	runtimeGo      = planning.ResourceRef{Kind: planning.ResourceKindRuntime, Name: "go"}
	packageRipgrep = planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "ripgrep"}
	dotBash        = planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"}
)

const validCatalogTOML = `
schema = "dniebles.catalog"
version = 1

[[tools]]
id = "git"
description = "Version control"
os = ["linux", "darwin"]

[[runtimes]]
id = "go"
description = "Go toolchain"
depends_on = ["tool:git"]
config_required = ["go.env"]
os = ["linux"]
arch = ["amd64"]
wsl = false

[[packages]]
id = "ripgrep"
description = "Fast text search"
depends_on = ["tool:git"]

[[dotfiles]]
id = "bash"
description = "Bash dotfiles"

[[bundles]]
id = "cli"
resources = ["tool:git", "package:ripgrep"]

[[profiles]]
id = "dev"
bundles = ["cli"]
resources = ["runtime:go"]
`
