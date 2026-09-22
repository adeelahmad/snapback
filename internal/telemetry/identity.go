package telemetry

import "runtime"

// Identity is the build triple a telemetry event carries: nothing about the
// machine or the person running it.
type Identity struct {
	Version string
	OS      string
	Arch    string
}

// NewIdentity returns the identity for version, filling OS and Arch from the
// runtime constants.
func NewIdentity(version string) Identity {
	return Identity{Version: version, OS: runtime.GOOS, Arch: runtime.GOARCH}
}

// Attrs returns the version, os and arch attributes in that order, or the
// error [NewAttr] returned for one of them.
func (i Identity) Attrs() ([]Attr, error) {
	pairs := [...]struct{ key, value string }{
		{"version", i.Version},
		{"os", i.OS},
		{"arch", i.Arch},
	}
	attrs := make([]Attr, 0, len(pairs))
	for _, p := range pairs {
		attr, err := NewAttr(p.key, p.value)
		if err != nil {
			return nil, err
		}
		attrs = append(attrs, attr)
	}
	return attrs, nil
}
