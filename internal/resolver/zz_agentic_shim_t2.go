// agentic:shim

// Package resolver is a RED compile shim for S3-03 T2; the scaffolder replaces it.
package resolver

import "errors"

// RootSpec is a shim.
type RootSpec struct {
	ID, LocalPath string
}

// Match is a shim.
type Match struct {
	RootID, Rel string
}

// ErrOutsideRoots and ErrAmbiguousRoot are shims.
var ErrOutsideRoots, ErrAmbiguousRoot error = errors.New("shim outside roots"), errors.New("shim ambiguous root")

// SelectRoot is a shim that deliberately returns a wrong result.
func SelectRoot(roots []RootSpec, dir string) (Match, error) {
	return Match{}, nil
}

// DirectoryKey is a shim that deliberately returns a wrong result.
func DirectoryKey(rootID, rel string) string {
	return ""
}
