// Package config holds the Snapback configuration types.
package config

import "time"

// Config is the whole configuration file.
type Config struct {
	Version         int          `yaml:"version"`
	LinkName        string       `yaml:"link_name"`
	Timestamps      string       `yaml:"timestamps"`
	StateDir        string       `yaml:"state_dir"`
	HistoryMount    string       `yaml:"history_mount"`
	BackendMountDir string       `yaml:"backend_mount_dir"`
	Web             Web          `yaml:"web"`
	Catalog         Catalog      `yaml:"catalog"`
	Views           Views        `yaml:"views"`
	Discovery       Discovery    `yaml:"discovery"`
	Repositories    []Repository `yaml:"repositories"`
	Roots           []Root       `yaml:"roots"`
	Service         Service      `yaml:"service"`
	Telemetry       Telemetry    `yaml:"telemetry,omitempty"`
}

// Telemetry records the answer to the setup opt-in question. It is off by
// default; the exporter that would send the counters is a later sprint.
type Telemetry struct {
	Enabled bool `yaml:"enabled"`
}

// Repository is one Restic repository.
type Repository struct {
	ID           string            `yaml:"id"`
	Repository   string            `yaml:"repository"`
	ResticBinary string            `yaml:"restic_binary"`
	RcloneBinary string            `yaml:"rclone_binary"`
	PasswordFile string            `yaml:"password_file"`
	CacheDir     string            `yaml:"cache_dir"`
	NoCache      bool              `yaml:"no_cache"`
	LockMode     string            `yaml:"lock_mode"`
	MountPoint   string            `yaml:"mount_point,omitempty"`
	Environment  map[string]string `yaml:"environment"`
}

// Root is one local directory tree backed by a repository.
type Root struct {
	ID                   string          `yaml:"id"`
	LocalPath            string          `yaml:"local_path"`
	RepositoryID         string          `yaml:"repository_id"`
	PrefixMap            []PrefixMapping `yaml:"prefix_map"`
	Snapshots            SnapshotFilter  `yaml:"snapshots"`
	SeedPaths            []SeedPath      `yaml:"seed_paths"`
	ExcludeRelativePaths []string        `yaml:"exclude_relative_paths"`
	Snap                 SnapSettings    `yaml:"snap"`
}

// PrefixMapping maps a snapshot source path on a host to a tree prefix.
type PrefixMapping struct {
	Hostname   string `yaml:"hostname"`
	SourcePath string `yaml:"source_path"`
	TreePrefix string `yaml:"tree_prefix"`
}

// SnapshotFilter selects which snapshots a root shows.
type SnapshotFilter struct {
	Hostname         string   `yaml:"hostname"`
	TagsAll          []string `yaml:"tags_all"`
	SourcePathsExact []string `yaml:"source_paths_exact"`
}

// SeedPath is one path that discovery seeds `.snapshot` links under.
type SeedPath struct {
	Path     string `yaml:"path"`
	MaxDepth int    `yaml:"max_depth"`
}

// SnapSettings configures ad-hoc snapshots of a root.
type SnapSettings struct {
	Tags []string `yaml:"tags"`
}

// Revision identifies the bytes of a stored configuration file.
type Revision string

// Web configures the local web UI.
type Web struct {
	Enabled        bool     `yaml:"enabled"`
	Listen         string   `yaml:"listen"`
	Bind           string   `yaml:"bind"`
	AllowedOrigins []string `yaml:"allowed_origins,omitempty"`
	OpenBrowser    bool     `yaml:"open_browser"`
	AssetsDir      string   `yaml:"assets_dir"`
}

// Catalog configures snapshot catalog refresh and caching.
type Catalog struct {
	RefreshInterval      time.Duration `yaml:"refresh_interval"`
	PrewarmSnapshots     int           `yaml:"prewarm_snapshots"`
	PrewarmConcurrency   int           `yaml:"prewarm_concurrency"`
	ProbeConcurrency     int           `yaml:"probe_concurrency"`
	PresenceCacheEntries int           `yaml:"presence_cache_entries"`
	PresenceCacheTTL     time.Duration `yaml:"presence_cache_ttl"`
	ReaderPolicy         ReaderPolicy  `yaml:"reader_policy"`
}

// ReaderPolicy throttles processes that enumerate snapshot directories.
type ReaderPolicy struct {
	DenyProcesses []string `yaml:"deny_processes"`
	BurstLimit    int      `yaml:"burst_limit"`
}

// Views configures alternative snapshot presentations.
type Views struct {
	Rsnapshot     bool          `yaml:"rsnapshot"`
	RsnapshotKeep RsnapshotKeep `yaml:"rsnapshot_keep"`
}

// RsnapshotKeep is how many snapshots each rsnapshot interval shows.
type RsnapshotKeep struct {
	Hourly  int `yaml:"hourly"`
	Daily   int `yaml:"daily"`
	Weekly  int `yaml:"weekly"`
	Monthly int `yaml:"monthly"`
}

// Discovery configures where `.snapshot` links appear.
type Discovery struct {
	Mode     string       `yaml:"mode"`
	Shell    bool         `yaml:"shell"`
	Seed     SeedSettings `yaml:"seed"`
	OnAccess OnAccess     `yaml:"on_access"`
}

// SeedSettings bounds seed-mode discovery.
type SeedSettings struct {
	InodeThreshold  float64 `yaml:"inode_threshold"`
	MaxLinksPerPath int     `yaml:"max_links_per_path"`
}

// OnAccess configures on-access discovery.
type OnAccess struct {
	AllowProcesses []string      `yaml:"allow_processes"`
	HandlerTimeout time.Duration `yaml:"handler_timeout"`
}

// Service configures how Snapback runs as a service.
type Service struct {
	Manager   string `yaml:"manager"`
	Scope     string `yaml:"scope"`
	RunAsUser string `yaml:"run_as_user"`
}
