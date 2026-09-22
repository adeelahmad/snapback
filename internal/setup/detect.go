// Package setup reports what `snapback setup` can infer about a machine
// without writing anything: it only reads the environment through injected
// seams and never touches the filesystem or the Restic repository.
package setup

import "errors"

// Result holds everything detection could infer. Every zero value is
// meaningful: an empty string or a nil slice means "not detected".
type Result struct {
	RepoURI        string
	CredentialFile string
	ResticPath     string
	RcloneFound    bool
	Roots          []string
	Hostname       string
	Reasons        []string
}

// Deps carries the seams detection reads the machine through. A nil seam is
// treated as unset or absent, never as a reason to panic.
type Deps struct {
	Getenv   func(string) string
	LookPath func(string) (string, error)
	Getwd    func() (string, error)
	Hostname func() (string, error)
}

// Detect reports what setup can infer from deps.
func Detect(deps Deps) (Result, error) {
	return Result{}, errors.New("setup: detect not implemented")
}
