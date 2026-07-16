package execution

import "errors"

const (
	homebrewInstallerURL    = "https://raw.githubusercontent.com/Homebrew/install/4b0227cf8416504142d23893368c2e1d211d5191/install.sh"
	homebrewInstallerDigest = "99287f194a8b3c9e6b0203a11a5fa54518be57209343e6bb954dec4635796d9d"

	homebrewInstallerCommitPermalink = "https://github.com/Homebrew/install/blob/4b0227cf8416504142d23893368c2e1d211d5191/install.sh"
	homebrewInstallerRetrievedAt     = "2026-07-16"
	homebrewInstallerReviewMethod    = "curl --fail --location <raw-URL> | sha256sum; GitHub API verified commit signature"
	homebrewDefaultBinary            = "/home/linuxbrew/.linuxbrew/bin/brew"
)

var ErrHomebrewRedirect = errors.New("Homebrew installer redirect or effective URL change rejected")

// validatePinnedDownload accepts only the reviewed literal URL. Downloaders
// must pass their effective URL so redirects and substitutions fail closed.
func validatePinnedDownload(requestedURL, effectiveURL string) error {
	if requestedURL != homebrewInstallerURL || effectiveURL != homebrewInstallerURL {
		return ErrHomebrewRedirect
	}
	return nil
}
