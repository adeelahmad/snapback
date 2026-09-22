package resticfx

// InitArgs builds the argv for `restic init`.
func InitArgs(repo, pwFile string) []string {
	panic("SUB-AGENT-TODO: fresh slice: -r repo --password-file pwFile init")
}

// BackupArgs builds the argv for `restic backup`.
func BackupArgs(repo, pwFile, dir string) []string {
	panic("SUB-AGENT-TODO: fresh slice: -r repo --password-file pwFile backup --quiet dir")
}

// SnapshotsArgs builds the argv for `restic snapshots --json`.
func SnapshotsArgs(repo, pwFile string) []string {
	panic("SUB-AGENT-TODO: fresh slice: -r repo --password-file pwFile snapshots --json")
}

// LsArgs builds the argv for `restic ls --json`.
func LsArgs(repo, pwFile, snapshotID string) []string {
	panic("SUB-AGENT-TODO: fresh slice: -r repo --password-file pwFile ls --json snapshotID")
}

// MountArgs builds the argv for `restic mount` with the ids/%I path template.
func MountArgs(repo, pwFile, mnt string) []string {
	panic("SUB-AGENT-TODO: fresh slice: -r repo --password-file pwFile mount --path-template ids/%I mnt (mnt last, no other flags)")
}

// UnmountCommand returns the platform unmount command for mnt.
func UnmountCommand(goos, mnt string) (string, []string, error) {
	panic("SUB-AGENT-TODO: darwin -> umount mnt; linux -> fusermount3 -u mnt; error for any other GOOS")
}
