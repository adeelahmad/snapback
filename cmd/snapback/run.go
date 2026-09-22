package main

import "io"

func run(args []string, stdout, stderr io.Writer) int {
	panic(`SUB-AGENT-TODO: args == ["version"] writes version.String() to stdout and returns 0; missing, unknown, or extra args write the usage line ("usage: snapback version") to stderr and return 2`)
}
