// agentic:shim

package resolver

// Latest is a RED compile shim; it deliberately reports no snapshot.
func Latest(_ []Eligible) (Eligible, bool) {
	return Eligible{}, false
}
