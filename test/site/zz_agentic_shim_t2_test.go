// agentic:shim

package site_test

// contrastRatio returns the WCAG 2.x contrast ratio of fg on bg.
// Shim body is deliberately wrong so the contrast tests fail by assertion.
func contrastRatio(fg, bg [3]uint8) float64 {
	_, _ = fg, bg
	return 0
}
