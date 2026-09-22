package restic

// CommandLine renders a restic invocation for logs with every secret redacted.
//
// The output is a single space-separated line: the env pairs in the order given,
// then the argv in the order given. An argv field is quoted with %q only when it
// contains a space; an env pair quotes only its value, as KEY="with space".
// Nil or empty args and env render as the empty string.
func CommandLine(args, env []string) string {
	return ""
}
