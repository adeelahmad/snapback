package cli

import (
	"io"
	"strconv"
	"strings"
	"unicode"
)

// nextNothing is the next-step line used when there is no command to suggest.
const nextNothing = "(nothing to run)"

// Next returns the next-step line naming command. The line is always a single
// line: an empty or whitespace-only command yields a placeholder, and a command
// holding a control character is quoted so nothing raw reaches the terminal.
// Printing a hint never takes the command down, so Next does not panic.
func Next(command string) string {
	if strings.TrimSpace(command) == "" {
		return "next: " + nextNothing + "\n"
	}
	if strings.ContainsFunc(command, unicode.IsControl) {
		return "next: " + strconv.Quote(command) + "\n"
	}
	return "next: " + command + "\n"
}

// WriteNext writes the next-step line naming command to w.
func WriteNext(w io.Writer, command string) error {
	_, err := io.WriteString(w, Next(command))
	return err
}
