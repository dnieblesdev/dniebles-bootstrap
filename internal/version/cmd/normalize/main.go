package main

import (
	"flag"
	"fmt"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/version"
)

func main() {
	var v string
	flag.StringVar(&v, "version", "", "Git version metadata to normalize")
	flag.Parse()

	fmt.Println(version.NormalizeGitVersion(v))
}
