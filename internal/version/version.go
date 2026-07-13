// Package version exposes the build-time version for dbootstrap.
//
// The Version variable is initialized to "dev" so that ordinary local builds
// report a stable default. Release builds override it with -ldflags -X.
package version

// Version is the current dbootstrap version. It may be overridden at link
// time using -ldflags -X github.com/dnieblesdev/dniebles-bootstrap/internal/version.Version=<value>.
var Version = "dev"
