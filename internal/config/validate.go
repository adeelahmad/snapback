package config

import (
	"fmt"
	"net"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/adeelahmad/snapback/internal/errcode"
)

var (
	idPattern     = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)
	envKeyPattern = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)
)

// FieldError is one invalid field. Path is the YAML path, such as
// repositories[0].password_file. Code is empty for ordinary errors and names
// a specific errcode for the v0.1 rejections.
type FieldError struct {
	Path string
	Msg  string
	Code errcode.Code
}

// ValidationError collects every invalid field found by Validate.
type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for _, f := range e.Fields {
		parts = append(parts, f.Path+": "+f.Msg)
	}
	return strings.Join(parts, "; ")
}

func (e *ValidationError) Unwrap() error {
	return errcode.New(errcode.InvalidConfig, "config.validate", nil)
}

// validator accumulates field errors.
type validator struct {
	errs []FieldError
}

func (v *validator) add(path, msg string) {
	v.errs = append(v.errs, FieldError{Path: path, Msg: msg})
}

// absPath requires an absolute, clean path, which also rules out "~".
func (v *validator) absPath(path, p string) {
	if !filepath.IsAbs(p) || filepath.Clean(p) != p {
		v.add(path, "must be an absolute, clean path")
	}
}

// relPath requires a relative, clean path free of "..".
func (v *validator) relPath(path, p string) {
	if filepath.IsAbs(p) || filepath.Clean(p) != p || p == "." ||
		slices.Contains(strings.Split(p, string(filepath.Separator)), "..") {
		v.add(path, "must be a relative, clean path without ..")
	}
}

func (v *validator) id(path, id string) {
	if !idPattern.MatchString(id) {
		v.add(path, "must match ^[a-z][a-z0-9_-]{0,31}$")
	}
}

func (v *validator) atLeast(path string, n, floor int) {
	if n < floor {
		v.add(path, fmt.Sprintf("must be >= %d", floor))
	}
}

// Validate checks every field rule plus the topology and credential rules and
// returns a *ValidationError listing all failures, or nil.
func Validate(c *Config) error {
	v := &validator{}
	v.top(c)
	v.catalog(c)
	v.discovery(c)
	v.service(c)
	v.repositories(c)
	v.roots(c)
	// An empty mount is already reported above; topology on it would compare
	// against "." and add noise.
	if c.HistoryMount != "" && c.BackendMountDir != "" {
		v.errs = append(v.errs, checkTopology(c)...)
	}
	v.errs = append(v.errs, checkCredentials(c)...)
	if len(v.errs) > 0 {
		return &ValidationError{Fields: v.errs}
	}
	return nil
}

func (v *validator) top(c *Config) {
	if c.Version != 1 {
		v.add("version", "must be 1")
	}
	if c.LinkName == "" || c.LinkName == "." || c.LinkName == ".." || strings.ContainsAny(c.LinkName, "/\x00") {
		v.add("link_name", "must be a non-empty file name other than . or ..")
	}
	if c.Timestamps != "utc" && c.Timestamps != "local" {
		v.add("timestamps", "must be utc or local")
	}
	v.absPath("state_dir", c.StateDir)
	v.absPath("history_mount", c.HistoryMount)
	v.absPath("backend_mount_dir", c.BackendMountDir)
	host, port, err := net.SplitHostPort(c.Web.Listen)
	if err == nil {
		_, err = strconv.ParseUint(port, 10, 16)
	}
	if err != nil || (host != "127.0.0.1" && host != "::1" && host != "localhost") {
		v.add("web.listen", "must be host:port on 127.0.0.1, ::1 or localhost")
	}
	keep := c.Views.RsnapshotKeep
	v.atLeast("views.rsnapshot_keep.hourly", keep.Hourly, 0)
	v.atLeast("views.rsnapshot_keep.daily", keep.Daily, 0)
	v.atLeast("views.rsnapshot_keep.weekly", keep.Weekly, 0)
	v.atLeast("views.rsnapshot_keep.monthly", keep.Monthly, 0)
}

func (v *validator) catalog(c *Config) {
	cat := c.Catalog
	if cat.RefreshInterval <= 0 {
		v.add("catalog.refresh_interval", "must be > 0")
	}
	v.atLeast("catalog.prewarm_snapshots", cat.PrewarmSnapshots, 0)
	v.atLeast("catalog.prewarm_concurrency", cat.PrewarmConcurrency, 1)
	v.atLeast("catalog.probe_concurrency", cat.ProbeConcurrency, 1)
	v.atLeast("catalog.presence_cache_entries", cat.PresenceCacheEntries, 0)
	if cat.PresenceCacheTTL < 0 {
		v.add("catalog.presence_cache_ttl", "must be >= 0")
	}
	v.atLeast("catalog.reader_policy.burst_limit", cat.ReaderPolicy.BurstLimit, 1)
}

