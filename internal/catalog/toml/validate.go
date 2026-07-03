package toml

import (
	"fmt"
	"strings"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

func validate(raw catalogFile) error {
	refs := map[planning.ResourceRef]bool{}

	if err := collectResourceRefs(refs, planning.ResourceKindTool, raw.Tools); err != nil {
		return err
	}
	if err := collectResourceRefs(refs, planning.ResourceKindRuntime, raw.Runtimes); err != nil {
		return err
	}
	if err := collectResourceRefs(refs, planning.ResourceKindPackage, raw.Packages); err != nil {
		return err
	}

	bundles := map[string]bool{}
	for i, entry := range raw.Bundles {
		if entry.ID == "" {
			return fmt.Errorf("bundle[%d] missing required id", i)
		}
		if bundles[entry.ID] {
			return fmt.Errorf("duplicate bundle id %q", entry.ID)
		}
		bundles[entry.ID] = true
		if err := validateRefs(fmt.Sprintf("bundle %q resources", entry.ID), entry.Resources, refs); err != nil {
			return err
		}
	}

	profiles := map[string]bool{}
	for i, entry := range raw.Profiles {
		if entry.ID == "" {
			return fmt.Errorf("profile[%d] missing required id", i)
		}
		if profiles[entry.ID] {
			return fmt.Errorf("duplicate profile id %q", entry.ID)
		}
		profiles[entry.ID] = true
		for _, bundle := range entry.Bundles {
			if bundle == "" {
				return fmt.Errorf("profile %q has empty bundle reference", entry.ID)
			}
			if !bundles[bundle] {
				return fmt.Errorf("profile %q references unknown bundle %q", entry.ID, bundle)
			}
		}
		if err := validateRefs(fmt.Sprintf("profile %q resources", entry.ID), entry.Resources, refs); err != nil {
			return err
		}
	}

	if err := validateDependencyRefs(raw.Tools, refs, planning.ResourceKindTool); err != nil {
		return err
	}
	if err := validateDependencyRefs(raw.Runtimes, refs, planning.ResourceKindRuntime); err != nil {
		return err
	}
	if err := validateDependencyRefs(raw.Packages, refs, planning.ResourceKindPackage); err != nil {
		return err
	}

	return nil
}

func collectResourceRefs(refs map[planning.ResourceRef]bool, kind planning.ResourceKind, entries []resourceEntry) error {
	for i, entry := range entries {
		if entry.ID == "" {
			return fmt.Errorf("%s[%d] missing required id", kind, i)
		}
		ref := planning.ResourceRef{Kind: kind, Name: entry.ID}
		if refs[ref] {
			return fmt.Errorf("duplicate resource id %q for kind %q", entry.ID, kind)
		}
		refs[ref] = true
	}
	return nil
}

func validateDependencyRefs(entries []resourceEntry, known map[planning.ResourceRef]bool, kind planning.ResourceKind) error {
	for _, entry := range entries {
		if err := validateRefs(fmt.Sprintf("%s %q depends_on", kind, entry.ID), entry.DependsOn, known); err != nil {
			return err
		}
	}
	return nil
}

func validateRefs(context string, values []string, known map[planning.ResourceRef]bool) error {
	for _, value := range values {
		ref, err := parseRef(value)
		if err != nil {
			return fmt.Errorf("%s has malformed ref %q: %w", context, value, err)
		}
		if !known[ref] {
			return fmt.Errorf("%s references unknown resource %q", context, value)
		}
	}
	return nil
}

func parseRef(value string) (planning.ResourceRef, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return planning.ResourceRef{}, fmt.Errorf("expected kind:name")
	}

	kind := planning.ResourceKind(parts[0])
	if !supportedKind(kind) {
		return planning.ResourceRef{}, fmt.Errorf("unsupported resource kind %q", parts[0])
	}

	return planning.ResourceRef{Kind: kind, Name: parts[1]}, nil
}

func supportedKind(kind planning.ResourceKind) bool {
	switch kind {
	case planning.ResourceKindTool, planning.ResourceKindRuntime, planning.ResourceKindPackage:
		return true
	default:
		return false
	}
}
