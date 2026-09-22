package latency

// AllowedRemote is the only rclone remote the latency harness may touch.
const AllowedRemote = "gdrive:snapback-stage1"

// CheckRemote reports an error unless remote exactly equals AllowedRemote.
func CheckRemote(remote string) error {
	panic("SUB-AGENT-TODO: T1 return nil only when remote == AllowedRemote (exact match); otherwise an error naming the given remote and AllowedRemote")
}

// RemoteFromEnv reads SNAPBACK_RCLONE_REMOTE through getenv and checks it.
func RemoteFromEnv(getenv func(string) string) (string, error) {
	panic("SUB-AGENT-TODO: T1 read getenv(\"SNAPBACK_RCLONE_REMOTE\"); unset/empty is an error naming the variable and AllowedRemote; otherwise run CheckRemote and return the value")
}

// RepoSpec returns the restic repository spec for AllowedRemote.
func RepoSpec() string {
	panic("SUB-AGENT-TODO: T1 return \"rclone:\" + AllowedRemote, i.e. \"rclone:gdrive:snapback-stage1\"")
}
