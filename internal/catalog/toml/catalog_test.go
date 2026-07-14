package toml

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
	tomlv2 "github.com/pelletier/go-toml/v2"
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

func TestDecodePreservesMetadata(t *testing.T) {
	input := `
schema = "dniebles.catalog"
version = 1

[[tools]]
id = "git"
description = "Version control"
install = { provider = "apt", package = "git" }
presence = { kind = "command_exists", name = "git" }

[[packages]]
id = "ripgrep"
description = "Fast text search"
install = { provider = "brew", package = "ripgrep" }
presence = { kind = "path", name = "rg" }

[[runtimes]]
id = "go"
description = "Go toolchain"
install = { provider = "brew", package = "golang" }

[[dotfiles]]
id = "bash"
description = "Bash dotfiles"
`

	catalog, err := Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	want := planning.Catalog{
		Profiles: map[string]planning.Profile{},
		Bundles:  map[string]planning.Bundle{},
		Resources: map[planning.ResourceRef]planning.Resource{
			toolGit: {
				Ref:         toolGit,
				Description: "Version control",
				DependsOn:   []planning.ResourceRef{},
				Install:     &planning.InstallMetadata{Provider: "apt", Package: "git"},
				Presence:    &planning.PresenceMetadata{Kind: "command_exists", Name: "git"},
			},
			packageRipgrep: {
				Ref:         packageRipgrep,
				Description: "Fast text search",
				DependsOn:   []planning.ResourceRef{},
				Install:     &planning.InstallMetadata{Provider: "brew", Package: "ripgrep"},
				Presence:    &planning.PresenceMetadata{Kind: "path", Name: "rg"},
			},
			runtimeGo: {
				Ref:         runtimeGo,
				Description: "Go toolchain",
				DependsOn:   []planning.ResourceRef{},
				Install:     &planning.InstallMetadata{Provider: "brew", Package: "golang"},
			},
			dotBash: {
				Ref:         dotBash,
				Description: "Bash dotfiles",
				DependsOn:   []planning.ResourceRef{},
			},
		},
	}
	if !reflect.DeepEqual(catalog, want) {
		t.Fatalf("Decode() = %#v, want %#v", catalog, want)
	}
}

