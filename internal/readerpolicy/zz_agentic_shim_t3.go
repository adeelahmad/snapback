// agentic:shim

package readerpolicy

// Events is a compile shim with a deliberately wrong body; GREEN replaces it.
func (p *Policy) Events() []ThrottleEvent {
	return nil
}
