package cli

import (
	"io"
	"strings"
)

// Next returns the next-step line naming command, which must be a literal
// runnable command. An empty, whitespace-only or multi-line command is a
// programming error and panics rather than printing a bare next-step line.
func Next(command string) string {
	if strings.TrimSpace(command) == "" {
		panic("cli: next-step command is empty")
	}
	if strings.Contains(command, "\n") {
		panic("cli: next-step command contains a newline")
	}
	return "next: " + command + "\n"
}

// WriteNext writes the next-step line naming command to w.
func WriteNext(w io.Writer, command string) error {
	_, err := io.WriteString(w, Next(command))
	return err
}
