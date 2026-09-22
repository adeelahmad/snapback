package cli

import (
	"strconv"
	"strings"
	"unicode"
)

// NextPath returns the next-step line for running verb against path. The path
// is quoted when it is empty or holds a space or a control character, so the
// suggested command stays one runnable line; otherwise it is shown verbatim.
func NextPath(verb, path string) string {
	return "next: " + verb + " " + renderNextPath(path) + "\n"
}

func renderNextPath(path string) string {
	if path == "" || strings.ContainsFunc(path, needsNextQuote) {
		return strconv.Quote(path)
	}
	return path
}

func needsNextQuote(r rune) bool {
	return r == ' ' || unicode.IsControl(r)
}
