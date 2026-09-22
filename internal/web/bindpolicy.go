package web

import (
	"fmt"
	"net"
	"net/url"
	"slices"

	"github.com/adeelahmad/snapback/internal/errcode"
)

const (
	// policyOp names this decision in the errors it returns.
	policyOp = "web.Policy"
	// defaultBind mirrors the config Web.Listen default, which config keeps
	// unexported.
	defaultBind = "127.0.0.1:0"
	// remoteBindWarning is the single line a non-loopback bind must print.
	remoteBindWarning = "warning: snapback web is reachable from other machines on %s; the session token is the only protection"
)

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

// Policy decides the bind address for bind, allowRemote and allowOrigins. An
// empty bind falls back to the loopback default. A non-loopback bind needs
// allowRemote and carries a warning; without the flag it is refused. Every
// origin must be a scheme+host URL with no path, query or fragment.
func Policy(bind string, allowRemote bool, allowOrigins []string) (BindPolicy, error) {
	if bind == "" {
		bind = defaultBind
	}
	host, _, err := net.SplitHostPort(bind)
	if err != nil {
		return BindPolicy{}, errcode.New(errcode.InvalidConfig, policyOp, err)
	}
	for _, origin := range allowOrigins {
		if err := checkOrigin(origin); err != nil {
			return BindPolicy{}, errcode.New(errcode.InvalidConfig, policyOp, err)
		}
	}
	p := BindPolicy{Bind: bind, AllowOrigins: slices.Clone(allowOrigins)}
	if !isLoopbackHost(host) {
		if !allowRemote {
			return BindPolicy{}, errcode.New(errcode.InvalidConfig, policyOp,
				fmt.Errorf("bind address %q is not loopback; pass --allow-remote to serve other machines", bind))
		}
		p.Warning = fmt.Sprintf(remoteBindWarning, bind)
	}
	return p, nil
}

// isLoopbackHost reports whether host is a loopback IP or the loopback name.
func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// checkOrigin rejects anything that is not a bare http or https scheme+host.
func checkOrigin(origin string) error {
	bad := fmt.Errorf("--allow-origin %q is not a scheme://host origin", origin)
	u, err := url.Parse(origin)
	if err != nil {
		return bad
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return bad
	}
	if u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return bad
	}
	return nil
}
