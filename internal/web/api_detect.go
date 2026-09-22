package web

import (
	"context"
	"net/http"

	"github.com/adeelahmad/snapback/internal/setup"
)

// detectResult is the read-only machine detection GET /api/detect reports.
type detectResult struct {
	ResticPath     string         `json:"restic_path"`
	RclonePath     string         `json:"rclone_path"`
	RepoURI        string         `json:"repo_uri"`
	CredentialFile string         `json:"credential_file"`
	Hostname       string         `json:"hostname"`
	Roots          []string       `json:"roots"`
	PrefixMap      []detectPrefix `json:"prefix_map"`
}

// detectPrefix is one derived prefix mapping as the API reports it.
type detectPrefix struct {
	Hostname   string `json:"hostname"`
	SourcePath string `json:"source_path"`
	TreePrefix string `json:"tree_prefix"`
}

// handleAPIDetect reports what setup can infer about this machine. It only
// reads: the environment through Options.SetupDeps and, at most once, the
// repository through the read-only probe.
func (s *Server) handleAPIDetect(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.detect(r.Context()))
}

// detect infers the machine's restic setup through Options.SetupDeps and,
// when a runner and a full repository address are known, derives the prefix
// map for the first root from one read-only probe. It writes nothing.
func (s *Server) detect(ctx context.Context) detectResult {
	res, _ := setup.Detect(s.opts.SetupDeps)
	out := detectResult{
		ResticPath:     res.ResticPath,
		RepoURI:        res.RepoURI,
		CredentialFile: res.CredentialFile,
		Hostname:       res.Hostname,
		Roots:          res.Roots,
	}
	if res.RcloneFound && s.opts.SetupDeps.LookPath != nil {
		if p, err := s.opts.SetupDeps.LookPath("rclone"); err == nil {
			out.RclonePath = p
		}
	}
	if s.opts.SetupRunner == nil || out.ResticPath == "" || out.RepoURI == "" ||
		out.CredentialFile == "" || len(out.Roots) == 0 {
		return out
	}
	snaps, err := setup.Probe(ctx, s.opts.SetupRunner, out.ResticPath, out.RepoURI, out.CredentialFile)
	if err != nil || len(snaps) == 0 {
		return out
	}
	probed := make([]setup.ProbedSnapshot, 0, len(snaps))
	for _, snap := range snaps {
		probed = append(probed, setup.ProbedSnapshot{Hostname: snap.Hostname, Paths: snap.Paths})
	}
	host, mappings, _ := setup.DerivePrefixMap(out.Roots[0], probed)
	if host != "" {
		out.Hostname = host
	}
	for _, m := range mappings {
		out.PrefixMap = append(out.PrefixMap, detectPrefix{
			Hostname:   m.Hostname,
			SourcePath: m.SourcePath,
			TreePrefix: m.TreePrefix,
		})
	}
	return out
}
