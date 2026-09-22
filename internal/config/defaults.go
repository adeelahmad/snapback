package config

import (
	"errors"
	"os"
	"path/filepath"
	"time"
)

// defaults returns a Config pre-filled with every static default.
func defaults() *Config {
	return &Config{
		LinkName:   ".snapshot",
		Timestamps: "utc",
		Web:        Web{Enabled: true, Listen: "127.0.0.1:0"},
		Catalog: Catalog{
			RefreshInterval:      60 * time.Second,
			PrewarmSnapshots:     2,
			PrewarmConcurrency:   2,
			ProbeConcurrency:     4,
			PresenceCacheEntries: 10000,
			PresenceCacheTTL:     5 * time.Minute,
			ReaderPolicy: ReaderPolicy{
				DenyProcesses: []string{"rg", "fd", "find", "rsync", "mdworker", "mds"},
				BurstLimit:    50,
			},
		},
		Views: Views{RsnapshotKeep: RsnapshotKeep{Daily: 7, Weekly: 4, Monthly: 6}},
		Discovery: Discovery{
			Mode:     "seed",
			Shell:    true,
			Seed:     SeedSettings{InodeThreshold: 0.90, MaxLinksPerPath: 500000},
			OnAccess: OnAccess{HandlerTimeout: 20 * time.Millisecond},
		},
		Service: Service{Manager: "auto", Scope: "user"},
	}
}

// ApplyDefaults fills the defaults that depend on decoded values or the
// environment, so a configuration assembled field by field carries the same
// defaults as a parsed one.
func ApplyDefaults(c *Config) {
	if c.StateDir == "" {
		if xs := os.Getenv("XDG_STATE_HOME"); filepath.IsAbs(xs) {
			c.StateDir = filepath.Join(xs, "snapback")
		} else {
			c.StateDir = filepath.Join(os.Getenv("HOME"), ".local", "state", "snapback")
		}
	}
	if c.Web.Bind == "" {
		c.Web.Bind = c.Web.Listen
	}
	if c.HistoryMount == "" {
		c.HistoryMount = filepath.Join(c.StateDir, "mounts", "history")
	}
	if c.BackendMountDir == "" {
		c.BackendMountDir = filepath.Join(c.StateDir, "mounts", "repositories")
	}
	for i := range c.Repositories {
		if c.Repositories[i].LockMode == "" {
			c.Repositories[i].LockMode = "normal"
		}
	}
}

// Default returns the configuration used before any config file exists: the
// static and environment defaults with no repositories or roots.
func Default() *Config {
	c := defaults()
	c.Version = 1
	ApplyDefaults(c)
	return c
}

// DefaultPath returns the default config file path under the XDG config home.
func DefaultPath() (string, error) {
	if xc := os.Getenv("XDG_CONFIG_HOME"); filepath.IsAbs(xc) {
		return filepath.Join(xc, "snapback", "config.yaml"), nil
	}
	home := os.Getenv("HOME")
	if home == "" {
		return "", errors.New("config: neither XDG_CONFIG_HOME nor HOME is set")
	}
	return filepath.Join(home, ".config", "snapback", "config.yaml"), nil
}
