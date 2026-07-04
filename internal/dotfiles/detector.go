// Package dotfiles provides read-only detection of local dotfile module availability.
//
// It never clones, applies, installs, or mutates dotfiles. A module is considered
// available only when its directory exists under the configured dotfiles base path.
package dotfiles

import (
	"os"
	"path/filepath"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

// PathExists reports whether a path is present on the filesystem.
type PathExists func(path string) bool

// ReadDir returns the entries in a directory.
type ReadDir func(path string) ([]os.DirEntry, error)

// Detector inspects a catalog and reports which dotfile modules are locally available.
type Detector struct {
	// BasePath is the directory that contains dotfile module directories.
	// When empty, it defaults to $HOME/.dotfiles.
	BasePath string
	// Exists reports whether a path is present. When nil, os.Stat is used.
	Exists PathExists
	// ReadDir lists directory entries. When nil, os.ReadDir is used.
	ReadDir ReadDir
}

// Detect returns installation state for the given catalog using the default filesystem seams.
func Detect(catalog planning.Catalog) planning.InstallationState {
	return Detector{}.Detect(catalog)
}

// Detect returns installation state for the given catalog using the detector's seams.
func (d Detector) Detect(catalog planning.Catalog) planning.InstallationState {
	base := d.basePath()
	exists := d.existsFunc()
	readDir := d.readDirFunc()

	if !exists(base) {
		return planning.InstallationState{PresentResources: map[planning.ResourceRef]bool{}}
	}

	entries, err := readDir(base)
	if err != nil {
		return planning.InstallationState{PresentResources: map[planning.ResourceRef]bool{}}
	}

	modules := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			modules[entry.Name()] = true
		}
	}

	present := map[planning.ResourceRef]bool{}
	for ref := range catalog.Resources {
		if ref.Kind != planning.ResourceKindDotfile {
			continue
		}
		if modules[ref.Name] {
			present[ref] = true
		}
	}

	return planning.InstallationState{PresentResources: present}
}

func (d Detector) basePath() string {
	if d.BasePath != "" {
		return d.BasePath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".dotfiles")
}

func (d Detector) existsFunc() PathExists {
	if d.Exists != nil {
		return d.Exists
	}
	return func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	}
}

func (d Detector) readDirFunc() ReadDir {
	if d.ReadDir != nil {
		return d.ReadDir
	}
	return os.ReadDir
}
