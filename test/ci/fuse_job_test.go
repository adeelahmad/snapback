package ci_test

import (
	"regexp"
	"strings"
	"testing"
)

const (
	fuseJobName       = "fuse-linux"
	resticReleaseURL  = "https://github.com/restic/restic/releases/download/v0.19.0/restic_0.19.0_linux_amd64.bz2"
	resticLinuxSHA256 = "13176fe6d89d4357947a2cd107218ab2873a5f9d8e1ac2d4cd1c8e07e6839c21"
)

var (
	sha256HexRe     = regexp.MustCompile(`\b[0-9a-f]{64}\b`)
	goVersionFileRe = regexp.MustCompile(`go-version-file:\s*['"]?go\.mod`)
)

// fuseJobBlock returns the fuse-linux job text, failing the test when the job is absent.
func fuseJobBlock(t *testing.T) string {
	t.Helper()
	block := jobBlock(readCI(t), fuseJobName)
	if strings.TrimSpace(block) == "" {
		t.Fatalf("ci.yml has no %q job", fuseJobName)
	}
	return block
}

// aptInstallTokens returns the whitespace tokens after `apt-get install` in the fuse-linux apt step.
func aptInstallTokens(t *testing.T) []string {
	t.Helper()
	step := stepContaining(fuseJobBlock(t), "apt-get install")
	if step == "" {
		t.Fatalf("fuse-linux job has no apt-get install step")
	}
	for _, line := range strings.Split(step, "\n") {
		_, after, found := strings.Cut(line, "apt-get install")
		if found {
			return strings.Fields(after)
		}
	}
	t.Fatalf("apt-get install line not found in step:\n%s", step)
	return nil
}

func hasToken(tokens []string, want string) bool {
	for _, tok := range tokens {
		if tok == want {
			return true
		}
	}
	return false
}

func TestFuseJobExistsOnUbuntuLatest(t *testing.T) {
	block := fuseJobBlock(t)
	for _, line := range strings.Split(block, "\n") {
		if strings.TrimSpace(line) == "runs-on: ubuntu-latest" {
			return
		}
	}
	t.Errorf("fuse-linux job has no `runs-on: ubuntu-latest` line:\n%s", block)
}

func TestFuseJobCheckoutAndSetupGoPinned(t *testing.T) {
	block := fuseJobBlock(t)
	assertContainsAll(t, block, "actions/checkout@v4", "actions/setup-go@v5")
	step := stepContaining(block, "actions/setup-go")
	if !goVersionFileRe.MatchString(step) {
		t.Errorf("fuse-linux setup-go step does not use go-version-file: go.mod:\n%s", step)
	}
}

func TestFuseJobInstallsFuse3(t *testing.T) {
	step := stepContaining(fuseJobBlock(t), "apt-get install")
	if step == "" {
		t.Fatalf("fuse-linux job has no apt-get install step")
	}
	if !strings.Contains(step, "apt-get update") {
		t.Errorf("apt step does not run apt-get update:\n%s", step)
	}
	if !hasToken(aptInstallTokens(t), "fuse3") {
		t.Errorf("apt-get install line does not install fuse3:\n%s", step)
	}
}

func TestFuseJobInstallsCrawlerTools(t *testing.T) {
	tokens := aptInstallTokens(t)
	for _, want := range []string{"ripgrep", "fd-find", "rsync"} {
		if !hasToken(tokens, want) {
			t.Errorf("apt-get install tokens %v missing %q", tokens, want)
		}
	}
	if hasToken(tokens, "fd") {
		t.Errorf("apt-get install tokens %v include bare `fd`, which is not an Ubuntu package (use fd-find)", tokens)
	}
}

func TestFuseJobDoesNotInstallDistroRestic(t *testing.T) {
	tokens := aptInstallTokens(t)
	if len(tokens) == 0 {
		t.Fatalf("apt-get install line has no packages")
	}
	if hasToken(tokens, "restic") {
		t.Errorf("apt-get install tokens %v include distro restic; restic must come from the pinned 0.19.0 release", tokens)
	}
}

func TestFuseJobDownloadsPinnedRestic(t *testing.T) {
	block := fuseJobBlock(t)
	if !strings.Contains(block, resticReleaseURL) {
		t.Fatalf("fuse-linux job does not download %s", resticReleaseURL)
	}
	if strings.Contains(block, "releases/latest") {
		t.Errorf("fuse-linux job references releases/latest")
	}
	step := stepContaining(block, resticReleaseURL)
	if strings.Contains(step, "${{") {
		t.Errorf("restic download step uses an expression instead of pinned literals:\n%s", step)
	}
}

func TestFuseJobVerifiesResticChecksum(t *testing.T) {
	step := stepContaining(fuseJobBlock(t), resticReleaseURL)
	if step == "" {
		t.Fatalf("fuse-linux job has no step containing %s", resticReleaseURL)
	}
	hashes := sha256HexRe.FindAllString(step, -1)
	if len(hashes) != 1 {
		t.Fatalf("restic step has %d 64-hex literals, want exactly 1:\n%s", len(hashes), step)
	}
	if hashes[0] != resticLinuxSHA256 {
		t.Errorf("restic checksum = %s, want official %s", hashes[0], resticLinuxSHA256)
	}
	check := strings.Index(step, "sha256sum -c")
	if check < 0 {
		check = strings.Index(step, "sha256sum --check")
	}
	unpack := strings.Index(step, "bunzip2")
	if check < 0 || unpack < 0 || check > unpack {
		t.Errorf("restic step must run sha256sum -c before bunzip2 (check at %d, bunzip2 at %d):\n%s", check, unpack, step)
	}
}

func TestFuseJobAssertsResticVersion(t *testing.T) {
	step := stepContaining(fuseJobBlock(t), "restic version")
	if step == "" {
		t.Fatalf("fuse-linux job never runs `restic version`")
	}
	// Only text after `restic version` counts, so the download URL cannot satisfy the check.
	_, after, _ := strings.Cut(step, "restic version")
	if !strings.Contains(after, "0.19.0") {
		t.Errorf("restic version step does not check for 0.19.0:\n%s", step)
	}
}
