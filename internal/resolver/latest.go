package resolver

// Latest returns the newest eligible snapshot, which is the head of an
// EligibleFor result, and false when there is none.
func Latest(eligible []Eligible) (Eligible, bool) {
	panic("SUB-AGENT-TODO: T6 Latest — return (eligible[0], true) when len(eligible) > 0, else (Eligible{}, false); EligibleFor already sorts newest first, so never re-sort, fall back, probe or consult anything beyond the slice head")
}
