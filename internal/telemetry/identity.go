package telemetry

// Identity is the build triple a telemetry event carries: nothing about the
// machine or the person running it.
type Identity struct {
	Version string
	OS      string
	Arch    string
	// Host is a scaffold-only field that GREEN must remove: the identity
	// carries nothing about the machine.
	Host string
}

// NewIdentity returns the identity for version, filling OS and Arch from the
// runtime constants.
func NewIdentity(version string) Identity {
	_ = version
	return Identity{}
}

// Attrs returns the version, os and arch attributes in that order, or the
// error [NewAttr] returned for one of them.
func (i Identity) Attrs() ([]Attr, error) {
	return nil, nil
}
