package resolver

// Latest returns the newest eligible snapshot, which is the head of an
// EligibleFor result, and false when there is none.
func Latest(eligible []Eligible) (Eligible, bool) {
	if len(eligible) == 0 {
		return Eligible{}, false
	}
	return eligible[0], true
}
