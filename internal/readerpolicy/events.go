package readerpolicy

// Events returns a copy of the recorded throttle events, oldest first.
func (p *Policy) Events() []ThrottleEvent {
	panic("SUB-AGENT-TODO: record a ThrottleEvent only on the transition into deny/burst; keep a MaxEvents ring (DefaultMaxEvents when 0) that drops the oldest; return a copy")
}
