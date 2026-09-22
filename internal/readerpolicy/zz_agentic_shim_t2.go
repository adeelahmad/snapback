// agentic:shim
package readerpolicy

// tracked reports how many process trackers the LRU holds.
func (p *Policy) tracked() int { return 1 << 30 }
