package setup

import "context"

// nextRun is the command to run once the repository already holds the root.
const nextRun = "snapback run"

// emptyNote explains an advice of snap on a repository with no snapshots.
const emptyNote = "repository has no snapshots yet"

// Advice is the next command setup points the operator at, with the notes that
// explain why that command and not another.
type Advice struct {
	Next  string
	Notes []string
}

// Plan reads the detected repository once and returns the detection result the
// repository confirms, together with the next step. It is read-only: an empty
// repository, or one holding nothing for the first root, is advice to snap that
// root rather than an error. A result detection could not complete, or a nil
// runner, leaves both the result and the advice untouched.
func Plan(ctx context.Context, run Runner, r Result) (Result, Advice, error) {
	if run == nil || r.RepoURI == "" || r.CredentialFile == "" || r.ResticPath == "" || len(r.Roots) == 0 {
		return r, Advice{Next: nextRun}, nil
	}
	snaps, err := Probe(ctx, run, r.ResticPath, r.RepoURI, r.CredentialFile)
	if err != nil {
		return r, Advice{Next: nextRun}, err
	}

	local := r.Roots[0]
	if len(snaps) == 0 {
		return r, Advice{Next: "snapback snap " + local, Notes: []string{emptyNote}}, nil
	}

	probed := make([]ProbedSnapshot, 0, len(snaps))
	for _, s := range snaps {
		probed = append(probed, ProbedSnapshot{Hostname: s.Hostname, Paths: s.Paths})
	}
	host, mappings, reason := derivePrefixMap(local, probed)
	if host != "" {
		r.Hostname = host
	}
	r.PrefixMappings = mappings
	if reason != "" {
		return r, Advice{Next: "snapback snap " + local, Notes: []string{reason}}, nil
	}
	return r, Advice{Next: nextRun}, nil
}
