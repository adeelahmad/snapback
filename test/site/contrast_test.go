package site_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"testing"
)

// designTokens is the subset of web/tokens.json the contrast tests inspect.
type designTokens struct {
	Name  string `json:"name"`
	Color struct {
		Themes []struct {
			ID string `json:"id"`
		} `json:"themes"`
		Tokens []struct {
			Name  string          `json:"name"`
			Value json.RawMessage `json:"value"`
		} `json:"tokens"`
	} `json:"color"`
}

var themes = []string{"light", "dark"}

// contrastPair is a foreground token drawn on a background token.
type contrastPair struct {
	fg, bg string
}

var textPairs = []contrastPair{
	{"ink", "canvas"},
	{"ink", "surface"},
	{"ink-body", "canvas"},
	{"ink-body", "surface"},
	{"ink-body", "canvas-sunken"},
	{"ink-muted", "canvas"},
	{"ink-muted", "surface"},
	{"accent-text", "canvas"},
	{"accent-text", "surface"},
	{"on-accent", "accent"},
}

var nonTextPairs = []contrastPair{
	{"line-strong", "canvas"},
}

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// colorTable maps theme -> token name -> raw value. A token whose value is a
// plain string applies to every theme.
func colorTable(t *testing.T, tok designTokens) map[string]map[string]string {
	t.Helper()
	table := map[string]map[string]string{}
	for _, th := range themes {
		table[th] = map[string]string{}
	}
	for _, c := range tok.Color.Tokens {
		var single string
		if err := json.Unmarshal(c.Value, &single); err == nil {
			for _, th := range themes {
				table[th][c.Name] = single
			}
			continue
		}
		var perTheme map[string]string
		if err := json.Unmarshal(c.Value, &perTheme); err != nil {
			t.Fatalf("colour token %q: value is neither a string nor a theme map: %v", c.Name, err)
		}
		for _, th := range themes {
			table[th][c.Name] = perTheme[th]
		}
	}
	return table
}

func loadColorTable(t *testing.T) map[string]map[string]string {
	t.Helper()
	var tok designTokens
	loadJSON(t, "web/tokens.json", &tok)
	return colorTable(t, tok)
}

// parseHex decodes a #rrggbb colour into 0-255 channels.
func parseHex(s string) ([3]uint8, error) {
	var rgb [3]uint8
	if !hexColor.MatchString(s) {
		return rgb, fmt.Errorf("%q is not #rrggbb", s)
	}
	for i := range 3 {
		v, err := strconv.ParseUint(s[1+2*i:3+2*i], 16, 8)
		if err != nil {
			return rgb, fmt.Errorf("parse %q: %w", s, err)
		}
		rgb[i] = uint8(v)
	}
	return rgb, nil
}

func checkPairs(t *testing.T, pairs []contrastPair, minRatio float64) {
	t.Helper()
	table := loadColorTable(t)
	for _, th := range themes {
		for _, p := range pairs {
			fg, err := parseHex(table[th][p.fg])
			if err != nil {
				t.Errorf("%s/%s in %s: fg: %v", p.fg, p.bg, th, err)
				continue
			}
			bg, err := parseHex(table[th][p.bg])
			if err != nil {
				t.Errorf("%s/%s in %s: bg: %v", p.fg, p.bg, th, err)
				continue
			}
			if got := contrastRatio(fg, bg); got < minRatio {
				t.Errorf("contrastRatio(%s, %s) in %s = %.2f, want >= %.2f", p.fg, p.bg, th, got, minRatio)
			}
		}
	}
}

func TestTextContrastAA(t *testing.T) {
	checkPairs(t, textPairs, 4.5)
}

func TestNonTextContrast(t *testing.T) {
	checkPairs(t, nonTextPairs, 3.0)
}

func TestContrastRatioKnownValues(t *testing.T) {
	tests := []struct {
		fg, bg   string
		min, max float64
	}{
		{"#000000", "#ffffff", 20.995, 21.005},
		{"#ffffff", "#ffffff", 0.995, 1.005},
		{"#767676", "#ffffff", 4.54, 4.55},
	}
	for _, tt := range tests {
		fg, err := parseHex(tt.fg)
		if err != nil {
			t.Fatalf("parseHex(%q): %v", tt.fg, err)
		}
		bg, err := parseHex(tt.bg)
		if err != nil {
			t.Fatalf("parseHex(%q): %v", tt.bg, err)
		}
		got := contrastRatio(fg, bg)
		if got < tt.min || got >= tt.max {
			t.Errorf("contrastRatio(%s, %s) = %.4f, want in [%.3f, %.3f)", tt.fg, tt.bg, got, tt.min, tt.max)
		}
	}
}

func TestTokenFileShape(t *testing.T) {
	var tok designTokens
	loadJSON(t, "web/tokens.json", &tok)

	if tok.Name != "snapback" {
		t.Errorf("tokens name = %q, want %q", tok.Name, "snapback")
	}
	var gotThemes []string
	for _, th := range tok.Color.Themes {
		gotThemes = append(gotThemes, th.ID)
	}
	if !slices.Equal(gotThemes, themes) {
		t.Errorf("tokens themes = %v, want %v", gotThemes, themes)
	}

	table := colorTable(t, tok)
	var referenced []string
	for _, p := range slices.Concat(textPairs, nonTextPairs) {
		referenced = append(referenced, p.fg, p.bg)
	}
	for _, th := range themes {
		for _, name := range referenced {
			v, ok := table[th][name]
			if !ok {
				t.Errorf("colour token %q missing in %s", name, th)
				continue
			}
			if !hexColor.MatchString(v) {
				t.Errorf("colour token %q in %s = %q, want #rrggbb", name, th, v)
			}
		}
	}
}
