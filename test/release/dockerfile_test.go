package release_test

import (
	"regexp"
	"strings"
	"testing"
)

const dockerfilePath = "Dockerfile.release"

func readDockerfile(t *testing.T) string {
	t.Helper()
	return readRepoFile(t, dockerfilePath)
}

// dockerfileLines returns the Dockerfile's instruction lines with comments, blank
// lines and backslash continuations folded away, so a multi-line RUN reads as one line.
func dockerfileLines(t *testing.T, text string) []string {
	t.Helper()
	var out []string
	var pending string
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, "\\") {
			pending += strings.TrimSpace(strings.TrimSuffix(line, "\\")) + " "
			continue
		}
		out = append(out, strings.TrimSpace(pending+line))
		pending = ""
	}
	if pending != "" {
		out = append(out, strings.TrimSpace(pending))
	}
	return out
}

func instructions(t *testing.T, text, verb string) []string {
	t.Helper()
	var out []string
	for _, line := range dockerfileLines(t, text) {
		if strings.HasPrefix(strings.ToUpper(line), verb+" ") {
			out = append(out, line)
		}
	}
	return out
}

// TestDockerfileBasesPinnedByDigest requires every FROM to name an alpine tag AND a
// sha256 digest, so no rebuild can silently pick up a different base image.
func TestDockerfileBasesPinnedByDigest(t *testing.T) {
	text := readDockerfile(t)
	froms := instructions(t, text, "FROM")
	if len(froms) == 0 {
		t.Fatalf("%s has no FROM instruction", dockerfilePath)
	}
	pinned := regexp.MustCompile(`^FROM\s+alpine:[0-9][^\s@]*@sha256:[0-9a-f]{64}(\s+(?i:AS)\s+\S+)?$`)
	for _, line := range froms {
		if !pinned.MatchString(line) {
			t.Errorf("FROM is not alpine pinned by digest: %q", line)
		}
	}
}

// TestDockerfileApkPackagesPinned requires each apk add to use --no-cache and to pin
// every package as name=version; a bare package name floats with the mirror.
func TestDockerfileApkPackagesPinned(t *testing.T) {
	text := readDockerfile(t)
	pkg := regexp.MustCompile(`^[a-z0-9][a-z0-9.+_-]*=[0-9][^\s]*$`)
	found := false
	for _, line := range dockerfileLines(t, text) {
		if !strings.Contains(line, "apk add") {
			continue
		}
		found = true
		if !strings.Contains(line, "--no-cache") {
			t.Errorf("apk add without --no-cache: %q", line)
		}
		_, args, _ := strings.Cut(line, "apk add")
		for _, tok := range strings.Fields(args) {
			if strings.HasPrefix(tok, "-") || tok == "&&" || tok == "\\" {
				continue
			}
			if strings.ContainsAny(tok, "&|;") {
				break
			}
			if !pkg.MatchString(tok) {
				t.Errorf("apk package %q is not pinned as name=version in %q", tok, line)
			}
		}
	}
	if !found {
		t.Errorf("%s installs no apk packages; ruling 1 requires fuse3 pinned in the build stage", dockerfilePath)
	}
	if !regexp.MustCompile(`\bfuse3=[0-9]`).MatchString(text) {
		t.Errorf("%s does not install a pinned fuse3=<version>", dockerfilePath)
	}
}

// TestDockerfileResticVersionAndDigestPinned requires the restic version and its
// SHA256 to be declared as build args with literal, auditable values.
func TestDockerfileResticVersionAndDigestPinned(t *testing.T) {
	text := readDockerfile(t)
	version := regexp.MustCompile(`(?m)^(?:ARG|ENV)\s+RESTIC_VERSION[= ]v?\d+\.\d+\.\d+\s*$`)
	if !version.MatchString(text) {
		t.Errorf("%s has no ARG/ENV RESTIC_VERSION pinned to a semver", dockerfilePath)
	}
	sum := regexp.MustCompile(`(?m)^(?:ARG|ENV)\s+RESTIC_SHA256[= ][0-9a-f]{64}\s*$`)
	if !sum.MatchString(text) {
		t.Errorf("%s has no ARG/ENV RESTIC_SHA256 pinned to a 64-hex digest", dockerfilePath)
	}
}

