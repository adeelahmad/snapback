package web

// BindPolicy is the decided listen address, the accepted browser origins and
// the operator warning that a non-loopback bind must print.
type BindPolicy struct {
	// Bind is the address the web server listens on.
	Bind string
	// AllowOrigins holds the extra browser origins the server accepts,
	// verbatim as scheme+host[:port].
	AllowOrigins []string
	// Warning is the single line to print for a remote bind, empty for a
	// loopback bind.
	Warning string
}

// Policy decides the bind address for bind, allowRemote and allowOrigins.
//
// SUB-AGENT-TODO: compile shim only; GREEN implements the real decision.
func Policy(bind string, allowRemote bool, allowOrigins []string) (BindPolicy, error) {
	_, _, _ = bind, allowRemote, allowOrigins
	return BindPolicy{}, nil
}
