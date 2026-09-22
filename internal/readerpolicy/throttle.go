package readerpolicy

import (
	"container/list"
	"time"

	"github.com/adeelahmad/snapback/internal/mount"
)

// trackerKey identifies a process by PID and name so a reused PID starts fresh.
type trackerKey struct {
	pid  uint32
	name string
}

// tracker holds the recent readdir times of one process, oldest first.
type tracker struct {
	key     trackerKey
	hits    []time.Time
	blocked bool
}

// throttled records ev for the process (pid, name) and reports whether that
// process has exceeded the readdir burst limit inside the window.
func (p *Policy) throttled(ev Event, name string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := trackerKey{pid: ev.PID, name: name}
	isReadDir := ev.Op == mount.OpReadDir
	el, ok := p.trackers[key]
	if !ok {
		if !isReadDir {
			return false
		}
		el = p.track(key)
	} else {
		p.lru.MoveToFront(el)
	}

	tr := el.Value.(*tracker)
	now := p.now()
	cutoff := now.Add(-p.window())
	i := 0
	for i < len(tr.hits) && !tr.hits[i].After(cutoff) {
		i++
	}
	tr.hits = tr.hits[i:]
	if isReadDir {
		tr.hits = append(tr.hits, now)
		if len(tr.hits) > p.cfg.BurstLimit+1 {
			tr.hits = tr.hits[1:]
		}
	}
	over := len(tr.hits) > p.cfg.BurstLimit
	p.transition(tr, over, "burst")
	return over
}

// track adds a tracker for key, evicting the least recently used one past
// MaxPIDs. The caller must hold p.mu.
func (p *Policy) track(key trackerKey) *list.Element {
	el := p.lru.PushFront(&tracker{key: key})
	p.trackers[key] = el
	if p.lru.Len() > p.maxPIDs() {
		oldest := p.lru.Back()
		p.lru.Remove(oldest)
		delete(p.trackers, oldest.Value.(*tracker).key)
	}
	return el
}

func (p *Policy) window() time.Duration {
	if p.cfg.Window > 0 {
		return p.cfg.Window
	}
	return DefaultWindow
}

func (p *Policy) maxPIDs() int {
	if p.cfg.MaxPIDs > 0 {
		return p.cfg.MaxPIDs
	}
	return DefaultMaxPIDs
}

// tracked reports how many process trackers the LRU holds.
func (p *Policy) tracked() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lru.Len()
}
