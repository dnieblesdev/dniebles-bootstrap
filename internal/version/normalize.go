package version

import "strings"

// safeGitVersionChars is the set of runes that are safe to use in filesystem
// paths and archive names derived from Git metadata.
const safeGitVersionChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-"

// NormalizeGitVersion converts arbitrary Git metadata (tag, branch, or describe
// output) into a string safe for use in file paths and archive names. It keeps
// only [A-Za-z0-9._-], collapses each run of invalid characters (including '/')
// into a single '-', trims leading and trailing separators ('-', '_', '.'), and
// falls back to "dev" if the result is empty.
func NormalizeGitVersion(v string) string {
	if v == "" {
		return "dev"
	}

	var b strings.Builder
	lastWasInvalid := false
	for _, r := range v {
		if strings.ContainsRune(safeGitVersionChars, r) {
			b.WriteRune(r)
			lastWasInvalid = false
			continue
		}
		if !lastWasInvalid {
			b.WriteRune('-')
			lastWasInvalid = true
		}
	}

	sanitized := strings.Trim(b.String(), "-_.")
	if sanitized == "" {
		return "dev"
	}
	return sanitized
}
