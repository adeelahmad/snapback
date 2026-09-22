package fidelity

import "io"

// ParseResticLs parses the newline-delimited JSON of restic ls --json and
// returns the file and symlink entries with paths relative to root.
func ParseResticLs(r io.Reader, root string) ([]Meta, error) {
	panic("SUB-AGENT-TODO: T2 read restic 0.19.0 NDJSON; skip struct_type snapshot line and dir nodes; keep file and symlink; map path (strip root prefix, relative, / separators), size, mode (numeric fs.FileMode), mtime (RFC 3339 nanos), linktarget; malformed line returns error naming the line number")
}
