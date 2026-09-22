package main

import (
	"fmt"
	"io"

	"github.com/adeelahmad/snapback/internal/version"
)

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "version" {
		_, _ = fmt.Fprint(stdout, version.String())
		return 0
	}
	_, _ = fmt.Fprintln(stderr, "usage: snapback version")
	return 2
}
