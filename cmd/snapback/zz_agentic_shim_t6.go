// agentic:shim
package main

import "github.com/adeelahmad/snapback/internal/cli"

func coreCommands(_ cli.Deps) []cli.Command {
	return []cli.Command{{Name: "shim"}}
}
