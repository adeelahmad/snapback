package site_test

import "math"

// contrastRatio returns the WCAG 2.x contrast ratio of fg on bg.
func contrastRatio(fg, bg [3]uint8) float64 {
	l1, l2 := relativeLuminance(fg), relativeLuminance(bg)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// relativeLuminance is the WCAG 2.x relative luminance of an sRGB colour.
func relativeLuminance(c [3]uint8) float64 {
	weights := [3]float64{0.2126, 0.7152, 0.0722}
	var l float64
	for i, v := range c {
		s := float64(v) / 255
		if s <= 0.04045 {
			s /= 12.92
		} else {
			s = math.Pow((s+0.055)/1.055, 2.4)
		}
		l += weights[i] * s
	}
	return l
}
