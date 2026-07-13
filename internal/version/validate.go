package version

import (
	"fmt"
	"regexp"
	"strings"
)

const maxVersionLength = 64

var (
	versionPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._+-]{0,63}$`)

	// releaseTagPattern matches v-prefixed SemVer 2.0.0 core versions with
	// optional prerelease and build metadata. It does not enforce the leading-zero
	// rule for numeric prerelease identifiers; that is checked separately.
	releaseTagPattern = regexp.MustCompile(`^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)

	semverIdentifierPattern = regexp.MustCompile(`^[0-9A-Za-z-]+$`)
)

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

// ValidateReleaseTag accepts only v-prefixed strict SemVer tags. It returns
// true when the tag includes a prerelease component (for example v1.2.3-rc.1)
// and false for stable or build-metadata-only tags (for example v1.2.3 or
// v1.2.3+build.123).
func ValidateReleaseTag(v string) (isPrerelease bool, err error) {
	if v == "" {
		return false, fmt.Errorf("release tag must not be empty")
	}
	if len(v) > maxVersionLength {
		return false, fmt.Errorf("release tag %q is too long (max %d characters)", v, maxVersionLength)
	}

	matches := releaseTagPattern.FindStringSubmatch(v)
	if matches == nil {
		return false, fmt.Errorf("release tag %q is not a valid v-prefixed SemVer tag", v)
	}

	prerelease := matches[4]
	if prerelease != "" {
		if err := validateSemVerPrerelease(prerelease); err != nil {
			return false, fmt.Errorf("release tag %q has invalid prerelease: %w", v, err)
		}
	}

	build := matches[5]
	if build != "" {
		if err := validateSemVerBuildMetadata(build); err != nil {
			return false, fmt.Errorf("release tag %q has invalid build metadata: %w", v, err)
		}
	}

	return prerelease != "", nil
}

func validateSemVerPrerelease(v string) error {
	for _, id := range strings.Split(v, ".") {
		if id == "" {
			return fmt.Errorf("prerelease identifier must not be empty")
		}
		if !semverIdentifierPattern.MatchString(id) {
			return fmt.Errorf("prerelease identifier %q contains invalid characters", id)
		}
		if isAllDigits(id) {
			if len(id) > 1 && id[0] == '0' {
				return fmt.Errorf("numeric prerelease identifier %q must not have leading zeros", id)
			}
		}
	}
	return nil
}

func validateSemVerBuildMetadata(v string) error {
	for _, id := range strings.Split(v, ".") {
		if id == "" {
			return fmt.Errorf("build metadata identifier must not be empty")
		}
		if !semverIdentifierPattern.MatchString(id) {
			return fmt.Errorf("build metadata identifier %q contains invalid characters", id)
		}
	}
	return nil
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
