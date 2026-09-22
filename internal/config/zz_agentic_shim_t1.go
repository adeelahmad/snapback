// agentic:shim

// Package config is a compile shim for S3-01 T1; the scaffolder replaces it
// with types.go. The shapes here are deliberately wrong.
package config

// Config is a shim with no fields.
type Config struct{}

// Repository is a shim with no fields.
type Repository struct{}

// Root is a shim with no fields.
type Root struct{}

// SeedPath is a shim with no fields.
type SeedPath struct{}

// Revision is a shim with the wrong kind.
type Revision int

// Catalog is a shim with no fields.
type Catalog struct{}

// OnAccess is a shim with no fields.
type OnAccess struct{}
