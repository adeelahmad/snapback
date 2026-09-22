package resticfx

import (
	"os"
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
	panic("SUB-AGENT-TODO: reject absolute/.. escapes; mkdir parents; write Size deterministic bytes; chmod Mode after write; os.Chtimes ModTime; return Lstat metadata")
}
