package cli

import "flag"

// WriteOK writes a success result.
func WriteOK(env Env, jsonOut bool, data any) int {
	panic(`SUB-AGENT-TODO: jsonOut -> {"ok":true,"data":data} on env.Stdout; else human text on env.Stdout; return 0`)
}

// WriteError writes a failure result.
func WriteError(env Env, cmd string, jsonOut bool, err error) int {
	panic(`SUB-AGENT-TODO: jsonOut -> {"ok":false,"code":<errcode>,"error":msg,"fix":<per-code fix table, all 11 errcode codes non-empty>} on env.Stdout; else "snapback <cmd>: <msg>" plus "fix: <action>" on env.Stderr; return ExitCode(err)`)
}

// ParseFlags parses args with fs, adding --json.
func ParseFlags(fs *flag.FlagSet, args []string) (jsonOut bool, pos []string, err error) {
	panic("SUB-AGENT-TODO: register --json bool on fs, fs.Parse(args) ('--' ends flags), return the json value and fs.Args(); parse failures become *UsageError")
}
