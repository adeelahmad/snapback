package cli

// NextPath returns the next-step line for running verb against path.
//
// RED SHIM (NEXT-1): this renders path verbatim, which is today's behaviour and
// what next_edge_test.go proves wrong. GREEN replaces the body so that a path
// holding a control character or a space is rendered with %q-style quoting.
func NextPath(verb, path string) string {
	return "next: " + verb + " " + path + "\n"
}
