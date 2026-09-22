package readerpolicy

// ProcName returns the command name of process pid from /proc, or "" when it
// cannot be read (for example on systems without /proc).
func ProcName(pid uint32) string {
	panic(`SUB-AGENT-TODO: return procNameAt("/proc", pid)`)
}

// procNameAt reads <root>/<pid>/comm and returns it without the trailing
// newline, or "" on any error.
func procNameAt(root string, pid uint32) string {
	panic(`SUB-AGENT-TODO: read filepath.Join(root, strconv.FormatUint(uint64(pid), 10), "comm"); on error return ""; else strings.TrimSuffix(content, "\n")`)
}
