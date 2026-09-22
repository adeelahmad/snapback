// Package config holds the Snapback configuration types.
//
// SUB-AGENT-TODO: declare every contract and section field with its §12
// snake_case yaml tag (tasks.md § T1); add PrefixMapping, SnapshotFilter and
// the section structs; no functions or methods.
package config

// Config is the whole configuration file.
type Config struct{}

// Repository is one Restic repository.
type Repository struct{}

// Root is one local directory tree backed by a repository.
type Root struct{}

// SeedPath is one path that discovery seeds `.snapshot` links under.
type SeedPath struct{}

// Revision identifies the bytes of a stored configuration file.
type Revision string

// Catalog configures snapshot catalog refresh and caching.
type Catalog struct{}

// OnAccess configures on-access discovery.
type OnAccess struct{}
