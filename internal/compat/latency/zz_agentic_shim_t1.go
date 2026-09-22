// agentic:shim

package latency

import "errors"

// AllowedRemote is a compile shim; the value is deliberately wrong.
const AllowedRemote = "wrong:remote"

// CheckRemote is a compile shim; it refuses everything with an unhelpful error.
func CheckRemote(remote string) error {
	_ = remote
	return errors.New("shim refusal")
}

// RemoteFromEnv is a compile shim; it accepts nothing and names nothing.
func RemoteFromEnv(getenv func(string) string) (string, error) {
	return "", nil
}

// RepoSpec is a compile shim; the value is deliberately wrong.
func RepoSpec() string {
	return "wrong:spec"
}
