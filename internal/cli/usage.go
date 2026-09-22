package cli

// Usage is the help text of a single subcommand.
type Usage struct {
	Synopsis string
	Args     string
	Example  string
}

// String renders the usage as the text a command prints for -h.
func (u Usage) String() string {
	return ""
}
