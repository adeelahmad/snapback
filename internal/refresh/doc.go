// Package refresh keeps Snapback's view of Restic snapshots current: it
// reconciles listed snapshots with the backend mount, publishes new
// generations and holds a small in-memory presence cache.
package refresh
