package readerpolicy

// tracked reports how many process trackers the LRU holds.
func (p *Policy) tracked() int {
	panic("SUB-AGENT-TODO: return the number of (pid, name) trackers in the LRU, read under the Policy mutex; Allow keeps a sliding-window readdir ring per tracker and evicts the least recently used tracker once MaxPIDs (default DefaultMaxPIDs) is exceeded")
}
