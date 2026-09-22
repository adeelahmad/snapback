package web

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/webui"
)

// lines splits a textarea value into its non-empty trimmed lines.
func lines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

// baseConfig returns a copy of the stored configuration that the form fields
// are applied to, so a failed save leaves the stored one untouched.
func (s *Server) baseConfig() (*config.Config, config.Revision, error) {
	if s.opts.Backend == nil {
		return nil, "", errors.New("no configuration backend")
	}
	cfg, rev, err := s.opts.Backend.Config()
	if err != nil {
		return nil, "", err
	}
	out := config.Config{}
	if cfg != nil {
		out = *cfg
	}
	out.Repositories = append([]config.Repository(nil), out.Repositories...)
	out.Roots = append([]config.Root(nil), out.Roots...)
	return &out, rev, nil
}

// applyRoots replaces cfg's roots with one per local path, reusing the root
// already at that position so its ID and repository survive an edit.
func applyRoots(cfg *config.Config, paths []string) {
	old := cfg.Roots
	roots := make([]config.Root, 0, len(paths))
	for i, p := range paths {
		var root config.Root
		switch {
		case i < len(old):
			root = old[i]
		default:
			root.ID = filepath.Base(p)
			if len(cfg.Repositories) > 0 {
				root.RepositoryID = cfg.Repositories[0].ID
			}
		}
		root.LocalPath = p
		roots = append(roots, root)
	}
	cfg.Roots = roots
}

// applyFilters sets the snapshot filter of every root from the Config form's
// "host=" and "tag=" lines. It is the reverse of fillConfigView.
func applyFilters(cfg *config.Config, filters []string) error {
	var host string
	var tags []string
	for _, f := range filters {
		switch key, val, _ := strings.Cut(f, "="); key {
		case "host":
			host = val
		case "tag":
			tags = append(tags, val)
		default:
			return fmt.Errorf("filter %q: want host=<name> or tag=<name>", f)
		}
	}
	for i := range cfg.Roots {
		cfg.Roots[i].Snapshots.Hostname = host
		cfg.Roots[i].Snapshots.TagsAll = tags
	}
	return nil
}

// applyConfigForm copies the Config form fields of r onto cfg, reporting the
// fields the user has to correct.
func applyConfigForm(cfg *config.Config, r *http.Request) []string {
	var errs []string
	applyRoots(cfg, lines(r.PostFormValue("roots")))
	if err := applyFilters(cfg, lines(r.PostFormValue("filters"))); err != nil {
		errs = append(errs, err.Error())
	}
	excl := lines(r.PostFormValue("exclusions"))
	seeds := lines(r.PostFormValue("seed_paths"))
	for i := range cfg.Roots {
		cfg.Roots[i].ExcludeRelativePaths = excl
		cfg.Roots[i].SeedPaths = nil
		for _, p := range seeds {
			cfg.Roots[i].SeedPaths = append(cfg.Roots[i].SeedPaths, config.SeedPath{Path: p})
		}
	}
	cfg.Discovery.Mode = r.PostFormValue("discovery_mode")
	if dir := strings.TrimSpace(r.PostFormValue("cache_dir")); len(cfg.Repositories) > 0 {
		cfg.Repositories[0].CacheDir = dir
	}
	switch raw := strings.TrimSpace(r.PostFormValue("refresh_interval")); raw {
	case "":
		cfg.Catalog.RefreshInterval = 0
	default:
		d, err := time.ParseDuration(raw)
		if err != nil {
			errs = append(errs, fmt.Sprintf("refresh interval %q: want a duration such as 15m", raw))
			break
		}
		cfg.Catalog.RefreshInterval = d
	}
	return errs
}

