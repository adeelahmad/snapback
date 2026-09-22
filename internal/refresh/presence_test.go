package refresh

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"testing"
	"time"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func checkGet(t *testing.T, c *PresenceCache, key string, wantPresent, wantOK bool) {
	t.Helper()
	gotPresent, gotOK := c.Get(key)
	if gotPresent != wantPresent || gotOK != wantOK {
		t.Errorf("Get(%q) = (%v, %v), want (%v, %v)", key, gotPresent, gotOK, wantPresent, wantOK)
	}
}

func TestPresencePutGetAndTTL(t *testing.T) {
	clk := newFakeClock()
	c := NewPresenceCache(4, time.Minute, clk.Now)

	c.Put("a", true)
	c.Put("b", false)
	checkGet(t, c, "a", true, true)
	checkGet(t, c, "b", false, true)

	clk.Advance(61 * time.Second)
	checkGet(t, c, "a", false, false)
	if got, want := c.Len(), 1; got != want {
		t.Errorf("Len() after expired Get = %d, want %d", got, want)
	}
}

func TestPresenceBoundedLRU(t *testing.T) {
	clk := newFakeClock()
	c := NewPresenceCache(2, time.Minute, clk.Now)

	c.Put("a", true)
	c.Put("b", true)
	checkGet(t, c, "a", true, true)
	c.Put("c", true)

	if got, want := c.Len(), 2; got != want {
		t.Errorf("Len() = %d, want %d", got, want)
	}
	checkGet(t, c, "b", false, false)
	checkGet(t, c, "a", true, true)
	checkGet(t, c, "c", true, true)

	for i := range 1000 {
		c.Put(fmt.Sprintf("k%d", i), i%2 == 0)
		if got := c.Len(); got < 0 || got > 2 {
			t.Fatalf("Len() after put %d = %d, want 0..2", i, got)
		}
	}
}

func TestPresenceDisabledWhenZero(t *testing.T) {
	tests := []struct {
		name       string
		maxEntries int
		ttl        time.Duration
	}{
		{name: "zero entries", maxEntries: 0, ttl: time.Minute},
		{name: "zero ttl", maxEntries: 4, ttl: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clk := newFakeClock()
			c := NewPresenceCache(tt.maxEntries, tt.ttl, clk.Now)

			c.Put("a", true)
			checkGet(t, c, "a", false, false)
			if got, want := c.Len(), 0; got != want {
				t.Errorf("Len() = %d, want %d", got, want)
			}
		})
	}
}

func TestPresenceClearIsLocalOnly(t *testing.T) {
	dir := t.TempDir()
	cacheFile := filepath.Join(dir, "cache", "x")
	if err := os.MkdirAll(filepath.Dir(cacheFile), 0o755); err != nil {
		t.Fatal(err)
	}
	want := []byte("restic cache blob\n")
	if err := os.WriteFile(cacheFile, want, 0o644); err != nil {
		t.Fatal(err)
	}

	clk := newFakeClock()
	c := NewPresenceCache(4, time.Minute, clk.Now)
	keys := []string{"a", "b", "c"}
	for _, k := range keys {
		c.Put(k, true)
	}

	c.Clear()

	if got, want := c.Len(), 0; got != want {
		t.Errorf("Len() after Clear = %d, want %d", got, want)
	}
	for _, k := range keys {
		checkGet(t, c, k, false, false)
	}
	got, err := os.ReadFile(cacheFile)
	if err != nil {
		t.Fatalf("cache file after Clear: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("cache file after Clear = %q, want %q", got, want)
	}

	src, err := os.ReadFile("presence.go")
	if err != nil {
		t.Fatalf("read presence.go: %v", err)
	}
	if len(src) == 0 {
		t.Fatal("presence.go is empty")
	}
	if loc := regexp.MustCompile(`\bos\.`).FindIndex(src); loc != nil {
		t.Errorf("presence.go contains an os. call at byte %d, want none", loc[0])
	}
}
