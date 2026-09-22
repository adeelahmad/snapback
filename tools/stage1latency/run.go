package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/adeelahmad/snapback/internal/compat/latency"
)

// run parses args, performs one latency run and writes the JSON result; it returns the process exit code.
func run(args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("stage1latency", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("out", "", "path to write the latency JSON result (required)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *out == "" {
		_, _ = fmt.Fprintln(stderr, "stage1latency: -out <path> is required")
		return 2
	}
	remote, err := latency.RemoteFromEnv(getenv)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "stage1latency:", err)
		return 2
	}

	res, runErr := latency.Run(context.Background(), newConfig(remote))
	code := 0
	if runErr != nil {
		_, _ = fmt.Fprintln(stderr, "stage1latency:", runErr)
		code = 1
	}
	// A zero Remote means Run failed before any remote work, so there is no Result to write.
	if res.Remote == "" {
		return 1
	}
	if !res.RemoteDeleted {
		_, _ = fmt.Fprintf(stderr, "stage1latency: REMOTE %s NOT DELETED; purge it by hand\n", latency.AllowedRemote)
		code = 1
	}
	b, err := latency.Encode(res)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "stage1latency:", err)
		return 1
	}
	if err := os.WriteFile(*out, b, 0o600); err != nil {
		_, _ = fmt.Fprintln(stderr, "stage1latency:", err)
		return 1
	}
	_, _ = fmt.Fprintln(stdout, *out)
	return code
}
