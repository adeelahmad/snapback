package cli

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/discovery/seed"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/links"
)

// planCall records one PlanPath call.
type planCall struct {
	root, seedPath string
	maxDepth       int
	excludes       []string
}

// seedRig records the seed operations and doubles as the Linker.
type seedRig struct {
	mu           sync.Mutex
	plan         seed.Plan
	plans        []planCall
	preflights   []bool
	preflightErr error
	runs         []seed.Plan
	runLinkers   []seed.Linker
}

func (r *seedRig) Ensure(context.Context, string) (links.Result, error) { return links.Result{}, nil }
func (r *seedRig) List() ([]links.Record, error)                        { return nil, nil }
func (r *seedRig) Repair(context.Context) (links.RepairReport, error) {
	return links.RepairReport{}, nil
}
func (r *seedRig) RemoveManaged(context.Context) (links.RepairReport, error) {
	return links.RepairReport{}, nil
}

// seedDeps returns Deps over r loading cfg.
func seedDeps(r *seedRig, cfg config.Config) Deps {
	return Deps{
		Linker:     r,
		LoadConfig: func(string) (config.Config, error) { return cfg, nil },
		Getwd:      func() (string, error) { return cfg.Roots[0].LocalPath, nil },
		PlanPath: func(root, seedPath string, maxDepth int, excludes []string) (seed.Plan, error) {
			r.mu.Lock()
			defer r.mu.Unlock()
			r.plans = append(r.plans, planCall{root, seedPath, maxDepth, slices.Clone(excludes)})
			return r.plan, nil
		},
		Preflight: func(_ seed.Plan, force bool) error {
			r.mu.Lock()
			defer r.mu.Unlock()
			r.preflights = append(r.preflights, force)
			return r.preflightErr
		},
		RunSeed: func(_ context.Context, l seed.Linker, p seed.Plan) (seed.Report, error) {
			r.mu.Lock()
			defer r.mu.Unlock()
			r.runs = append(r.runs, p)
			r.runLinkers = append(r.runLinkers, l)
			return seed.Report{Linked: p.Count}, nil
		},
	}
}

