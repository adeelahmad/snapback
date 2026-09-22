package cli

import "strings"

// usageIndent prefixes every line of a usage block body.
const usageIndent = "  "

// Usage is the help text of a single subcommand.
type Usage struct {
	// Synopsis is the invocation without the program name, such as
	// "seed [flags] DIR".
	Synopsis string
	// Args describes the positional arguments, one per line. It is omitted
	// when empty.
	Args string
	// Example is a sample invocation. It is omitted when empty.
	Example string
}

// String renders the usage as the text a command prints for -h: a synopsis
// line, then an optional Args block and an optional Example block, separated
// by single blank lines and ending in exactly one newline.
func (u Usage) String() string {
	var b strings.Builder
	b.WriteString("Usage: snapback ")
	b.WriteString(u.Synopsis)
	b.WriteString("\n")
	writeUsageBlock(&b, "Args", u.Args)
	writeUsageBlock(&b, "Example", u.Example)
	return b.String()
}

// writeUsageBlock appends a blank line, the block heading and the indented
// body. It writes nothing when body is empty.
func writeUsageBlock(b *strings.Builder, heading, body string) {
	if body == "" {
		return
	}
	b.WriteString("\n")
	b.WriteString(heading)
	b.WriteString(":\n")
	for _, line := range strings.Split(strings.TrimRight(body, "\n"), "\n") {
		b.WriteString(usageIndent)
		b.WriteString(line)
		b.WriteString("\n")
	}
}
