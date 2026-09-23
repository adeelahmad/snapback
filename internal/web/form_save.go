package web

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strings"

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

// configFormValues drops the form keys that are not configuration keys, so
// Decode does not report the CSRF token, the revision or a typed password as
// unknown fields.
func configFormValues(v url.Values) url.Values {
	out := make(url.Values, len(v))
	for name, vals := range v {
		switch {
		case name == csrfField, name == "revision":
		case strings.HasSuffix(name, ".password"), strings.HasSuffix(name, ".password_mode"):
		default:
			out[name] = vals
		}
	}
	return out
}

// applyCredentials writes the password typed for each repository to the
// credential store and points that repository at the stored file, so the
// secret itself never reaches the configuration.
func (s *Server) applyCredentials(cfg *config.Config, form url.Values) error {
	for i := range cfg.Repositories {
		if credentialMode(form, i) != "typed" {
			continue
		}
		secret := form.Get(fmt.Sprintf("repositories[%d].password", i))
		path, err := storeCredential(s.opts.StateDir, cfg.Repositories[i].ID, secret)
		if err != nil {
			return err
		}
		if path != "" {
			cfg.Repositories[i].PasswordFile = path
		}
	}
	return nil
}

// control turns one field into its control, keeping a select value the
// options do not list so a rejected choice stays visible on its control.
func control(f webui.Field, msg string) webui.Control {
	c := webui.Control{
		Kind:    f.Kind,
		Path:    f.Path,
		Label:   f.Label,
		Help:    f.Help,
		Value:   f.Value,
		Options: f.Options,
		Error:   msg,
	}
	switch f.Kind {
	case webui.KindChips:
		c.Options = nil
		for _, v := range f.Values {
			c.Options = append(c.Options, webui.Option{Value: v, Label: v})
		}
	case webui.KindSelect:
		listed := func(o webui.Option) bool { return o.Value == f.Value }
		if f.Value != "" && !slices.ContainsFunc(c.Options, listed) {
			c.Options = append(slices.Clone(c.Options), webui.Option{Value: f.Value, Label: f.Value})
		}
	}
	return c
}

// setupRepoPrefix is the field prefix the Setup form's single repository
// supplies its password under, matching the Config form's key paths.
const setupRepoPrefix = "repositories[0]"

// passwordControls are the two controls a repository supplies its password
// with: the mode, and the secret itself, which is never echoed back.
func passwordControls(prefix string, byPath map[string]string, form url.Values) []webui.Control {
	mode := form.Get(prefix + ".password_mode")
	if mode != "typed" {
		mode = "file"
	}
	return []webui.Control{{
		Kind:  webui.KindSelect,
		Path:  prefix + ".password_mode",
		Label: "Password mode",
		Value: mode,
		Options: []webui.Option{
			{Value: "file", Label: "Password file"},
			{Value: "typed", Label: "Typed password"},
		},
		Error: byPath[prefix+".password_mode"],
	}, {
		Kind:  webui.KindPassword,
		Path:  prefix + ".password",
		Label: "Password",
		Error: byPath[prefix+".password"],
	}}
}

// renderConfigForm re-renders the Config form under status with the values
// the user submitted, each error anchored to its own control.
func (s *Server) renderConfigForm(w http.ResponseWriter, r *http.Request, cfg *config.Config, byPath map[string]string, banner []string, status int) {
	v := webui.ConfigView{
		Chrome:   s.chrome(r, "config", "Config"),
		Revision: r.PostFormValue("revision"),
		Errors:   banner,
	}
	if cfg != nil {
		v.FormSections = FormSections(cfg, byPath, r.PostForm)
		telemetry, err := TelemetryPanel(cfg.Telemetry.Enabled)
		if err != nil {
			v.Errors = append(v.Errors, err.Error())
		}
		v.Telemetry = telemetry
	}
	w.WriteHeader(status)
	s.render(w, "config", v)
}

// handleConfigSave saves the Config form: every field is named after its YAML
// key path, so the whole configuration is decoded from the form, written
// under the revision the form carried and validated by config.Save. It
// redirects on success and re-renders the submitted values on any failure.
func (s *Server) handleConfigSave(w http.ResponseWriter, r *http.Request) {
	cfg, decoded := Decode(configFormValues(r.PostForm))
	byPath := make(map[string]string, len(decoded))
	for _, f := range decoded {
		if _, seen := byPath[f.Path]; !seen {
			byPath[f.Path] = f.Msg
		}
	}
	var banner []string
	if err := s.applyCredentials(cfg, r.PostForm); err != nil {
		banner = append(banner, err.Error())
	}
	config.ApplyDefaults(cfg)
	if len(byPath) > 0 || len(banner) > 0 {
		s.renderConfigForm(w, r, cfg, byPath, banner, http.StatusBadRequest)
		return
	}
	if _, err := s.opts.Backend.SaveConfig(cfg, config.Revision(r.PostFormValue("revision"))); err != nil {
		byPath, banner = fieldErrors(err)
		if errors.Is(err, config.ErrRevisionConflict) {
			banner = []string{"the configuration file changed since this page was loaded; reload and apply your edits again"}
		}
		s.renderConfigForm(w, r, cfg, byPath, banner, http.StatusBadRequest)
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
	v.MountPoint = mountPointField(nil, goos, s.setupHome())
	v.MountPoint.Value = strings.TrimSpace(r.PostFormValue(mountPointKey))
	v.PasswordControls = passwordControls(setupRepoPrefix, nil, r.PostForm)
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
	repo.MountPoint = strings.TrimSpace(r.PostFormValue(mountPointKey))
	applyRoots(cfg, lines(r.PostFormValue("roots")))
	config.ApplyDefaults(cfg)
	if err := s.applyCredentials(cfg, r.PostForm); err != nil {
		s.render(w, "setup", s.setupFormView(r, []string{err.Error()}))
		return
	}

	if msg := mountPointError(cfg); msg != "" {
		v := s.setupFormView(r, nil)
		v.MountPoint.Error = msg
		w.WriteHeader(http.StatusBadRequest)
		s.render(w, "setup", v)
		return
	}
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
