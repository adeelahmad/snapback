package daemon

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/status"
)

// TestStatusCommandJSONSnakeCaseKeys pins the public wire shape of
// "snapback status --json": the envelope's data object uses snake_case keys
// only, so consumers can rely on them.
func TestStatusCommandJSONSnakeCaseKeys(t *testing.T) {
	xdg := filepath.Dir(sockPath(t))
	sock := ipc.SocketPath(func(string) string { return xdg }, "")
	want := status.Snapshot{
		State: "degraded",
		Repos: []status.Repo{
			{ID: "repoA", State: "ready"},
			{ID: "repoB", State: "failed", Code: errcode.RepoUnavailable},
		},
		LastRefresh:   fixedNow,
		Generation:    7,
		EligibleCount: map[string]int{"repoA": 3},
		Links:         2,
		Discovery:     "running",
		WebURL:        "http://127.0.0.1:8080/",
	}
	l, err := ipc.Listen(sock)
	if err != nil {
		t.Fatalf("ipc.Listen(%q) = %v", sock, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		_ = ipc.Serve(ctx, l, func(_ context.Context, req ipc.Request) ipc.Response {
			if req.Op != ipc.OpStatus {
				return ipc.Response{Code: errcode.InvalidConfig, Error: "unexpected op " + req.Op}
			}
			b, err := json.Marshal(want)
			if err != nil {
				return ipc.Response{Code: errcode.InvalidConfig, Error: err.Error()}
			}
			return ipc.Response{OK: true, Data: b}
		}, ipc.ServeOptions{UID: uint32(os.Getuid())})
	}()

	env, stdout, stderr := cmdEnv(xdg, "")
	code := runWithin(t, 2*time.Second, func() int { return StatusCommand().Run(ctx, env, []string{"--json"}) })
	if code != 0 {
		t.Fatalf("StatusCommand().Run(--json) = %d, want 0 (stderr %q)", code, stderr.String())
	}

	var got struct {
		OK   bool `json:"ok"`
		Data struct {
			State         string         `json:"state"`
			EligibleCount map[string]int `json:"eligible_count"`
			Repos         []struct {
				ID    string `json:"id"`
				State string `json:"state"`
				Code  string `json:"code"`
			} `json:"repos"`
			WebURL string `json:"web_url"`
		} `json:"data"`
	}
	out := stdout.String()
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal(stdout %q) = %v", out, err)
	}
	if !got.OK {
		t.Errorf("status --json ok = false, want true")
	}
	if got.Data.State != want.State {
		t.Errorf("status --json data.state = %q, want %q (stdout %s)", got.Data.State, want.State, out)
	}
	if got.Data.EligibleCount["repoA"] != 3 {
		t.Errorf("status --json data.eligible_count = %v, want map[repoA:3] (stdout %s)", got.Data.EligibleCount, out)
	}
	if got.Data.WebURL != want.WebURL {
		t.Errorf("status --json data.web_url = %q, want %q (stdout %s)", got.Data.WebURL, want.WebURL, out)
	}
	if len(got.Data.Repos) != len(want.Repos) {
		t.Fatalf("status --json data.repos = %v, want %d entries (stdout %s)", got.Data.Repos, len(want.Repos), out)
	}
	for i, repo := range got.Data.Repos {
		if repo.ID != want.Repos[i].ID || repo.State != want.Repos[i].State {
			t.Errorf("status --json data.repos[%d] = %+v, want id %q state %q", i, repo, want.Repos[i].ID, want.Repos[i].State)
		}
	}
	for _, mixed := range []string{`"State"`, `"Repos"`, `"EligibleCount"`, `"LastRefresh"`, `"WebURL"`, `"Generation"`, `"Links"`} {
		if strings.Contains(out, mixed) {
			t.Errorf("status --json stdout = %s, want no %s key", out, mixed)
		}
	}
}
