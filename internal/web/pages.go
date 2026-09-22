package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/webui"
)

// chrome returns the shared layout data for page name, carrying r's session
// CSRF token.
func (s *Server) chrome(r *http.Request, name webui.PageName, title string) webui.Chrome {
	return webui.Chrome{Title: title, Active: name, CSRFToken: s.csrfToken(r)}
}

// render writes page name with data as HTML.
func (s *Server) render(w http.ResponseWriter, name webui.PageName, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.opts.Pages.Render(w, name, data); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.render(w, "status", webui.StatusView{Chrome: s.chrome(r, "status", "Status")})
}

func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	s.render(w, "setup", webui.SetupView{Chrome: s.chrome(r, "setup", "Setup")})
}

func (s *Server) handleIntegrations(w http.ResponseWriter, r *http.Request) {
	s.render(w, "integrations", webui.IntegrationsView{Chrome: s.chrome(r, "integrations", "Integrations")})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	v := webui.ConfigView{Chrome: s.chrome(r, "config", "Config")}
	if s.opts.Backend != nil {
		_, rev, err := s.opts.Backend.Config()
		if err != nil {
			v.Errors = append(v.Errors, err.Error())
		}
		v.Revision = string(rev)
	}
	s.render(w, "config", v)
}

// handleHistory lists the roots and, when root and snapshot are given, the
// entries of path in that snapshot. html/template escapes every name.
func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	v := webui.HistoryView{Chrome: s.chrome(r, "history", "History")}
	if s.opts.History == nil {
		s.render(w, "history", v)
		return
	}
	for _, root := range s.opts.History.Roots() {
		v.Roots = append(v.Roots, webui.RootItem{Name: root.ID, Path: root.Path, RepoState: root.State})
	}
	q := r.URL.Query()
	root, id := q.Get("root"), provider.SnapshotID(q.Get("snapshot"))
	if root != "" && id != "" {
		dir, err := s.resolve(root, q.Get("path"))
		if err != nil {
			writeError(w, http.StatusBadRequest, errcode.InvalidConfig, err)
			return
		}
		entries, err := s.opts.History.List(r.Context(), root, dir, id)
		if err != nil {
			http.Error(w, "list failed", http.StatusInternalServerError)
			return
		}
		for _, e := range entries {
			state := "ok"
			if e.Absent {
				state = "absent"
			}
			v.Entries = append(v.Entries, webui.Entry{
				Name:     e.Name,
				Size:     strconv.FormatInt(e.Size, 10),
				Modified: e.ModTime.Format(time.RFC3339),
				State:    state,
			})
		}
	}
	s.render(w, "history", v)
}
