package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/version"
)

func main() {
	var v string
	var release bool
	flag.StringVar(&v, "version", "", "version string to validate")
	flag.BoolVar(&release, "release", false, "use strict release-tag validation and emit prerelease state")
	flag.Parse()

	if release {
		isPrerelease, err := version.ValidateReleaseTag(v)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("prerelease=%t\n", isPrerelease)
		return
	}

	if err := version.Validate(v); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
