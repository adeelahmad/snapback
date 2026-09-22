package readerpolicy

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ProcName returns the command name of the thread group that owns thread tid,
// read from /proc, or "" when it cannot be read (for example on systems
// without /proc).
func ProcName(tid uint32) string {
	return procNameAt("/proc", tid)
}

// procNameAt resolves tid to its thread-group leader via <root>/<tid>/status
// and returns that leader's <root>/<tgid>/comm without the trailing newline.
// FUSE reports the calling thread, whose name may differ from the process
// name. It falls back to the thread's own comm when status is unreadable or
// carries no Tgid line, and returns "" when nothing is readable.
func procNameAt(root string, tid uint32) string {
	if tgid, ok := procTgidAt(root, tid); ok && tgid != tid {
		if name := procCommAt(root, tgid); name != "" {
			return name
		}
	}
	return procCommAt(root, tid)
}

// procCommAt reads <root>/<pid>/comm without the trailing newline, or "" on
// any error.
func procCommAt(root string, pid uint32) string {
	b, err := os.ReadFile(filepath.Join(root, strconv.FormatUint(uint64(pid), 10), "comm"))
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(string(b), "\n")
}

// procTgidAt reads the Tgid field of <root>/<tid>/status, reporting whether it
// was found.
func procTgidAt(root string, tid uint32) (uint32, bool) {
	b, err := os.ReadFile(filepath.Join(root, strconv.FormatUint(uint64(tid), 10), "status"))
	if err != nil {
		return 0, false
	}
	for line := range strings.Lines(string(b)) {
		rest, found := strings.CutPrefix(line, "Tgid:")
		if !found {
			continue
		}
		tgid, err := strconv.ParseUint(strings.TrimSpace(rest), 10, 32)
		if err != nil {
			return 0, false
		}
		return uint32(tgid), true
	}
	return 0, false
}
