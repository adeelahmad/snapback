package resticfx

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileSpec describes one file WriteTree creates.
type FileSpec struct {
	Path    string
	Size    int
	Mode    os.FileMode
	ModTime time.Time
}

// FileMeta is the metadata of a written file as read back via os.Lstat.
type FileMeta struct {
	Path    string
	Size    int
	Mode    os.FileMode
	ModTime time.Time
}

// WriteTree writes a deterministic file tree under root.
func WriteTree(root string, specs []FileSpec) ([]FileMeta, error) {
	metas := make([]FileMeta, 0, len(specs))
	for _, spec := range specs {
		rel := filepath.FromSlash(spec.Path)
		if !filepath.IsLocal(rel) {
			return nil, fmt.Errorf("fixture path %q escapes root", spec.Path)
		}
		full := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return nil, fmt.Errorf("mkdir for %q: %w", spec.Path, err)
		}
		data := make([]byte, spec.Size)
		for i := range data {
			data[i] = byte(i % 251)
		}
		if err := os.WriteFile(full, data, 0o600); err != nil {
			return nil, fmt.Errorf("write %q: %w", spec.Path, err)
		}
		if err := os.Chmod(full, spec.Mode.Perm()); err != nil {
			return nil, fmt.Errorf("chmod %q: %w", spec.Path, err)
		}
		if err := os.Chtimes(full, spec.ModTime, spec.ModTime); err != nil {
			return nil, fmt.Errorf("chtimes %q: %w", spec.Path, err)
		}
		fi, err := os.Lstat(full)
		if err != nil {
			return nil, fmt.Errorf("lstat %q: %w", spec.Path, err)
		}
		metas = append(metas, FileMeta{Path: spec.Path, Size: int(fi.Size()), Mode: fi.Mode(), ModTime: fi.ModTime()})
	}
	return metas, nil
}
