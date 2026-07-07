package toml

import (
	"io"
	"os"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
	gotoml "github.com/pelletier/go-toml/v2"
)

// LoadFile decodes a repository-local TOML catalog file into planning inputs.
func LoadFile(path string) (planning.Catalog, error) {
	file, err := os.Open(path)
	if err != nil {
		return planning.Catalog{}, err
	}
	defer file.Close()

	return Decode(file)
}

// Decode decodes TOML catalog data into the format-agnostic planning domain.
func Decode(r io.Reader) (planning.Catalog, error) {
	var raw catalogFile
	if err := gotoml.NewDecoder(r).Decode(&raw); err != nil {
		return planning.Catalog{}, err
	}

	if err := validate(raw); err != nil {
		return planning.Catalog{}, err
	}

	return mapCatalog(raw), nil
}

func mapCatalog(raw catalogFile) planning.Catalog {
	catalog := planning.Catalog{
		Profiles:  make(map[string]planning.Profile, len(raw.Profiles)),
		Bundles:   make(map[string]planning.Bundle, len(raw.Bundles)),
		Resources: make(map[planning.ResourceRef]planning.Resource, len(raw.Tools)+len(raw.Runtimes)+len(raw.Packages)+len(raw.Dotfiles)),
	}

	for _, entry := range raw.Profiles {
		catalog.Profiles[entry.ID] = planning.Profile{
			Name:      entry.ID,
			Bundles:   append([]string(nil), entry.Bundles...),
			Resources: mustParseRefs(entry.Resources),
		}
	}

	for _, entry := range raw.Bundles {
		catalog.Bundles[entry.ID] = planning.Bundle{
			Name:      entry.ID,
			Resources: mustParseRefs(entry.Resources),
		}
	}

	mapResources(catalog.Resources, planning.ResourceKindTool, raw.Tools)
	mapResources(catalog.Resources, planning.ResourceKindRuntime, raw.Runtimes)
	mapResources(catalog.Resources, planning.ResourceKindPackage, raw.Packages)
	mapResources(catalog.Resources, planning.ResourceKindDotfile, raw.Dotfiles)

	return catalog
}

func mapResources(resources map[planning.ResourceRef]planning.Resource, kind planning.ResourceKind, entries []resourceEntry) {
	for _, entry := range entries {
		ref := planning.ResourceRef{Kind: kind, Name: entry.ID}
		resources[ref] = planning.Resource{
			Ref:         ref,
			Description: entry.Description,
			DependsOn:   mustParseRefs(entry.DependsOn),
			ConfigPolicy: planning.ConfigPolicy{
				RequiredKeys: append([]string(nil), entry.ConfigRequired...),
			},
			Conditions: planning.EnvironmentConditions{
				OS:     append([]string(nil), entry.OS...),
				Arch:   append([]string(nil), entry.Arch...),
				Distro: append([]string(nil), entry.Distro...),
				WSL:    cloneBool(entry.WSL),
			},
			Install:  cloneInstallMetadata(entry.Install),
			Presence: clonePresenceMetadata(entry.Presence),
		}
	}
}

func mustParseRefs(values []string) []planning.ResourceRef {
	refs := make([]planning.ResourceRef, 0, len(values))
	for _, value := range values {
		ref, err := parseRef(value)
		if err != nil {
			panic(err)
		}
		refs = append(refs, ref)
	}
	return refs
}

func cloneBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func cloneInstallMetadata(entry *installEntry) *planning.InstallMetadata {
	if entry == nil {
		return nil
	}
	return &planning.InstallMetadata{
		Provider: entry.Provider,
		Package:  entry.Package,
	}
}

func clonePresenceMetadata(entry *presenceEntry) *planning.PresenceMetadata {
	if entry == nil {
		return nil
	}
	return &planning.PresenceMetadata{
		Kind: entry.Kind,
		Name: entry.Name,
	}
}