func TestDecodeOptionalDefaultProfile(t *testing.T) {
	tests := []struct {
		name        string
		defaultLine string
		wantDefault string
	}{
		{name: "missing default remains empty"},
		{name: "blank default is preserved", defaultLine: `default_profile = ""`, wantDefault: ""},
		{name: "unknown default is preserved", defaultLine: `default_profile = "ops"`, wantDefault: "ops"},
		{name: "declared default is mapped", defaultLine: `default_profile = "dev"`, wantDefault: "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			catalog, err := Decode(strings.NewReader(fmt.Sprintf(`
schema = "dniebles.catalog"
version = 1
%s

[[profiles]]
id = "dev"
`, tt.defaultLine)))
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if catalog.DefaultProfile != tt.wantDefault {
				t.Fatalf("DefaultProfile = %q, want %q", catalog.DefaultProfile, tt.wantDefault)
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
		{
			name:    "install metadata missing provider",
			input:   "[[tools]]\nid = \"git\"\ninstall = { package = \"git\" }",
			wantErr: "install metadata requires non-empty provider and package",
		},
		{
			name:    "install metadata missing package",
			input:   "[[tools]]\nid = \"git\"\ninstall = { provider = \"apt\" }",
			wantErr: "install metadata requires non-empty provider and package",
		},
		{
			name:    "install provider asdf is not supported",
			input:   "[[tools]]\nid = \"git\"\ninstall = { provider = \"asdf\", package = \"git\" }",
			wantErr: "install provider \"asdf\" is not supported",
		},
		{
			name:    "presence metadata unsupported kind",
			input:   "[[tools]]\nid = \"git\"\npresence = { kind = \"registry\", name = \"git\" }",
			wantErr: "presence metadata has unsupported kind",
		},
		{
			name:    "presence metadata missing name",
			input:   "[[tools]]\nid = \"git\"\npresence = { kind = \"path\" }",
			wantErr: "presence metadata requires non-empty kind and name",
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

func TestDefaultCatalogIntegrityUsesRawDeclarations(t *testing.T) {
	raw := loadRawCatalog(t, "../../../catalog/bootstrap.toml")
	if err := validateRawCatalog(raw); err != nil {
		t.Fatalf("validate raw default catalog: %v", err)
	}

	catalog, err := LoadFile("../../../catalog/bootstrap.toml")
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	assertDecodedCatalogMatchesRaw(t, catalog, raw)
	assertProfilePlansMatchRawClosure(t, catalog, raw)
}

func TestDefaultCatalogHomebrewTargets(t *testing.T) {
	catalog, err := LoadFile("../../../catalog/bootstrap.toml")
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	targets := map[planning.ResourceRef]struct {
		packageName string
		command     string
	}{
		{Kind: planning.ResourceKindPackage, Name: "nvim"}:     {packageName: "neovim", command: "nvim"},
		{Kind: planning.ResourceKindTool, Name: "git"}:         {packageName: "git", command: "git"},
		{Kind: planning.ResourceKindPackage, Name: "tealdeer"}: {packageName: "tealdeer", command: "tldr"},
		{Kind: planning.ResourceKindPackage, Name: "zsh"}:      {packageName: "zsh", command: "zsh"},
		{Kind: planning.ResourceKindPackage, Name: "bat"}:      {packageName: "bat", command: "bat"},
		{Kind: planning.ResourceKindPackage, Name: "eza"}:      {packageName: "eza", command: "eza"},
		{Kind: planning.ResourceKindPackage, Name: "fd"}:       {packageName: "fd", command: "fd"},
		{Kind: planning.ResourceKindPackage, Name: "fzf"}:      {packageName: "fzf", command: "fzf"},
		{Kind: planning.ResourceKindPackage, Name: "gh"}:       {packageName: "gh", command: "gh"},
		{Kind: planning.ResourceKindPackage, Name: "starship"}: {packageName: "starship", command: "starship"},
		{Kind: planning.ResourceKindPackage, Name: "zoxide"}:   {packageName: "zoxide", command: "zoxide"},
	}

	dev := planning.BuildPlan(catalog, planning.PlanRequest{Profile: "dev"}, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"}, planning.ConfigState{PresentKeys: map[string]bool{"go.env": true}}, planning.InstallationState{})
	planned := map[planning.ResourceRef]bool{}
	for _, step := range dev.Plan.Steps {
		planned[step.Ref] = true
	}

	for ref, want := range targets {
		resource, ok := catalog.Resources[ref]
		if !ok {
			t.Errorf("default catalog missing %s", ref)
			continue
		}
		if resource.Install == nil || resource.Install.Provider != "brew" || resource.Install.Package != want.packageName {
			t.Errorf("%s install = %#v, want brew package %q", ref, resource.Install, want.packageName)
		}
		if resource.Presence == nil || resource.Presence.Kind != "command_exists" || resource.Presence.Name != want.command {
			t.Errorf("%s presence = %#v, want command_exists %q", ref, resource.Presence, want.command)
		}
		if !planned[ref] {
			t.Errorf("dev profile plan does not include %s", ref)
		}
	}
}

func TestRawCatalogRejectsOrphanedBrewResourceEvenWhenPointSelectable(t *testing.T) {
	raw := rawCatalog{
		Tools:    []rawResource{{ID: "workflow", Install: rawInstall{Provider: "brew", Package: "workflow"}, Presence: rawPresence{Kind: "command_exists", Name: "workflow"}}, {ID: "orphan", Install: rawInstall{Provider: "brew", Package: "orphan"}, Presence: rawPresence{Kind: "command_exists", Name: "orphan"}}},
		Profiles: []rawProfile{{ID: "dev", Resources: []string{"tool:workflow"}}},
	}
	if err := validateRawCatalog(raw); err == nil || !strings.Contains(err.Error(), "orphan") {
		t.Fatalf("validateRawCatalog() error = %v, want orphan rejection", err)
	}

	catalog, err := Decode(strings.NewReader(`
[[tools]]
id = "workflow"

[[tools]]
id = "orphan"

[[profiles]]
id = "dev"
resources = ["tool:workflow"]
`))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	result := planning.BuildPlan(catalog, planning.PlanRequest{Resources: []planning.ResourceRef{{Kind: planning.ResourceKindTool, Name: "orphan"}}}, planning.EnvironmentFacts{}, planning.ConfigState{}, planning.InstallationState{})
	if got := refsFromSteps(result.Plan.Steps); !reflect.DeepEqual(got, []planning.ResourceRef{{Kind: planning.ResourceKindTool, Name: "orphan"}}) {
		t.Fatalf("point selection = %#v, want orphan selectable independently", got)
	}
}

type rawCatalog struct {
	Tools    []rawResource `toml:"tools"`
	Runtimes []rawResource `toml:"runtimes"`
	Packages []rawResource `toml:"packages"`
	Dotfiles []rawResource `toml:"dotfiles"`
	Bundles  []rawBundle   `toml:"bundles"`
	Profiles []rawProfile  `toml:"profiles"`
}

type rawResource struct {
	ID        string      `toml:"id"`
	DependsOn []string    `toml:"depends_on"`
	Install   rawInstall  `toml:"install"`
	Presence  rawPresence `toml:"presence"`
}

type rawInstall struct {
	Provider string `toml:"provider"`
	Package  string `toml:"package"`
}

type rawPresence struct {
	Kind string `toml:"kind"`
	Name string `toml:"name"`
}

type rawBundle struct {
	ID        string   `toml:"id"`
	Resources []string `toml:"resources"`
}

type rawProfile struct {
	ID        string   `toml:"id"`
	Bundles   []string `toml:"bundles"`
	Resources []string `toml:"resources"`
}

func loadRawCatalog(t *testing.T, path string) rawCatalog {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read raw catalog: %v", err)
	}
	var raw rawCatalog
	if err := tomlv2.Unmarshal(contents, &raw); err != nil {
		t.Fatalf("decode raw catalog: %v", err)
	}
	return raw
}

func validateRawCatalog(raw rawCatalog) error {
	resources, err := raw.resources()
	if err != nil {
		return err
	}
	bundles := map[string]rawBundle{}
	for _, bundle := range raw.Bundles {
		if bundle.ID == "" {
			return fmt.Errorf("bundle has empty id")
		}
		if _, duplicate := bundles[bundle.ID]; duplicate {
			return fmt.Errorf("duplicate bundle %q", bundle.ID)
		}
		bundles[bundle.ID] = bundle
		for _, value := range bundle.Resources {
			ref, err := rawRef(value)
			if err != nil {
				return fmt.Errorf("bundle %q: %w", bundle.ID, err)
			}
			if _, ok := resources[ref]; !ok {
				return fmt.Errorf("bundle %q references unknown resource %q", bundle.ID, value)
			}
		}
	}

	for ref, resource := range resources {
		for _, value := range resource.DependsOn {
			dependency, err := rawRef(value)
			if err != nil {
				return fmt.Errorf("resource %s: %w", ref, err)
			}
			if _, ok := resources[dependency]; !ok {
				return fmt.Errorf("resource %s references unknown dependency %q", ref, value)
			}
		}
		if resource.Install.Provider != "brew" {
			continue
		}
		if resource.Install.Package == "" {
			return fmt.Errorf("brew resource %s has empty package", ref)
		}
		if resource.Presence.Kind != "command_exists" || resource.Presence.Name == "" {
			return fmt.Errorf("brew resource %s has invalid presence metadata", ref)
		}
	}

	workflow := map[planning.ResourceRef]bool{}
	for _, profile := range raw.Profiles {
		closure, err := rawProfileClosure(profile, resources, bundles)
		if err != nil {
			return err
		}
		for ref := range closure {
			workflow[ref] = true
		}
	}
	for ref, resource := range resources {
		if resource.Install.Provider == "brew" && !workflow[ref] {
			return fmt.Errorf("brew resource %s is orphaned from declared profile roots", ref)
		}
	}
	return nil
}

func (raw rawCatalog) resources() (map[planning.ResourceRef]rawResource, error) {
	resources := map[planning.ResourceRef]rawResource{}
	for _, section := range []struct {
		kind      planning.ResourceKind
		resources []rawResource
	}{
		{planning.ResourceKindTool, raw.Tools},
		{planning.ResourceKindRuntime, raw.Runtimes},
		{planning.ResourceKindPackage, raw.Packages},
		{planning.ResourceKindDotfile, raw.Dotfiles},
	} {
		for _, resource := range section.resources {
			ref := planning.ResourceRef{Kind: section.kind, Name: resource.ID}
			if resource.ID == "" {
				return nil, fmt.Errorf("resource has empty id")
			}
			if _, duplicate := resources[ref]; duplicate {
				return nil, fmt.Errorf("duplicate resource %s", ref)
			}
			resources[ref] = resource
		}
	}
	return resources, nil
}

func rawProfileClosure(profile rawProfile, resources map[planning.ResourceRef]rawResource, bundles map[string]rawBundle) (map[planning.ResourceRef]bool, error) {
	closure := map[planning.ResourceRef]bool{}
	var visit func(planning.ResourceRef) error
	visit = func(ref planning.ResourceRef) error {
		if closure[ref] {
			return nil
		}
		resource, ok := resources[ref]
		if !ok {
			return fmt.Errorf("profile %q references unknown resource %s", profile.ID, ref)
		}
		closure[ref] = true
		for _, value := range resource.DependsOn {
			dependency, err := rawRef(value)
			if err != nil {
				return fmt.Errorf("resource %s: %w", ref, err)
			}
			if err := visit(dependency); err != nil {
				return err
			}
		}
		return nil
	}
	for _, value := range profile.Resources {
		ref, err := rawRef(value)
		if err != nil {
			return nil, fmt.Errorf("profile %q: %w", profile.ID, err)
		}
		if err := visit(ref); err != nil {
			return nil, err
		}
	}
	for _, bundleID := range profile.Bundles {
		bundle, ok := bundles[bundleID]
		if !ok {
			return nil, fmt.Errorf("profile %q references unknown bundle %q", profile.ID, bundleID)
		}
		for _, value := range bundle.Resources {
			ref, err := rawRef(value)
			if err != nil {
				return nil, fmt.Errorf("bundle %q: %w", bundleID, err)
			}
			if err := visit(ref); err != nil {
				return nil, err
			}
		}
	}
	return closure, nil
}

func rawRef(value string) (planning.ResourceRef, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return planning.ResourceRef{}, fmt.Errorf("invalid resource ref %q", value)
	}
	kinds := map[string]planning.ResourceKind{"tool": planning.ResourceKindTool, "runtime": planning.ResourceKindRuntime, "package": planning.ResourceKindPackage, "dotfile": planning.ResourceKindDotfile}
	kind, ok := kinds[parts[0]]
	if !ok {
		return planning.ResourceRef{}, fmt.Errorf("unsupported resource kind %q", parts[0])
	}
	return planning.ResourceRef{Kind: kind, Name: parts[1]}, nil
}

func assertDecodedCatalogMatchesRaw(t *testing.T, catalog planning.Catalog, raw rawCatalog) {
	t.Helper()
	resources, err := raw.resources()
	if err != nil {
		t.Fatalf("raw resources: %v", err)
	}
	if len(catalog.Resources) != len(resources) || len(catalog.Bundles) != len(raw.Bundles) || len(catalog.Profiles) != len(raw.Profiles) {
		t.Fatalf("decoded section counts = resources:%d bundles:%d profiles:%d, raw = resources:%d bundles:%d profiles:%d", len(catalog.Resources), len(catalog.Bundles), len(catalog.Profiles), len(resources), len(raw.Bundles), len(raw.Profiles))
	}
	for ref, rawResource := range resources {
		decoded, ok := catalog.Resources[ref]
		if !ok {
			t.Fatalf("decoded catalog missing raw resource %s", ref)
		}
		if rawResource.Install.Provider == "brew" && (decoded.Install == nil || decoded.Install.Provider == "" || decoded.Install.Package == "" || decoded.Presence == nil || decoded.Presence.Kind != "command_exists" || decoded.Presence.Name == "") {
			t.Fatalf("decoded brew metadata for %s = install:%#v presence:%#v", ref, decoded.Install, decoded.Presence)
		}
	}
	for _, bundle := range raw.Bundles {
		if _, ok := catalog.Bundles[bundle.ID]; !ok {
			t.Fatalf("decoded catalog missing raw bundle %q", bundle.ID)
		}
	}
	for _, profile := range raw.Profiles {
		if _, ok := catalog.Profiles[profile.ID]; !ok {
			t.Fatalf("decoded catalog missing raw profile %q", profile.ID)
		}
	}
}

func assertProfilePlansMatchRawClosure(t *testing.T, catalog planning.Catalog, raw rawCatalog) {
	t.Helper()
	resources, err := raw.resources()
	if err != nil {
		t.Fatalf("raw resources: %v", err)
	}
	bundles := map[string]rawBundle{}
	for _, bundle := range raw.Bundles {
		bundles[bundle.ID] = bundle
	}
	for _, profile := range raw.Profiles {
		t.Run(profile.ID, func(t *testing.T) {
			closure, err := rawProfileClosure(profile, resources, bundles)
			if err != nil {
				t.Fatalf("raw profile closure: %v", err)
			}
			first := planning.BuildPlan(catalog, planning.PlanRequest{Profile: profile.ID}, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"}, planning.ConfigState{PresentKeys: map[string]bool{"go.env": true}}, planning.InstallationState{})
			second := planning.BuildPlan(catalog, planning.PlanRequest{Profile: profile.ID}, planning.EnvironmentFacts{OS: "linux", Arch: "amd64"}, planning.ConfigState{PresentKeys: map[string]bool{"go.env": true}}, planning.InstallationState{})
			if !reflect.DeepEqual(first, second) {
				t.Fatalf("BuildPlan(%q) is not deterministic", profile.ID)
			}
			steps := refsFromSteps(first.Plan.Steps)
			if !reflect.DeepEqual(sortedRefs(steps), sortedRefsFromSet(closure)) {
				t.Fatalf("plan steps = %#v, raw closure = %#v", steps, sortedRefsFromSet(closure))
			}
			positions := map[planning.ResourceRef]int{}
			for index, ref := range steps {
				positions[ref] = index
			}
			for ref := range closure {
				for _, value := range resources[ref].DependsOn {
					dependency, _ := rawRef(value)
					if positions[dependency] >= positions[ref] {
						t.Fatalf("dependency %s must precede %s in %#v", dependency, ref, steps)
					}
				}
			}
		})
	}
}

func sortedRefs(refs []planning.ResourceRef) []planning.ResourceRef {
	sorted := append([]planning.ResourceRef(nil), refs...)
	sort.Slice(sorted, func(i, j int) bool { return fmt.Sprint(sorted[i]) < fmt.Sprint(sorted[j]) })
	return sorted
}

func sortedRefsFromSet(refs map[planning.ResourceRef]bool) []planning.ResourceRef {
	result := make([]planning.ResourceRef, 0, len(refs))
	for ref := range refs {
		result = append(result, ref)
	}
	return sortedRefs(result)
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
	packageJq      = planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "jq"}
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
