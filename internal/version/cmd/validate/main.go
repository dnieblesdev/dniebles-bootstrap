package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/version"
)

func main() {
	var v string
	flag.StringVar(&v, "version", "", "version string to validate")
	flag.Parse()

	if err := version.Validate(v); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
