package setup

import (
	"errors"
	"path/filepath"
	"strconv"

	"github.com/adeelahmad/snapback/internal/config"
)

// repoID is the identifier of the single repository setup writes.
const repoID = "main"

// Options carries what detection cannot infer: the caller supplies the XDG
// state directory the configuration is anchored in.
type Options struct {
	StateDir string
}

// ToConfig turns a detection result into the minimal configuration that
// describes it: one repository, one root per detected root, and the defaults
// for everything setup does not decide. It returns the validation error if the
// result does not describe a usable machine.
func ToConfig(r Result, o Options) (*config.Config, error) {
	switch {
	case r.RepoURI == "":
		return nil, errors.New("setup: no repository was detected")
	case r.CredentialFile == "":
		return nil, errors.New("setup: no repository password file was detected")
	case len(r.Roots) == 0:
		return nil, errors.New("setup: no backup root was detected")
	}

	c := config.Default()
	c.StateDir = o.StateDir
	c.HistoryMount = ""
	c.BackendMountDir = ""
	c.Repositories = []config.Repository{{
		ID:           repoID,
		Repository:   r.RepoURI,
		ResticBinary: r.ResticPath,
		PasswordFile: r.CredentialFile,
		CacheDir:     filepath.Join(o.StateDir, "cache"),
	}}
	c.Roots = rootsFor(r)

	config.ApplyDefaults(c)
	if err := config.Validate(c); err != nil {
		return nil, err
	}
	return c, nil
}

// rootsFor returns one root per detected path, each identified by the last
// element of its path and each mapping that path onto itself on the detected
// host.
func rootsFor(r Result) []config.Root {
	roots := make([]config.Root, 0, len(r.Roots))
	seen := map[string]bool{}
	for _, path := range r.Roots {
		root := config.Root{
			ID:           uniqueID(filepath.Base(path), seen),
			LocalPath:    path,
			RepositoryID: repoID,
			PrefixMap: []config.PrefixMapping{{
				Hostname:   r.Hostname,
				SourcePath: path,
				TreePrefix: path,
			}},
		}
		if r.Hostname != "" {
			root.Snapshots.Hostname = r.Hostname
		}
		roots = append(roots, root)
	}
	return roots
}

// uniqueID returns id, or id with the lowest numeric suffix that seen does not
// already hold, and records the result in seen.
func uniqueID(id string, seen map[string]bool) string {
	unique := id
	for n := 2; seen[unique]; n++ {
		unique = id + "-" + strconv.Itoa(n)
	}
	seen[unique] = true
	return unique
}
