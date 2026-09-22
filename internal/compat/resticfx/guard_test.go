package resticfx

import (
	"path/filepath"
	"testing"
)

func TestGuardRepo(t *testing.T) {
	g := Guard{TempRoot: t.TempDir(), Home: t.TempDir()}

	allowed := []string{
		filepath.Join(g.TempRoot, "repo"),
		filepath.Join(g.TempRoot, "a", "b"),
		"rclone:gdrive:snapback-stage1",
	}
	refused := []string{
		filepath.Join(g.Home, "x"),
		"rclone:gdrive:",
		"rclone:gdrive:other",
		"rclone:other:snapback-stage1",
		"/",
		g.TempRoot,
		"relative/repo",
		g.TempRoot + "/../escape",
		"sftp:host:/r",
	}

	if len(allowed) == 0 {
		t.Fatal("allowed set is empty; the refused checks would be vacuous")
	}

	for _, repo := range allowed {
		if err := GuardRepo(repo, g); err != nil {
			t.Errorf("GuardRepo(%q) = %v; want nil (allowed)", repo, err)
		}
	}
	for _, repo := range refused {
		if err := GuardRepo(repo, g); err == nil {
			t.Errorf("GuardRepo(%q) = nil; want error (refused)", repo)
		}
	}
}