// seedFixture returns a config over a fresh root R that has docs and src.
func seedFixture(t *testing.T) config.Config {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{"docs", "src"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	state := t.TempDir()
	return fixtureCfg(root, filepath.Join(state, "state"), filepath.Join(state, "history"), filepath.Join(state, "backend"))
}

func runSeed(d Deps, args ...string) (code int, stdout, stderr string) {
	env, out, errb := newEnv(nil)
	code = SeedCommand(d).Run(context.Background(), env, args)
	return code, out.String(), errb.String()
}

func TestSeedDryRunCreatesNoLinks(t *testing.T) {
	cfg := seedFixture(t)
	r := cfg.Roots[0].LocalPath
	dirs := []string{filepath.Join(r, "a"), filepath.Join(r, "b"), filepath.Join(r, "c")}
	rig := &seedRig{plan: seed.Plan{Dirs: dirs, Count: 3}}

	code, stdout, stderr := runSeed(seedDeps(rig, cfg), "--dry-run", "--json", r)

	if code != 0 {
		t.Fatalf("seed --dry-run --json R = %d, want 0 (stderr %q)", code, stderr)
	}
	var data struct {
		Count int      `json:"count"`
		Dirs  []string `json:"dirs"`
	}
	if err := json.Unmarshal(decodeEnvelope(t, []byte(stdout)).Data, &data); err != nil {
		t.Fatalf("decode data of %q: %v", stdout, err)
	}
	if data.Count != 3 {
		t.Errorf("seed --dry-run data.count = %d, want 3", data.Count)
	}
	if !slices.Equal(data.Dirs, dirs) {
		t.Errorf("seed --dry-run data.dirs = %q, want %q", data.Dirs, dirs)
	}
	if len(rig.preflights) != 1 {
		t.Errorf("seed --dry-run called Preflight %d times, want 1", len(rig.preflights))
	}
	if len(rig.runs) != 0 {
		t.Errorf("seed --dry-run called Run %d times, want 0", len(rig.runs))
	}
}

func TestSeedRunsPlanWithMaxDepth(t *testing.T) {
	cfg := seedFixture(t)
	cfg.Roots[0].ExcludeRelativePaths = []string{"tmp"}
	r := cfg.Roots[0].LocalPath
	docs := filepath.Join(r, "docs")
	want := seed.Plan{Dirs: []string{docs}, Count: 1}
	rig := &seedRig{plan: want}

	code, stdout, stderr := runSeed(seedDeps(rig, cfg), docs, "--max-depth", "2")

	if code != 0 {
		t.Fatalf("seed R/docs --max-depth 2 = %d, want 0 (stderr %q)", code, stderr)
	}
	if len(rig.plans) != 1 {
		t.Fatalf("seed R/docs called PlanPath %d times, want 1", len(rig.plans))
	}
	got := rig.plans[0]
	if got.root != r || got.seedPath != docs || got.maxDepth != 2 {
		t.Errorf("seed R/docs PlanPath(%q, %q, %d), want (%q, %q, 2)", got.root, got.seedPath, got.maxDepth, r, docs)
	}
	if !slices.Contains(got.excludes, "tmp") {
		t.Errorf("seed R/docs PlanPath excludes = %q, want them to include the configured %q", got.excludes, "tmp")
	}
	if len(rig.runs) != 1 {
		t.Fatalf("seed R/docs called Run %d times, want 1", len(rig.runs))
	}
	if !slices.Equal(rig.runs[0].Dirs, want.Dirs) || rig.runs[0].Count != want.Count {
		t.Errorf("seed R/docs Run got plan %+v, want %+v", rig.runs[0], want)
	}
	if rig.runLinkers[0] != seed.Linker(rig) {
		t.Errorf("seed R/docs Run got linker %v, want the Deps Linker", rig.runLinkers[0])
	}
	if stdout == "" {
		t.Errorf("seed R/docs stdout is empty, want a report")
	}
}

func TestSeedPreflightBudgetExceeded(t *testing.T) {
	budget := errcode.New(errcode.InodeBudgetExceeded, "seed.Preflight", errors.New("inode budget exceeded"))
	tests := []struct {
		name     string
		force    bool
		err      error
		wantCode int
		wantErr  string
		wantRuns int
	}{
		{name: "no force", err: budget, wantCode: 1, wantErr: "inode_budget_exceeded"},
		{name: "force", force: true, wantRuns: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := seedFixture(t)
			r := cfg.Roots[0].LocalPath
			rig := &seedRig{plan: seed.Plan{Dirs: []string{r}, Count: 1}, preflightErr: tt.err}
			args := []string{"--json", r}
			if tt.force {
				args = append(args, "--force")
			}

			code, stdout, _ := runSeed(seedDeps(rig, cfg), args...)

			if code != tt.wantCode {
				t.Errorf("seed %q = %d, want %d", args, code, tt.wantCode)
			}
			if want := []bool{tt.force}; !slices.Equal(rig.preflights, want) {
				t.Errorf("seed %q Preflight force = %v, want %v", args, rig.preflights, want)
			}
			if len(rig.runs) != tt.wantRuns {
				t.Errorf("seed %q called Run %d times, want %d", args, len(rig.runs), tt.wantRuns)
			}
			if tt.wantErr != "" {
				if got := decodeEnvelope(t, []byte(stdout)).Code; got != tt.wantErr {
					t.Errorf("seed %q code = %q, want %q", args, got, tt.wantErr)
				}
			}
		})
	}
}

func TestSeedConfiguredPaths(t *testing.T) {
	cfg := seedFixture(t)
	cfg.Roots[0].SeedPaths = []config.SeedPath{{Path: "docs", MaxDepth: 1}, {Path: "src", MaxDepth: 3}}
	r := cfg.Roots[0].LocalPath
	rig := &seedRig{plan: seed.Plan{Dirs: []string{r}, Count: 1}}

	code, _, stderr := runSeed(seedDeps(rig, cfg))

	if code != 0 {
		t.Fatalf("seed = %d, want 0 (stderr %q)", code, stderr)
	}
	want := []planCall{{root: r, seedPath: "docs", maxDepth: 1}, {root: r, seedPath: "src", maxDepth: 3}}
	if len(rig.plans) != len(want) {
		t.Fatalf("seed called PlanPath %d times, want %d", len(rig.plans), len(want))
	}
	for i, w := range want {
		got := rig.plans[i]
		if got.root != w.root || got.seedPath != w.seedPath || got.maxDepth != w.maxDepth {
			t.Errorf("seed PlanPath call %d = (%q, %q, %d), want (%q, %q, %d)", i, got.root, got.seedPath, got.maxDepth, w.root, w.seedPath, w.maxDepth)
		}
	}
}