// TestDockerfileResticDownloadVerified requires restic to come from the official
// GitHub release at the pinned version and to be checksum-verified before use.
func TestDockerfileResticDownloadVerified(t *testing.T) {
	text := readDockerfile(t)
	const releases = "https://github.com/restic/restic/releases/download/"
	if !strings.Contains(text, releases) {
		t.Errorf("%s does not download restic from %s", dockerfilePath, releases)
	}
	var download string
	for _, line := range dockerfileLines(t, text) {
		if strings.Contains(line, releases) {
			download = line
			break
		}
	}
	if download == "" {
		t.Fatalf("%s has no restic download instruction", dockerfilePath)
	}
	if !strings.Contains(download, "${RESTIC_VERSION}") && !strings.Contains(download, "$RESTIC_VERSION") {
		t.Errorf("restic download does not use the pinned RESTIC_VERSION: %q", download)
	}
	verify := regexp.MustCompile(`sha256sum\s+(-c|--check)`)
	if !verify.MatchString(download) {
		t.Errorf("restic download does not verify its SHA256 with sha256sum -c: %q", download)
	}
	if !strings.Contains(download, "RESTIC_SHA256") {
		t.Errorf("restic download does not check against RESTIC_SHA256: %q", download)
	}
}

// TestDockerfileCopiesPrebuiltBinary requires the image to take the static linux
// binary from the build context and to compile nothing itself.
func TestDockerfileCopiesPrebuiltBinary(t *testing.T) {
	text := readDockerfile(t)
	copyBinary := regexp.MustCompile(`(?m)^COPY\s+(?:--\S+\s+)*\S*snapback\s+/snapback\s*$`)
	if !copyBinary.MatchString(text) {
		t.Errorf("%s does not COPY the prebuilt snapback binary to /snapback", dockerfilePath)
	}
	for _, line := range dockerfileLines(t, text) {
		if regexp.MustCompile(`\bgo\s+(build|install|get|mod)\b`).MatchString(line) {
			t.Errorf("%s builds from source: %q", dockerfilePath, line)
		}
	}
	if strings.Contains(text, "golang:") {
		t.Errorf("%s must not use a golang base image; the binary is prebuilt", dockerfilePath)
	}
	copyRestic := regexp.MustCompile(`(?m)^COPY\s+--from=\S+\s+\S*restic\s+\S*restic\s*$`)
	if !copyRestic.MatchString(text) {
		t.Errorf("%s does not COPY the verified restic binary out of the build stage", dockerfilePath)
	}
}

// TestDockerfileEntrypoint pins the exec-form entrypoint the CLI contract depends on.
func TestDockerfileEntrypoint(t *testing.T) {
	text := readDockerfile(t)
	want := `ENTRYPOINT ["/snapback"]`
	found := false
	for _, line := range instructions(t, text, "ENTRYPOINT") {
		if line == want {
			found = true
		}
	}
	if !found {
		t.Errorf("%s does not set %s", dockerfilePath, want)
	}
}

// TestDockerfileNoUnpinnedFetches enforces the amended clause: no unpinned package
// and no unpinned download anywhere in the image definition.
func TestDockerfileNoUnpinnedFetches(t *testing.T) {
	text := readDockerfile(t)
	banned := []struct {
		pattern *regexp.Regexp
		why     string
	}{
		{regexp.MustCompile(`(?i)curl[^\n]*\|\s*(ba)?sh`), "pipes a downloaded script into a shell"},
		{regexp.MustCompile(`(?i)wget[^\n]*\|\s*(ba)?sh`), "pipes a downloaded script into a shell"},
		{regexp.MustCompile(`:latest\b`), "uses a latest tag"},
		{regexp.MustCompile(`(?i)\bapk\s+(upgrade|add\s+--upgrade)\b`), "upgrades packages at build time"},
		{regexp.MustCompile(`(?i)\bgo\s+install\b`), "installs a tool from the network"},
	}
	for _, line := range dockerfileLines(t, text) {
		for _, b := range banned {
			if b.pattern.MatchString(line) {
				t.Errorf("%s %s: %q", dockerfilePath, b.why, line)
			}
		}
	}
	froms := instructions(t, text, "FROM")
	if len(froms) == 0 {
		t.Errorf("%s defines no image at all", dockerfilePath)
	}
	for _, line := range froms {
		if !strings.Contains(line, "@sha256:") {
			t.Errorf("unpinned base image: %q", line)
		}
	}
}

// TestDockerfileNoticeListsRestic requires the bundled restic to be attributed in NOTICE.
func TestDockerfileNoticeListsRestic(t *testing.T) {
	notice := readRepoFile(t, "NOTICE")
	line := regexp.MustCompile(`(?m)^.*\brestic\b.*BSD-2-Clause.*$`)
	if !line.MatchString(notice) {
		t.Errorf("NOTICE has no restic entry naming BSD-2-Clause")
	}
	if !strings.Contains(notice, "https://github.com/restic/restic") {
		t.Errorf("NOTICE restic entry does not link https://github.com/restic/restic")
	}
}
