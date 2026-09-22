package doctor

import (
	"encoding/json"
	"strings"
	"testing"
)

// Verbose output is indented under the verdict line it belongs to. These are
// the two prefixes it adds; no other line in the report starts with them.
const (
	probePrefix    = "  probe: "
	observedPrefix = "  observed: "
)

// verboseBlock is one check's verdict line plus the indented lines under it.
type verboseBlock struct {
	name    string
	verdict string
	indent  []string
}

// parseReport splits the human report into one block per verdict line. A
// verdict line is any line that is not indented; the indented lines that
// follow it belong to that check.
func parseReport(t *testing.T, out string) []verboseBlock {
	t.Helper()
	var blocks []verboseBlock
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if strings.HasPrefix(line, "  ") {
			if len(blocks) == 0 {
				t.Fatalf("doctor report starts with an indented line %q; got %q", line, out)
			}
			blocks[len(blocks)-1].indent = append(blocks[len(blocks)-1].indent, line)
			continue
		}
		name, _, _ := strings.Cut(strings.TrimSpace(line), " ")
		blocks = append(blocks, verboseBlock{name: name, verdict: line})
	}
	if len(blocks) == 0 {
		t.Fatalf("doctor report has no verdict lines; got %q", out)
	}
	return blocks
}

// find returns the first indented line of b carrying prefix.
func (b verboseBlock) find(prefix string) (string, bool) {
	for i, line := range b.indent {
		if strings.HasPrefix(line, prefix) {
			return b.indent[i], true
		}
	}
	return "", false
}

// stripVerbose drops every probe and observed line from out.
func stripVerbose(out string) (rest string, dropped int) {
	var kept []string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, probePrefix) || strings.HasPrefix(line, observedPrefix) {
			dropped++
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n"), dropped
}

// TestDoctorVerbose pins the --verbose extra detail sink: every check reports
// the probe it ran and the raw observation it read, the report without the
// flag stays byte-identical to today's, --verbose --json carries the same
// detail per check, and no secret value ever reaches the verbose output.
func TestDoctorVerbose(t *testing.T) {
	t.Run("human output adds probe and observed lines", func(t *testing.T) {
		f := healthyProbes(t)
		code, out, stderr := runCommand(t, f, []string{"--verbose"})
		if code != 0 {
			t.Fatalf("doctor --verbose (healthy) = exit %d, want 0; stdout %q stderr %q", code, out, stderr)
		}
		for _, b := range parseReport(t, out) {
			probe, ok := b.find(probePrefix)
			if !ok {
				t.Errorf("doctor --verbose check %q has no %q line; block %q %q", b.name, strings.TrimSpace(probePrefix), b.verdict, b.indent)
				continue
			}
			if strings.TrimSpace(strings.TrimPrefix(probe, probePrefix)) == "" {
				t.Errorf("doctor --verbose check %q probe line %q is empty", b.name, probe)
			}
			observed, ok := b.find(observedPrefix)
			if !ok {
				t.Errorf("doctor --verbose check %q has no %q line; block %q %q", b.name, strings.TrimSpace(observedPrefix), b.verdict, b.indent)
				continue
			}
			if strings.TrimSpace(strings.TrimPrefix(observed, observedPrefix)) == "" {
				t.Errorf("doctor --verbose check %q observed line %q is empty", b.name, observed)
			}
		}
	})

	t.Run("without the flag the report is unchanged", func(t *testing.T) {
		f := healthyProbes(t)
		_, plain, _ := runCommand(t, f, nil)
		if _, dropped := stripVerbose(plain); dropped != 0 {
			t.Errorf("doctor (no flags) stdout has %d verbose lines, want 0; got %q", dropped, plain)
		}

		f = healthyProbes(t)
		_, verbose, _ := runCommand(t, f, []string{"--verbose"})
		rest, dropped := stripVerbose(verbose)
		if dropped == 0 {
			t.Fatalf("doctor --verbose stdout has no probe or observed line; got %q", verbose)
		}
		if rest != plain {
			t.Errorf("doctor --verbose stdout minus its verbose lines = %q, want it byte-identical to doctor %q", rest, plain)
		}
	})

	t.Run("json carries the detail per check", func(t *testing.T) {
		f := healthyProbes(t)
		code, out, stderr := runCommand(t, f, []string{"--verbose", "--json"})
		if code != 0 {
			t.Fatalf("doctor --verbose --json (healthy) = exit %d, want 0; stderr %q", code, stderr)
		}
		var verbose []map[string]any
		if err := json.Unmarshal([]byte(out), &verbose); err != nil {
			t.Fatalf("doctor --verbose --json stdout = %q, not a JSON array: %v", out, err)
		}
		if len(verbose) == 0 {
			t.Fatalf("doctor --verbose --json = empty array, want checks")
		}
		for _, c := range verbose {
			for _, key := range []string{"probe", "observed"} {
				got, _ := c[key].(string)
				if strings.TrimSpace(got) == "" {
					t.Errorf("doctor --verbose --json check %v has empty %q", c, key)
				}
			}
		}

		f = healthyProbes(t)
		_, plain, _ := runCommand(t, f, []string{"--json"})
		var plainChecks []map[string]any
		if err := json.Unmarshal([]byte(plain), &plainChecks); err != nil {
			t.Fatalf("doctor --json stdout = %q, not a JSON array: %v", plain, err)
		}
		for _, c := range plainChecks {
			for _, key := range []string{"probe", "observed"} {
				if _, ok := c[key]; ok {
					t.Errorf("doctor --json check %v carries %q, want it only under --verbose", c, key)
				}
			}
		}
	})

	t.Run("no secret reaches the verbose output", func(t *testing.T) {
		const (
			password  = "hunter2"
			secretEnv = "s3cr3t-access-key"
		)
		f := healthyProbes(t)
		f.cfg.Repositories[0].Environment = map[string]string{
			"AWS_SECRET_ACCESS_KEY": secretEnv,
			"RESTIC_PASSWORD":       password,
		}
		for _, args := range [][]string{{"--verbose"}, {"--verbose", "--json"}} {
			_, out, _ := runCommand(t, f, args)
			if _, dropped := stripVerbose(out); args[len(args)-1] == "--verbose" && dropped == 0 {
				t.Fatalf("doctor %v stdout has no verbose line; got %q", args, out)
			}
			for _, secret := range []string{password, secretEnv} {
				if strings.Contains(out, secret) {
					t.Errorf("doctor %v stdout leaks secret %q; got %q", args, secret, out)
				}
			}
		}
	})

	t.Run("usage lists the flag", func(t *testing.T) {
		f := healthyProbes(t)
		code, out, stderr := runCommand(t, f, []string{"-h"})
		if code != 0 {
			t.Fatalf("doctor -h = exit %d, want 0; stdout %q", code, out)
		}
		if !strings.Contains(stderr, "-verbose") {
			t.Errorf("doctor -h stderr = %q, want it to list %q", stderr, "-verbose")
		}
	})
}
