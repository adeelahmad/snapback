package fidelity

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
	"time"
)

// resticNode is one line of restic ls --json output.
type resticNode struct {
	StructType string      `json:"struct_type"`
	Type       string      `json:"type"`
	Path       string      `json:"path"`
	Size       int64       `json:"size"`
	Mode       fs.FileMode `json:"mode"`
	MTime      string      `json:"mtime"`
	LinkTarget string      `json:"linktarget"`
}

// ParseResticLs parses the newline-delimited JSON of restic ls --json and
// returns the file and symlink entries with paths relative to root.
func ParseResticLs(r io.Reader, root string) ([]Meta, error) {
	prefix := strings.TrimSuffix(root, "/") + "/"
	var metas []Meta
	sc := bufio.NewScanner(r)
	for line := 1; sc.Scan(); line++ {
		if strings.TrimSpace(sc.Text()) == "" {
			continue
		}
		var n resticNode
		if err := json.Unmarshal(sc.Bytes(), &n); err != nil {
			return nil, fmt.Errorf("restic ls line %d: %w", line, err)
		}
		if n.StructType != "node" || (n.Type != "file" && n.Type != "symlink") {
			continue
		}
		mt, err := time.Parse(time.RFC3339Nano, n.MTime)
		if err != nil {
			return nil, fmt.Errorf("restic ls line %d: mtime: %w", line, err)
		}
		metas = append(metas, Meta{
			Path:       path.Clean(strings.TrimPrefix(n.Path, prefix)),
			Size:       n.Size,
			Mode:       n.Mode,
			MTime:      mt,
			LinkTarget: n.LinkTarget,
		})
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read restic ls output: %w", err)
	}
	return metas, nil
}
