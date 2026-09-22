// agentic:shim
package resticfx

import (
	"errors"
	"os"
	"path/filepath"
)

var shimT2Shared = []string{"shim"}

func InitArgs(repo, pwFile string) []string { return shimT2Shared }

func BackupArgs(repo, pwFile, dir string) []string { return shimT2Shared }

func SnapshotsArgs(repo, pwFile string) []string { return shimT2Shared }

func LsArgs(repo, pwFile, snapshotID string) []string { return shimT2Shared }

func MountArgs(repo, pwFile, mnt string) []string { return shimT2Shared }

func UnmountCommand(goos, mnt string) (string, []string, error) {
	return "shim", nil, errors.New("shim")
}

func NewPasswordFile(dir string) (string, error) {
	p := filepath.Join(dir, "shim-pw")
	if err := os.WriteFile(p, []byte("shim"), 0o600); err != nil {
		return "", err
	}
	return p, nil
}