func (v *validator) discovery(c *Config) {
	switch c.Discovery.Mode {
	case "seed":
	case "on-access":
		v.errs = append(v.errs, FieldError{Path: "discovery.mode", Msg: "on-access is not available in v0.1", Code: errcode.OnAccessUnavailable})
	default:
		v.add("discovery.mode", "must be seed")
	}
	if t := c.Discovery.Seed.InodeThreshold; t <= 0 || t > 1 {
		v.add("discovery.seed.inode_threshold", "must be > 0 and <= 1")
	}
	v.atLeast("discovery.seed.max_links_per_path", c.Discovery.Seed.MaxLinksPerPath, 1)
}

func (v *validator) service(c *Config) {
	if c.Service.Scope != "user" && c.Service.Scope != "system" {
		v.add("service.scope", "must be user or system")
	}
	switch c.Service.Manager {
	case "auto", "systemd":
	case "launchd", "openrc":
		v.errs = append(v.errs, FieldError{Path: "service.manager", Msg: c.Service.Manager + " is not supported in v0.1", Code: errcode.UnsupportedServiceManager})
	default:
		v.add("service.manager", "must be auto or systemd")
	}
}

func (v *validator) repositories(c *Config) {
	if len(c.Repositories) == 0 {
		v.add("repositories", "must list at least one repository")
	}
	seen := map[string]bool{}
	for i, r := range c.Repositories {
		p := fmt.Sprintf("repositories[%d]", i)
		v.id(p+".id", r.ID)
		if seen[r.ID] {
			v.add(p+".id", "duplicate repository id "+r.ID)
		}
		seen[r.ID] = true
		if r.Repository == "" {
			v.add(p+".repository", "must be set")
		}
		v.absPath(p+".restic_binary", r.ResticBinary)
		if r.RcloneBinary != "" || strings.HasPrefix(r.Repository, "rclone:") {
			v.absPath(p+".rclone_binary", r.RcloneBinary)
		}
		v.absPath(p+".password_file", r.PasswordFile)
		if r.CacheDir != "" {
			v.absPath(p+".cache_dir", r.CacheDir)
			if r.NoCache {
				v.add(p+".no_cache", "cache_dir and no_cache are mutually exclusive")
			}
		}
		if r.LockMode != "normal" && r.LockMode != "none" {
			v.add(p+".lock_mode", "must be normal or none")
		}
		keys := make([]string, 0, len(r.Environment))
		for k := range r.Environment {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for _, k := range keys {
			kp := p + ".environment." + k
			switch {
			case !envKeyPattern.MatchString(k):
				v.add(kp, "key must match ^[A-Z_][A-Z0-9_]*$")
			case strings.HasPrefix(k, "RESTIC_PASSWORD"):
				v.add(kp, "use password_file")
			}
		}
	}
}

func (v *validator) roots(c *Config) {
	if len(c.Roots) == 0 {
		v.add("roots", "must list at least one root")
	}
	repos := map[string]bool{}
	for _, r := range c.Repositories {
		repos[r.ID] = true
	}
	seen := map[string]bool{}
	for i, r := range c.Roots {
		p := fmt.Sprintf("roots[%d]", i)
		v.id(p+".id", r.ID)
		if seen[r.ID] {
			v.add(p+".id", "duplicate root id "+r.ID)
		}
		seen[r.ID] = true
		v.absPath(p+".local_path", r.LocalPath)
		if !repos[r.RepositoryID] {
			v.add(p+".repository_id", "names no repository")
		}
		for j, m := range r.PrefixMap {
			mp := fmt.Sprintf("%s.prefix_map[%d]", p, j)
			v.absPath(mp+".source_path", m.SourcePath)
			v.absPath(mp+".tree_prefix", m.TreePrefix)
		}
		for j, s := range r.Snapshots.SourcePathsExact {
			v.absPath(fmt.Sprintf("%s.snapshots.source_paths_exact[%d]", p, j), s)
		}
		for j, s := range r.SeedPaths {
			sp := fmt.Sprintf("%s.seed_paths[%d]", p, j)
			v.relPath(sp+".path", s.Path)
			v.atLeast(sp+".max_depth", s.MaxDepth, 1)
		}
		for j, e := range r.ExcludeRelativePaths {
			v.relPath(fmt.Sprintf("%s.exclude_relative_paths[%d]", p, j), e)
		}
	}
}
