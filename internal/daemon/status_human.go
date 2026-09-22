package daemon

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/status"
)

// RenderHuman renders s as byte-stable, human-readable multi-line text: one
// line per fact, sorted identifiers, and a trailing newline.
func RenderHuman(s status.Snapshot) string {
	var b strings.Builder

	fmt.Fprintf(&b, "snapback: %s\n", orNone(s.State))
	writeRepos(&b, s.Repos)
	fmt.Fprintf(&b, "links: %d (%d with history)\n", s.Links, withHistory(s.EligibleCount))
	fmt.Fprintf(&b, "discovery: %s\n", orNone(s.Discovery))
	writeRefresh(&b, s.Generation, s.LastRefresh)

	if s.Prewarm.Warm > 0 || s.Prewarm.Pending > 0 {
		fmt.Fprintf(&b, "prewarm: %d warm, %d pending\n", s.Prewarm.Warm, s.Prewarm.Pending)
	}
	if s.WebURL != "" {
		fmt.Fprintf(&b, "web: %s\n", s.WebURL)
	}
	writeThrottle(&b, s)

	return b.String()
}

func writeRepos(b *strings.Builder, repos []status.Repo) {
	ready := 0
	var failed []string
	for _, r := range repos {
		if r.State == string(history.StateReady) {
			ready++
			continue
		}
		failed = append(failed, r.ID)
	}
	slices.Sort(failed)

	fmt.Fprintf(b, "repositories: %d ready, %d failed", ready, len(failed))
	if len(failed) > 0 {
		fmt.Fprintf(b, " (%s)", strings.Join(failed, ", "))
	}
	b.WriteString("\n")
}

func writeRefresh(b *strings.Builder, generation uint64, last time.Time) {
	if last.IsZero() {
		b.WriteString("refresh: never\n")
		return
	}
	fmt.Fprintf(b, "refresh: generation %d at %s\n", generation, last.UTC().Format(time.RFC3339))
}

func writeThrottle(b *strings.Builder, s status.Snapshot) {
	if len(s.Throttle) == 0 {
		return
	}
	parts := make([]string, 0, len(s.Throttle))
	for _, e := range s.Throttle {
		parts = append(parts, fmt.Sprintf("%s(%d) %s", e.Process, e.PID, e.Rule))
	}
	fmt.Fprintf(b, "throttled: %s\n", strings.Join(parts, ", "))
}

func withHistory(counts map[string]int) int {
	n := 0
	for _, c := range counts {
		if c > 0 {
			n++
		}
	}
	return n
}

func orNone(v string) string {
	if v == "" {
		return "none"
	}
	return v
}
