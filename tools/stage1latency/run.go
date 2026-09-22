package main

import "io"

// run parses args, performs one latency run and writes the JSON result; it returns the process exit code.
func run(args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	panic("SUB-AGENT-TODO: parse required -out <path>; remote via latency.RemoteFromEnv (refused remote or missing flag -> exit 2); latency.Run with newConfig(remote); write latency.Encode output to -out, even on failure when a Result exists; print the path; exit 0 only when Run succeeded and remote_deleted is true, else 1")
}
