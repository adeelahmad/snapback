package service

// UnitOptions describes the systemd unit to render.
type UnitOptions struct {
	Exe    string
	Config string
	Scope  string
	User   string
}

// SystemdUnit renders the systemd unit file for o.
func SystemdUnit(o UnitOptions) (string, error) {
	panic("SUB-AGENT-TODO: T2 render unit matching testdata/*.service.golden with systemd quoting; User= only in system scope; invalid options -> errcode.InvalidConfig")
}
