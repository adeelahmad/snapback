// Package setup reports what `snapback setup` can infer about a machine
// without writing anything: it only reads the environment through injected
// seams and never touches the filesystem or the Restic repository.
package setup

import "io/fs"

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

// Detect reports what setup can infer from deps. Seams left nil behave as if
// the environment, the binary, the working directory and the hostname were all
// unavailable, which yields a zero Result.
func Detect(deps Deps) (Result, error) {
	deps = withDefaults(deps)

	var res Result
	host, err := deps.Hostname()
	if err != nil {
		res.Reasons = append(res.Reasons, "hostname unavailable: "+err.Error())
		return res, nil
	}
	res.Hostname = host
	return res, nil
}

// withDefaults replaces every nil seam with one that reports nothing, so the
// rest of detection can call the seams unconditionally.
func withDefaults(deps Deps) Deps {
	if deps.Getenv == nil {
		deps.Getenv = func(string) string { return "" }
	}
	if deps.LookPath == nil {
		deps.LookPath = func(string) (string, error) { return "", fs.ErrNotExist }
	}
	if deps.Getwd == nil {
		deps.Getwd = func() (string, error) { return "", nil }
	}
	if deps.Hostname == nil {
		deps.Hostname = func() (string, error) { return "", nil }
	}
	return deps
}
