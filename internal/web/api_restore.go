package web

import "net/http"

// handleAPIRestore copies one file out of a snapshot into the live tree under
// a dated name, never overwriting an existing file.
func (s *Server) handleAPIRestore(w http.ResponseWriter, r *http.Request) {
	panic("SUB-AGENT-TODO: parse root/snapshot/file form fields, resolve+validate each " +
		"via s.resolve and a 64-hex snapshot id, reject a directory target, copy the " +
		"snapshot file into the live tree as 'name (YYYY-MM-DD).ext' opened with O_EXCL, " +
		"retrying with ' 2', ' 3', ... on EEXIST so an existing file is never overwritten, " +
		"then write {\"path\": <restored path>} as JSON (200) or a 400 on bad input")
}

// handleAPIOpen launches the host file manager on the directory containing
// one live-tree path, via s.opts.Opener.
func (s *Server) handleAPIOpen(w http.ResponseWriter, r *http.Request) {
	panic("SUB-AGENT-TODO: parse root/path form fields, resolve+validate path via " +
		"s.resolve, 400 on a path outside the root; if s.opts.Opener is nil respond 503; " +
		"otherwise call s.opts.Opener(r.Context(), <live dir containing path>) directly " +
		"(never through a shell) and respond 204 on success")
}
