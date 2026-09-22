package latency

import (
	"errors"
	"fmt"
)

// AllowedRemote is the only rclone remote the latency harness may touch.
const AllowedRemote = "gdrive:snapback-stage1"

const remoteEnvVar = "SNAPBACK_RCLONE_REMOTE"

// CheckRemote reports an error unless remote exactly equals AllowedRemote.
func CheckRemote(remote string) error {
	if remote != AllowedRemote {
		return fmt.Errorf("remote %q refused: only %q is allowed", remote, AllowedRemote)
	}
	return nil
}

// RemoteFromEnv reads SNAPBACK_RCLONE_REMOTE through getenv and checks it.
func RemoteFromEnv(getenv func(string) string) (string, error) {
	remote := getenv(remoteEnvVar)
	if remote == "" {
		return "", errors.New(remoteEnvVar + " is not set; set it to " + AllowedRemote)
	}
	if err := CheckRemote(remote); err != nil {
		return "", err
	}
	return remote, nil
}

// RepoSpec returns the restic repository spec for AllowedRemote.
func RepoSpec() string {
	return "rclone:" + AllowedRemote
}
