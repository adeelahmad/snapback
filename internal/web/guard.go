package web

import (
	"net"
	"net/http"
	"slices"
)

// guard sets security headers on every response, then rejects requests whose
// Host is neither the bound loopback address nor the configured bind (421) and
// cross-origin writes (403).
func (s *Server) guard(next http.Handler) http.Handler {
	_, port, _ := net.SplitHostPort(s.ln.Addr().String())
	allowed := map[string]bool{
		net.JoinHostPort("127.0.0.1", port): true,
		net.JoinHostPort("localhost", port): true,
		net.JoinHostPort("::1", port):       true,
	}
	if bind := s.opts.Policy.Bind; bind != "" {
		allowed[bind] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-Content-Type-Options", "nosniff")
		if !allowed[r.Host] {
			http.Error(w, "misdirected request", http.StatusMisdirectedRequest)
			return
		}
		if isWrite(r.Method) && s.isCrossOrigin(r) {
			http.Error(w, "cross-origin request", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isWrite(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	}
	return true
}

// isCrossOrigin reports whether r carries an Origin that is neither its own
// Host nor a configured allowed origin, or a Sec-Fetch-Site other than
// same-origin or none.
func (s *Server) isCrossOrigin(r *http.Request) bool {
	if o := r.Header.Get("Origin"); o != "" && o != "http://"+r.Host {
		if !slices.Contains(s.opts.Policy.AllowOrigins, o) {
			return true
		}
	}
	switch r.Header.Get("Sec-Fetch-Site") {
	case "", "same-origin", "none":
		return false
	}
	return true
}
