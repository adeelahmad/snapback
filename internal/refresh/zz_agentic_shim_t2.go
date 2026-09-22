// agentic:shim

package refresh

import "time"

// PresenceCache is a compile shim for S3-11 T2.
type PresenceCache struct{}

// NewPresenceCache is a compile shim for S3-11 T2.
func NewPresenceCache(maxEntries int, ttl time.Duration, now func() time.Time) *PresenceCache {
	return &PresenceCache{}
}

// Get is a compile shim for S3-11 T2.
func (c *PresenceCache) Get(key string) (present, ok bool) { return true, true }

// Put is a compile shim for S3-11 T2.
func (c *PresenceCache) Put(key string, present bool) {}

// Clear is a compile shim for S3-11 T2.
func (c *PresenceCache) Clear() {}

// Len is a compile shim for S3-11 T2.
func (c *PresenceCache) Len() int { return -1 }
