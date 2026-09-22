// agentic:shim
package web

import (
	"context"
	"io"
)

// cliEnv is the local stand-in for cli.Env until S3-09 merges.
type cliEnv struct {
	Stdout, Stderr io.Writer
	Getenv         func(string) string
	ConfigPath     string
}

// cliCommand is the local stand-in for cli.Command until S3-09 merges.
type cliCommand struct {
	Name, Summary string
	Run           func(ctx context.Context, env cliEnv, args []string) int
}

var goos = "shim"

var openBrowser = func(ctx context.Context, url string) error { return nil }

// Command is a deliberately wrong shim.
func Command() cliCommand {
	return cliCommand{Name: "shim", Run: func(context.Context, cliEnv, []string) int { return 99 }}
}

// ConfigCommand is a deliberately wrong shim.
func ConfigCommand() cliCommand {
	return cliCommand{Name: "shim", Run: func(context.Context, cliEnv, []string) int { return 99 }}
}
