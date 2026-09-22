package readerpolicy

// Events returns a copy of the recorded throttle events, oldest first.
func (p *Policy) Events() []ThrottleEvent {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]ThrottleEvent(nil), p.events...)
}

// denied records a deny event the first time the process (pid, name) is denied.
func (p *Policy) denied(pid uint32, name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := trackerKey{pid: pid, name: name}
	el, ok := p.trackers[key]
	if !ok {
		el = p.track(key)
	} else {
		p.lru.MoveToFront(el)
	}
	p.transition(el.Value.(*tracker), true, "deny")
}

// transition records an event when tr moves into the blocked state. The caller
// must hold p.mu.
func (p *Policy) transition(tr *tracker, blocked bool, rule string) {
	if blocked && !tr.blocked {
		p.events = append(p.events, ThrottleEvent{PID: tr.key.pid, Process: tr.key.name, Rule: rule, At: p.now()})
		if len(p.events) > p.maxEvents() {
			p.events = p.events[1:]
		}
	}
	tr.blocked = blocked
}

func (p *Policy) maxEvents() int {
	if p.cfg.MaxEvents > 0 {
		return p.cfg.MaxEvents
	}
	return DefaultMaxEvents
}
