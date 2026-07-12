package state

import (
	"os/exec"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

// PathLookup resolves an executable name to its filesystem path.
type PathLookup func(name string) (string, error)

// Detector inspects a catalog and reports which resources appear to already be installed.
type Detector struct {
	LookPath PathLookup
}

// Detect returns installation state for the given catalog using the default lookup.
func Detect(catalog planning.Catalog) planning.InstallationState {
	return Detector{}.Detect(catalog)
}

// Detect returns installation state for the given catalog using the detector's lookup seam.
func (d Detector) Detect(catalog planning.Catalog) planning.InstallationState {
	lookup := d.LookPath
	if lookup == nil {
		lookup = exec.LookPath
	}

	present := map[planning.ResourceRef]bool{}
	for ref, resource := range catalog.Resources {
		if !isDetectableKind(ref.Kind) || resource.Presence == nil ||
			resource.Presence.Kind != "command_exists" || resource.Presence.Name == "" {
			continue
		}
		if _, err := lookup(resource.Presence.Name); err == nil {
			present[ref] = true
		}
	}

	return planning.InstallationState{PresentResources: present}
}

func isDetectableKind(kind planning.ResourceKind) bool {
	return kind == planning.ResourceKindTool || kind == planning.ResourceKindRuntime
}
