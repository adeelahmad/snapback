// agentic:shim

// Package rawpath encodes raw path bytes reversibly as JSON.
package rawpath

// Path holds raw path bytes.
type Path []byte

// MarshalJSON is a deliberately wrong RED shim.
func (p Path) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

// UnmarshalJSON is a deliberately wrong RED shim.
func (p *Path) UnmarshalJSON(data []byte) error {
	return nil
}
