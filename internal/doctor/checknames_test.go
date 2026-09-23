package doctor

import (
	"context"
	"slices"
	"strings"
	"testing"
)

func TestCheckNamesIsNonEmptyAndUnique(t *testing.T) {
	names := CheckNames()
	if len(names) == 0 {
		t.Fatal("CheckNames() is empty, want the registered check names")
	}
	seen := make(map[string]bool, len(names))
	for _, n := range names {
		if seen[n] {
			t.Errorf("CheckNames() has duplicate name %q", n)
		}
		seen[n] = true
	}
}

func TestCheckNamesMatchesTheRunnersChecks(t *testing.T) {
	f := healthyProbes(t)
	got := Run(context.Background(), f.cfg, nil, f.probes)

	var fixed []string
	for _, c := range got {
		// Per-repository checks are named "repository:<id>" and "mapping:<id>";
		// their names depend on configuration, not on registration.
		if strings.Contains(c.Name, ":") {
			continue
		}
		fixed = append(fixed, c.Name)
	}

	if want := CheckNames(); !slices.Equal(fixed, want) {
		t.Fatalf("Run() fixed check names = %v, want CheckNames() = %v", fixed, want)
	}
}

func TestCheckNamesReturnsAFreshCopy(t *testing.T) {
	first := CheckNames()
	if len(first) == 0 {
		t.Fatal("CheckNames() is empty")
	}
	first[0] = "mutated"
	second := CheckNames()
	if second[0] == "mutated" {
		t.Fatal("CheckNames() shares backing storage across calls")
	}
}
