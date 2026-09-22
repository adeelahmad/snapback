package aliases

import (
	"strings"
	"testing"
	"time"
	_ "time/tzdata"
)

func loadLondon(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatalf("time.LoadLocation(%q): %v", "Europe/London", err)
	}
	return loc
}

func TestBaseNameUTC(t *testing.T) {
	ist := time.FixedZone("", 5*3600+30*60)
	tests := []struct {
		in   time.Time
		want string
	}{
		{time.Date(2026, 9, 20, 10, 30, 15, 0, time.UTC), "2026-09-20_1030Z"},
		{time.Date(2026, 1, 2, 3, 4, 59, 0, time.UTC), "2026-01-02_0304Z"},
		{time.Date(2026, 9, 20, 16, 0, 0, 0, ist), "2026-09-20_1030Z"},
	}
	for _, tt := range tests {
		if got := baseName(tt.in, time.UTC, false); got != tt.want {
			t.Errorf("baseName(%v, UTC, false) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestBaseNameLocalNumericOffset(t *testing.T) {
	london := loadLondon(t)
	est := time.FixedZone("", -5*3600)
	tests := []struct {
		in   time.Time
		loc  *time.Location
		want string
	}{
		{time.Date(2026, 7, 1, 9, 15, 0, 0, time.UTC), london, "2026-07-01_1015+0100"},
		{time.Date(2026, 12, 1, 9, 15, 0, 0, time.UTC), london, "2026-12-01_0915+0000"},
		{time.Date(2026, 7, 1, 9, 15, 0, 0, time.UTC), est, "2026-07-01_0415-0500"},
	}
	for _, tt := range tests {
		got := baseName(tt.in, tt.loc, true)
		if got != tt.want {
			t.Errorf("baseName(%v, %v, true) = %q, want %q", tt.in, tt.loc, got, tt.want)
		}
		for _, bad := range []string{"BST", "GMT", "Z"} {
			if strings.Contains(got, bad) {
				t.Errorf("baseName(%v, %v, true) = %q, contains %q", tt.in, tt.loc, got, bad)
			}
		}
	}
}

func TestRenderZone(t *testing.T) {
	london := loadLondon(t)
	tests := []struct {
		name string
		opts Options
		want *time.Location
	}{
		{"zero options", Options{}, time.UTC},
		{"loc without local", Options{Local: false, Loc: london}, time.UTC},
		{"local with loc", Options{Local: true, Loc: london}, london},
		{"local with nil loc", Options{Local: true}, time.UTC},
	}
	for _, tt := range tests {
		got := renderZone(tt.opts)
		if got == time.Local {
			t.Errorf("%s: renderZone(%+v) = time.Local, want never time.Local", tt.name, tt.opts)
		}
		if got != tt.want {
			t.Errorf("%s: renderZone(%+v) = %v, want %v", tt.name, tt.opts, got, tt.want)
		}
	}
}

func TestDateOfUsesRenderZone(t *testing.T) {
	in := time.Date(2026, 9, 20, 23, 30, 0, 0, time.UTC)
	tests := []struct {
		loc  *time.Location
		want string
	}{
		{time.UTC, "2026-09-20"},
		{time.FixedZone("", 2*3600), "2026-09-21"},
	}
	for _, tt := range tests {
		if got := dateOf(in, tt.loc); got != tt.want {
			t.Errorf("dateOf(%v, %v) = %q, want %q", in, tt.loc, got, tt.want)
		}
	}
}
