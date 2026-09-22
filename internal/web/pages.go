package web

import (
	"fmt"
	"net/http"
	"net/url"
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
		if cfg, _, err := s.opts.Backend.Config(); err == nil && cfg != nil {
			v.DiscoveryMode = cfg.Discovery.Mode
		}
		if lc, ok := s.opts.Backend.(linkCounter); ok {
			if n, err := lc.ManagedLinks(); err == nil {
				v.ManagedLinks = &n
			}
		}
	}
	if s.opts.Daemon != nil {
		ctl, err := webui.RenderDaemonControl(webui.DaemonControlView{Running: s.opts.Daemon.Running(), CSRFToken: s.csrfToken(r)})
		if err != nil {
			v.Errors = append(v.Errors, err.Error())
		} else {
			v.DaemonControl = ctl
		}
	}
	s.render(w, "status", v)
}

// linkCounter is a Backend that can count the registry-owned links.
type linkCounter interface {
	ManagedLinks() (int, error)
}

// statusView fills v from the daemon snapshot st. Metrics the snapshot does
// not report stay nil, so the page shows them as not reported.
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
		c, rev, err := s.opts.Backend.Config()
		if err != nil {
			v.Errors = append(v.Errors, err.Error())
		}
		cfg = c
		if rev == "" {
			tour, err := webui.RenderTour(webui.SetupTour())
			if err != nil {
				v.Errors = append(v.Errors, err.Error())
			}
			v.Tour = tour
		}
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
	d := s.detect(r.Context())
	restic = append(restic, d.ResticPath)
	rclone = append(rclone, d.RclonePath)
	if v.RepoURI == "" {
		v.RepoURI = d.RepoURI
	}
	if v.CredentialFile == "" {
		v.CredentialFile = d.CredentialFile
	}
	if len(v.Roots) == 0 {
		v.Roots = d.Roots
	}
	v.DetectedHost = webui.Control{
		Path:  "hostname",
		Label: "Snapshot hostname",
		Value: d.Hostname,
		Help:  "The hostname Snapback resolves snapshots under.",
	}
	v.PasswordControls = passwordControls(setupRepoPrefix, nil, nil)
	v.DetectedPrefix = webui.Control{Path: "prefix_map", Label: "Detected prefix map"}
	for _, m := range d.PrefixMap {
		v.DetectedPrefix.Options = append(v.DetectedPrefix.Options,
			webui.ControlOption{Value: m.SourcePath, Label: m.SourcePath})
	}
	v.MountPoint = mountPointField(cfg, goos, s.setupHome())
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

// handleInstances renders the backup instances as one card per repository.
func (s *Server) handleInstances(w http.ResponseWriter, r *http.Request) {
	v := webui.InstancesView{Chrome: s.chrome(r, "instances", "Instances")}
	if s.opts.Backend != nil {
		cfg, _, err := s.opts.Backend.Config()
		if err != nil {
			v.Errors = append(v.Errors, err.Error())
		}
		if cfg != nil {
			cards, err := webui.RenderInstanceCards(instanceCardViews(cfg))
			if err != nil {
				v.Errors = append(v.Errors, err.Error())
			}
			v.Cards = cards
		}
	}
	s.render(w, "instances_page", v)
}

// instanceCardViews maps the instance cards of cfg onto their view model.
func instanceCardViews(cfg *config.Config) []webui.InstanceCardView {
	cards := InstanceCards(cfg)
	out := make([]webui.InstanceCardView, len(cards))
	for i, c := range cards {
		v := webui.InstanceCardView{ID: c.ID, Type: c.Type, MountPoint: c.MountPoint}
		for _, f := range c.Fields {
			v.Controls = append(v.Controls, control(f, ""))
		}
		for _, root := range c.Roots {
			v.Roots = append(v.Roots, root.LocalPath)
		}
		out[i] = v
	}
	return out
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	v := webui.ConfigView{Chrome: s.chrome(r, "config", "Config"), Saved: r.URL.Query().Get("saved") == "1"}
	if s.opts.Backend != nil {
		cfg, rev, err := s.opts.Backend.Config()
		if err != nil {
			v.Errors = append(v.Errors, err.Error())
		}
		v.Revision = string(rev)
		if cfg != nil {
			v.FormSections = FormSections(cfg, nil, nil)
			fillConfigView(&v, cfg)
		}
	}
	s.render(w, "config", v)
}

// browse fills v's linked directories when p is empty, else p's snapshots
// newest first with id marked selected. Absent history for p leaves both
// empty.
func (s *Server) browse(v *webui.HistoryView, b snapshotBrowser, root, p string, id provider.SnapshotID) error {
	dir, err := s.resolve(root, p)
	if err != nil {
		return err
	}
	if p == "" {
		dirs, _ := b.LinkedDirs(root)
		for _, d := range dirs {
			v.LinkedDirs = append(v.LinkedDirs, webui.LinkedDir{Path: d, URL: historyURL(root, d)})
		}
		return nil
	}
	snaps, _ := b.Snapshots(root, dir)
	var warm map[provider.SnapshotID]bool
	if s.opts.Backend != nil {
		if st, ok := s.opts.Backend.Status().(status.Snapshot); ok {
			warm = st.Warm
		}
	}
	for _, sn := range snaps {
		v.Timeline = append(v.Timeline, webui.SnapshotTick{
			ID:       string(sn.ID),
			Time:     sn.Time.Format(time.RFC3339),
			Alias:    sn.Alias,
			Host:     sn.Host,
			URL:      historyURL(root, p) + "&snapshot=" + url.QueryEscape(string(sn.ID)),
			Warm:     warm[sn.ID],
			Selected: sn.ID == id,
		})
	}
	return nil
}

// historyURL is the history page for p in root, root first.
func historyURL(root, p string) string {
	return "/history?root=" + url.QueryEscape(root) + "&path=" + url.QueryEscape(p)
}

// rootRepoStates maps each configured root ID to its repository's state in
// the daemon status. It is empty without a Backend or a daemon status.
func (s *Server) rootRepoStates() map[string]string {
	out := map[string]string{}
	if s.opts.Backend == nil {
		return out
	}
	st, ok := s.opts.Backend.Status().(status.Snapshot)
	if !ok {
		return out
	}
	cfg, _, err := s.opts.Backend.Config()
	if err != nil || cfg == nil {
		return out
	}
	repos := map[string]string{}
	for _, repo := range st.Repos {
		repos[repo.ID] = repo.State
	}
	for _, root := range cfg.Roots {
		out[root.ID] = repos[root.RepositoryID]
	}
	return out
}

// handleHistory lists the roots and, when root and snapshot are given, the
// entries of path in that snapshot. html/template escapes every name.
func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	v := webui.HistoryView{Chrome: s.chrome(r, "history", "History")}
	if s.opts.History == nil {
		s.render(w, "history", v)
		return
	}
	repoStates := s.rootRepoStates()
	for _, root := range s.opts.History.Roots() {
		v.Roots = append(v.Roots, webui.RootItem{
			Name: root.ID, Path: root.Path, RepoState: repoStates[root.ID], MountState: root.State,
		})
	}
	q := r.URL.Query()
	root, id := q.Get("root"), provider.SnapshotID(q.Get("snapshot"))
	v.Root, v.Path = root, q.Get("path")
	if id != "" && !id.Valid() {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, errSnapshotID)
		return
	}
	if b, ok := s.opts.History.(snapshotBrowser); ok && root != "" {
		if err := s.browse(&v, b, root, q.Get("path"), id); err != nil {
			writeError(w, http.StatusBadRequest, errcode.InvalidConfig, err)
			return
		}
	}
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
