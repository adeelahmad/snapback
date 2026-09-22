package resolver

import "errors"

// RootSpec is a configured backup root: its ID and its local path.
type RootSpec struct {
	ID, LocalPath string
}

// Match is the root that contains a directory and the directory's
// slash-separated path relative to that root.
type Match struct {
	RootID, Rel string
}

var (
	// ErrOutsideRoots means no configured root contains the directory.
	ErrOutsideRoots = errors.New("outside configured roots")
	// ErrAmbiguousRoot means two roots share the same cleaned local path.
	ErrAmbiguousRoot = errors.New("ambiguous root")
)

// SelectRoot returns the root with the longest local path that contains dir.
func SelectRoot(roots []RootSpec, dir string) (Match, error) {
	panic("SUB-AGENT-TODO: T2 filepath.Clean dir+roots; relative dir -> ErrOutsideRoots; contain = equal or prefix root+\"/\" (\"/\" contains all); longest wins; tie -> ErrAmbiguousRoot; Rel no leading/trailing /; wrap fmt.Errorf(\"resolver: %q: %w\"), zero Match on error")
}