// configFormView re-renders the Config form with the values the user just
// submitted, so a rejected save does not lose their edits.
func (s *Server) configFormView(r *http.Request, errs []string) webui.ConfigView {
	return webui.ConfigView{
		Chrome:          s.chrome(r, "config", "Config"),
		Revision:        r.PostFormValue("revision"),
		Roots:           lines(r.PostFormValue("roots")),
		Filters:         lines(r.PostFormValue("filters")),
		Exclusions:      lines(r.PostFormValue("exclusions")),
		SeedPaths:       lines(r.PostFormValue("seed_paths")),
		DiscoveryMode:   r.PostFormValue("discovery_mode"),
		CacheDir:        r.PostFormValue("cache_dir"),
		RefreshInterval: r.PostFormValue("refresh_interval"),
		Errors:          errs,
	}
}

// handleConfigSave saves the Config form: it applies the fields to the stored
// configuration and writes it under the revision the form carried, which
// validates it through config.Save. It redirects on success and re-renders
// the form with the submitted values on any failure.
func (s *Server) handleConfigSave(w http.ResponseWriter, r *http.Request) {
	cfg, _, err := s.baseConfig()
	if err != nil {
		s.render(w, "config", s.configFormView(r, []string{err.Error()}))
		return
	}
	if errs := applyConfigForm(cfg, r); len(errs) > 0 {
		s.render(w, "config", s.configFormView(r, errs))
		return
	}
	config.ApplyDefaults(cfg)
	if _, err := s.opts.Backend.SaveConfig(cfg, config.Revision(r.PostFormValue("revision"))); err != nil {
		msg := err.Error()
		if errors.Is(err, config.ErrRevisionConflict) {
			msg = "the configuration file changed since this page was loaded; reload and apply your edits again"
		}
		s.render(w, "config", s.configFormView(r, []string{msg}))
		return
	}
	http.Redirect(w, r, "/config?saved=1", http.StatusSeeOther)
}

// setupFormView re-renders the Setup form with the submitted values.
func (s *Server) setupFormView(r *http.Request, errs []string) webui.SetupView {
	v := webui.SetupView{
		Chrome:         s.chrome(r, "setup", "Setup"),
		RepoURI:        r.PostFormValue("repo_uri"),
		CredentialFile: r.PostFormValue("credential_file"),
		Roots:          lines(r.PostFormValue("roots")),
		Errors:         errs,
	}
	v.ResticPaths = binaries([]string{r.PostFormValue("restic_path")}, "restic")
	v.RclonePaths = binaries([]string{r.PostFormValue("rclone_path")}, "rclone")
	return v
}

// handleSetupSave saves the Setup form: it applies the repository fields and
// the roots to the stored configuration, checks it with the setup validator
// and writes it, then sends the user to the Status page.
func (s *Server) handleSetupSave(w http.ResponseWriter, r *http.Request) {
	cfg, rev, err := s.baseConfig()
	if err != nil {
		s.render(w, "setup", s.setupFormView(r, []string{err.Error()}))
		return
	}
	if len(cfg.Repositories) == 0 {
		cfg.Repositories = []config.Repository{{ID: "main"}}
	}
	repo := &cfg.Repositories[0]
	repo.Repository = strings.TrimSpace(r.PostFormValue("repo_uri"))
	repo.ResticBinary = strings.TrimSpace(r.PostFormValue("restic_path"))
	repo.RcloneBinary = strings.TrimSpace(r.PostFormValue("rclone_path"))
	repo.PasswordFile = strings.TrimSpace(r.PostFormValue("credential_file"))
	applyRoots(cfg, lines(r.PostFormValue("roots")))
	config.ApplyDefaults(cfg)

	if s.opts.Validator != nil {
		if err := s.opts.Validator.Validate(r.Context(), cfg); err != nil {
			s.render(w, "setup", s.setupFormView(r, []string{err.Error()}))
			return
		}
	}
	if _, err := s.opts.Backend.SaveConfig(cfg, rev); err != nil {
		msg := err.Error()
		if errors.Is(err, config.ErrRevisionConflict) {
			msg = "the configuration file changed since this page was loaded; reload and apply your edits again"
		}
		s.render(w, "setup", s.setupFormView(r, []string{msg}))
		return
	}
	http.Redirect(w, r, "/status", http.StatusSeeOther)
}
