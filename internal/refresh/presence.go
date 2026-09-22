package refresh

import (
	"container/list"
	"sync"
	"time"
)

// PresenceCache remembers whether a path is present in a snapshot. It evicts
// the least recently used entry at its size limit, expires entries older than
// its TTL and is safe for concurrent use.
type PresenceCache struct {
	maxEntries int
	ttl        time.Duration
	now        func() time.Time

	mu    sync.Mutex
	order *list.List
	items map[string]*list.Element
}

type presenceEntry struct {
	key     string
	present bool
	stored  time.Time
}

// NewPresenceCache returns a cache holding at most maxEntries entries for ttl,
// using now as its clock. A zero maxEntries or ttl disables caching.
func NewPresenceCache(maxEntries int, ttl time.Duration, now func() time.Time) *PresenceCache {
	if now == nil {
		now = time.Now
	}
	return &PresenceCache{
		maxEntries: maxEntries,
		ttl:        ttl,
		now:        now,
		order:      list.New(),
		items:      make(map[string]*list.Element),
	}
}

func (c *PresenceCache) disabled() bool {
	return c.maxEntries <= 0 || c.ttl <= 0
}

// Get reports the cached presence for key and whether the cache held a live entry.
func (c *PresenceCache) Get(key string) (present, ok bool) {
	if c.disabled() {
		return false, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	el, found := c.items[key]
	if !found {
		return false, false
	}
	e := el.Value.(*presenceEntry)
	if c.now().Sub(e.stored) > c.ttl {
		c.order.Remove(el)
		delete(c.items, key)
		return false, false
	}
	c.order.MoveToFront(el)
	return e.present, true
}

// Put records the presence of key.
func (c *PresenceCache) Put(key string, present bool) {
	if c.disabled() {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	e := &presenceEntry{key: key, present: present, stored: c.now()}
	if el, found := c.items[key]; found {
		el.Value = e
		c.order.MoveToFront(el)
		return
	}
	c.items[key] = c.order.PushFront(e)
	for c.order.Len() > c.maxEntries {
		tail := c.order.Back()
		c.order.Remove(tail)
		delete(c.items, tail.Value.(*presenceEntry).key)
	}
}

// Clear drops every in-memory entry. It never touches Restic or rclone caches.
func (c *PresenceCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.order.Init()
	c.items = make(map[string]*list.Element)
}

// Len returns the number of cached entries.
func (c *PresenceCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}
