package web

import (
	"fmt"
	"net/http"
	"os/exec"
	"slices"
	"strconv"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/status"
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
	v := webui.StatusView{Chrome: s.chrome(r, "status", "Status")}
	if s.opts.Backend != nil {
		switch st := s.opts.Backend.Status().(type) {
		case status.Snapshot:
			v = statusView(v, st)
		case error:
			v.Errors = append(v.Errors, st.Error())
		}
	}
	s.render(w, "status", v)
}

// statusView fills v from the daemon snapshot st. Metrics the snapshot does
// not report stay nil, so the page shows them as not measured.
func statusView(v webui.StatusView, st status.Snapshot) webui.StatusView {
	for _, repo := range st.Repos {
		state := repo.State
		if state == "ready" {
			state = "mounted"
		}
		v.Mounts = append(v.Mounts, webui.MountStatus{Name: repo.ID, State: state})
		if repo.Code != "" {
			v.Errors = append(v.Errors, repo.ID+": "+string(repo.Code))
		}
	}
	if !st.LastRefresh.IsZero() {
		last := st.LastRefresh.Format(time.RFC3339)
		v.LastRefresh = &last
	}
	if st.EligibleCount != nil {
		n := 0
		for _, c := range st.EligibleCount {
			n += c
		}
		v.EligibleSnapshots = &n
	}
	if st.Throttle != nil {
		n := len(st.Throttle)
		v.ThrottleEvents = &n
	}
	p := st.Prewarm
	v.Prewarm = fmt.Sprintf("%d warm, %d cold, %d pending", p.Warm, p.Cold, p.Pending)
	v.DiscoveryMode = st.Discovery
	return v
}

// handleSetup prefills the Setup form from the stored config and offers the
// configured and PATH-detected restic and rclone binaries.
func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	v := webui.SetupView{Chrome: s.chrome(r, "setup", "Setup")}
	var cfg *config.Config
	if s.opts.Backend != nil {
		c, _, err := s.opts.Backend.Config()
		if err != nil {
			v.Errors = append(v.Errors, err.Error())
		}
		cfg = c
	}
	var restic, rclone []string
	if cfg != nil {
		for _, repo := range cfg.Repositories {
			restic = append(restic, repo.ResticBinary)
			rclone = append(rclone, repo.RcloneBinary)
		}
		if len(cfg.Repositories) > 0 {
			v.RepoURI = cfg.Repositories[0].Repository
			v.CredentialFile = cfg.Repositories[0].PasswordFile
		}
		for _, root := range cfg.Roots {
			v.Roots = append(v.Roots, root.LocalPath)
		}
	}
	v.ResticPaths = binaries(restic, "restic")
	v.RclonePaths = binaries(rclone, "rclone")
	s.render(w, "setup", v)
}

// binaries returns the non-empty configured paths followed by name's PATH
// lookup, without duplicates.
func binaries(configured []string, name string) []string {
	if p, err := exec.LookPath(name); err == nil {
		configured = append(configured, p)
	}
	var out []string
	for _, p := range configured {
		if p != "" && !slices.Contains(out, p) {
			out = append(out, p)
		}
	}
	return out
}

// fillConfigView copies the fields the Config form shows from cfg into v.
func fillConfigView(v *webui.ConfigView, cfg *config.Config) {
	for _, root := range cfg.Roots {
		v.Roots = append(v.Roots, root.LocalPath)
		if root.Snapshots.Hostname != "" {
			v.Filters = append(v.Filters, "host="+root.Snapshots.Hostname)
		}
		for _, tag := range root.Snapshots.TagsAll {
			v.Filters = append(v.Filters, "tag="+tag)
		}
		v.Exclusions = append(v.Exclusions, root.ExcludeRelativePaths...)
		for _, seed := range root.SeedPaths {
			v.SeedPaths = append(v.SeedPaths, seed.Path)
		}
	}
	for _, repo := range cfg.Repositories {
		if repo.CacheDir != "" {
			v.CacheDir = repo.CacheDir
			break
		}
	}
	v.DiscoveryMode = cfg.Discovery.Mode
	if cfg.Catalog.RefreshInterval > 0 {
		v.RefreshInterval = cfg.Catalog.RefreshInterval.String()
	}
}

func (s *Server) handleIntegrations(w http.ResponseWriter, r *http.Request) {
	s.render(w, "integrations", webui.IntegrationsView{Chrome: s.chrome(r, "integrations", "Integrations")})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	v := webui.ConfigView{Chrome: s.chrome(r, "config", "Config")}
	if s.opts.Backend != nil {
		cfg, rev, err := s.opts.Backend.Config()
		if err != nil {
			v.Errors = append(v.Errors, err.Error())
		}
		v.Revision = string(rev)
		if cfg != nil {
			fillConfigView(&v, cfg)
		}
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
