package version

import (
	"fmt"
	"regexp"
)

const maxVersionLength = 64

var versionPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._+-]{0,63}$`)

// Validate returns nil if v is safe to use as a version string in shell
// commands, filenames, and ldflags. An empty string is accepted because it
// signals that the default git-derived version should be used.
func Validate(v string) error {
	if v == "" {
		return nil
	}
	if len(v) > maxVersionLength {
		return fmt.Errorf("version %q is too long (max %d characters)", v, maxVersionLength)
	}
	if !versionPattern.MatchString(v) {
		return fmt.Errorf("version %q contains invalid characters", v)
	}
	return nil
}
