package cli

import "io"

// Next returns the next-step line naming command.
func Next(command string) string {
	return ""
}

// WriteNext writes the next-step line naming command to w.
func WriteNext(w io.Writer, command string) error {
	return nil
}
