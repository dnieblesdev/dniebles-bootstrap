package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

// defaultConfigBase is the convention-based dotfiles configuration directory.
const defaultConfigBase = ".dotfiles/config"

// PathExists reports whether a path exists.
type PathExists func(path string) bool

// KeyPathResolver maps a config key to a filesystem path under basePath.
// The second return value is false when the key is invalid or unsafe.
type KeyPathResolver func(basePath, key string) (string, bool)

// Detector inspects a catalog and reports which required config keys appear present.
type Detector struct {
	BasePath   string
	Exists     PathExists
	PathForKey KeyPathResolver
}

// Detect returns config state for the given catalog using the default detector.
func Detect(catalog planning.Catalog) planning.ConfigState {
	return Detector{}.Detect(catalog)
}

// Detect returns config state for the given catalog using the detector's seams.
func (d Detector) Detect(catalog planning.Catalog) planning.ConfigState {
	basePath := d.BasePath
	if basePath == "" {
		basePath = defaultBasePath()
	}

	exists := d.Exists
	if exists == nil {
		exists = defaultPathExists
	}

	pathForKey := d.PathForKey
	if pathForKey == nil {
		pathForKey = defaultKeyPathResolver
	}

	present := map[string]bool{}
	for _, key := range requiredKeys(catalog) {
		path, ok := pathForKey(basePath, key)
		if !ok {
			continue
		}
		if exists(path) {
			present[key] = true
		}
	}

	return planning.ConfigState{PresentKeys: present}
}

func defaultBasePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, defaultConfigBase)
}

func defaultPathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func defaultKeyPathResolver(basePath, key string) (string, bool) {
	if basePath == "" || key == "" || filepath.IsAbs(key) {
		return "", false
	}

	segments := strings.Split(key, ".")
	for _, segment := range segments {
		if segment == "" || segment == ".." || strings.Contains(segment, string(filepath.Separator)) || strings.Contains(segment, "/") {
			return "", false
		}
	}

	return filepath.Join(append([]string{basePath}, segments...)...), true
}

func requiredKeys(catalog planning.Catalog) []string {
	seen := map[string]bool{}
	var keys []string
	for _, resource := range catalog.Resources {
		for _, key := range resource.ConfigPolicy.RequiredKeys {
			if !seen[key] {
				seen[key] = true
				keys = append(keys, key)
			}
		}
	}
	return keys
}
