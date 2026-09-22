package refresh

import "time"

// PresenceCache remembers whether a path is present in a snapshot. It evicts
// the least recently used entry at its size limit, expires entries older than
// its TTL and is safe for concurrent use.
type PresenceCache struct{}

// NewPresenceCache returns a cache holding at most maxEntries entries for ttl,
// using now as its clock. A zero maxEntries or ttl disables caching.
func NewPresenceCache(maxEntries int, ttl time.Duration, now func() time.Time) *PresenceCache {
	panic("SUB-AGENT-TODO: store maxEntries, ttl, now (default time.Now); init LRU list + map under a mutex")
}

// Get reports the cached presence for key and whether the cache held a live entry.
func (c *PresenceCache) Get(key string) (present, ok bool) {
	panic("SUB-AGENT-TODO: disabled -> miss; lookup; entry older than ttl -> drop and miss; else move to front and return")
}

// Put records the presence of key.
func (c *PresenceCache) Put(key string, present bool) {
	panic("SUB-AGENT-TODO: disabled -> no-op; upsert with timestamp at front; evict LRU tail beyond maxEntries")
}

// Clear drops every in-memory entry. It never touches Restic or rclone caches.
func (c *PresenceCache) Clear() {
	panic("SUB-AGENT-TODO: reset the list and map under the mutex")
}

// Len returns the number of cached entries.
func (c *PresenceCache) Len() int {
	panic("SUB-AGENT-TODO: return map length under the mutex")
}
