// agentic:shim
package service

// UnitOptions is a T2 compile shim; SCAFFOLD replaces it.
type UnitOptions struct {
	Exe    string
	Config string
	Scope  string
	User   string
}

// SystemdUnit is a T2 compile shim with a deliberately wrong body.
func SystemdUnit(o UnitOptions) (string, error) {
	_ = o
	return "shim: not a unit\n", nil
}
