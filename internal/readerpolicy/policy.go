// Package readerpolicy decides which processes may read the history mount.
//
// It is resource protection and UX, not an access-control boundary: a
// same-user process can rename itself or reach the history another way.
package readerpolicy

import (
	"container/list"
	"path/filepath"
	"sync"
	"time"

	"github.com/adeelahmad/snapback/internal/mount"
)

// Defaults applied when the matching Config field is zero.
const (
	DefaultWindow    = 10 * time.Second
	DefaultMaxPIDs   = 1024
	DefaultMaxEvents = 128
)

// commLen is the length of a Linux process comm name (TASK_COMM_LEN-1).
const commLen = 15

// Config holds the reader policy settings from catalog.reader_policy.
type Config struct {
	Deny       []string
	BurstLimit int
	Window     time.Duration
	MaxPIDs    int
	MaxEvents  int
}

// Event is one filesystem request, field-for-field identical to the planned
// mount.Event so a consumer can convert between them.
type Event struct {
	Op   mount.Op
	Path string
	PID  uint32
}

// ThrottleEvent records a process's transition into being denied or throttled.
type ThrottleEvent struct {
	PID     uint32
	Process string
	Rule    string
	At      time.Time
}

// Policy applies a Config to incoming events.
type Policy struct {
	cfg      Config
	procName func(uint32) string
	now      func() time.Time

	mu       sync.Mutex
	trackers map[trackerKey]*list.Element
	lru      *list.List
	events   []ThrottleEvent
}

// New returns a Policy for cfg that resolves process names with procName and
// reads the time from now.
func New(cfg Config, procName func(uint32) string, now func() time.Time) *Policy {
	return &Policy{
		cfg:      cfg,
		procName: procName,
		now:      now,
		trackers: make(map[trackerKey]*list.Element),
		lru:      list.New(),
	}
}

// Allow reports whether ev may proceed.
func (p *Policy) Allow(ev Event) bool {
	name := p.procName(ev.PID)
	if name == "" {
		return true
	}
	name = filepath.Base(name)
	for _, d := range p.cfg.Deny {
		if name == d || (len(d) > commLen && name == d[:commLen]) {
			p.denied(ev.PID, name)
			return false
		}
	}
	if p.cfg.BurstLimit <= 0 {
		return true
	}
	return !p.throttled(ev, name)
}
