package readerpolicy

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ProcName returns the command name of process pid from /proc, or "" when it
// cannot be read (for example on systems without /proc).
func ProcName(pid uint32) string {
	return procNameAt("/proc", pid)
}

// procNameAt reads <root>/<pid>/comm and returns it without the trailing
// newline, or "" on any error.
func procNameAt(root string, pid uint32) string {
	b, err := os.ReadFile(filepath.Join(root, strconv.FormatUint(uint64(pid), 10), "comm"))
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(string(b), "\n")
}
