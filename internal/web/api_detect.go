package web

import "net/http"

// detectResult is the read-only machine detection GET /api/detect reports.
// SUB-AGENT-TODO: fill from setup.Detect and setup.Probe.
type detectResult struct {
	ResticPath     string   `json:"restic_path"`
	RclonePath     string   `json:"rclone_path"`
	RepoURI        string   `json:"repo_uri"`
	CredentialFile string   `json:"credential_file"`
	Hostname       string   `json:"hostname"`
	Roots          []string `json:"roots"`
	PrefixMap      []struct {
		Hostname   string `json:"hostname"`
		SourcePath string `json:"source_path"`
		TreePrefix string `json:"tree_prefix"`
	} `json:"prefix_map"`
}

// handleAPIDetect reports what setup can infer about this machine. It is a
// compile shim: it detects nothing yet.
func (s *Server) handleAPIDetect(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, detectResult{})
}
