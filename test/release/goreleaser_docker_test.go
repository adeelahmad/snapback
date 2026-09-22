package release_test

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// ghcrImage is the registry ruling for S5-35: snapback's release images live under the
// project's own GitHub Container Registry namespace.
const ghcrImage = "ghcr.io/adeelahmad/snapback"

// resticSHA256 pins the restic 0.18.0 release asset per architecture. The values are the
// published SHA256SUMS entries for restic_0.18.0_linux_amd64.bz2 and _linux_arm64.bz2;
// Dockerfile.release verifies the download against whichever one the build passes in.
var resticSHA256 = map[string]string{
	"amd64": "98f6dd8bf5b59058d04bfd8dab58e196cc2a680666ccee90275a3b722374438e",
	"arm64": "ce18179c25dc5f2e33e3c233ba1e580f9de1a4566d2977e8d9600210363ec209",
}

var hex64Re = regexp.MustCompile(`^[0-9a-f]{64}$`)

// entryList returns the list items under key inside a single list entry.
func entryList(entry, key string) []string {
	block := yamlBlock(entry, key)
	if block == "" {
		return nil
	}
	var out []string
	for _, line := range strings.Split(block, "\n")[1:] {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}
		out = append(out, yamlScalar(strings.TrimPrefix(trimmed, "- ")))
	}
	return out
}

// dockerEntries returns the dockers: entries keyed by their goarch.
func dockerEntries(t *testing.T, text string) map[string]string {
	t.Helper()
	block := topLevelBlock(text, "dockers")
	if block == "" {
		t.Fatalf("%s has no top-level dockers: block; the release ships no container image", goreleaserFile)
	}
	entries := splitListEntries(block)
	if len(entries) != 2 {
		t.Fatalf("dockers: has %d entries, want exactly 2 (linux/amd64 and linux/arm64)", len(entries))
	}
	byArch := make(map[string]string, len(entries))
	for _, entry := range entries {
		arch := entryScalar(entry, "goarch")
		if arch == "" {
			t.Fatalf("dockers entry has no goarch:\n%s", entry)
		}
		if _, dup := byArch[arch]; dup {
			t.Fatalf("dockers has two entries for goarch %q", arch)
		}
		byArch[arch] = entry
	}
	return byArch
}

// TestGoreleaserDockerImagesPerArch requires one linux image build per supported
// architecture, each built from the pinned Dockerfile.release through buildx and tagged
// with the release version plus its arch suffix.
func TestGoreleaserDockerImagesPerArch(t *testing.T) {
	byArch := dockerEntries(t, readRepoFile(t, goreleaserFile))
	for _, arch := range []string{"amd64", "arm64"} {
		entry, ok := byArch[arch]
		if !ok {
			t.Errorf("dockers has no entry for goarch %q", arch)
			continue
		}
		if got := entryScalar(entry, "goos"); got != "linux" {
			t.Errorf("dockers[%s] goos = %q, want %q", arch, got, "linux")
		}
		if got := entryScalar(entry, "dockerfile"); got != dockerfilePath {
			t.Errorf("dockers[%s] dockerfile = %q, want %q", arch, got, dockerfilePath)
		}
		if got := entryScalar(entry, "use"); got != "buildx" {
			t.Errorf("dockers[%s] use = %q, want %q", arch, got, "buildx")
		}
		want := ghcrImage + ":{{ .Version }}-" + arch
		got := entryList(entry, "image_templates")
		if len(got) != 1 || got[0] != want {
			t.Errorf("dockers[%s] image_templates = %v, want exactly [%q]", arch, got, want)
		}
	}
}

// TestGoreleaserDockerBuildArgsPinRestic requires each per-arch image build to select its
// platform explicitly and to pass the restic version and that architecture's published
// SHA256 as build args, so the image cannot pick up an unverified restic binary.
func TestGoreleaserDockerBuildArgsPinRestic(t *testing.T) {
	byArch := dockerEntries(t, readRepoFile(t, goreleaserFile))
	for _, arch := range []string{"amd64", "arm64"} {
		entry, ok := byArch[arch]
		if !ok {
			t.Errorf("dockers has no entry for goarch %q", arch)
			continue
		}
		flags := entryList(entry, "build_flag_templates")
		if len(flags) == 0 {
			t.Errorf("dockers[%s] has no build_flag_templates", arch)
			continue
		}
		joined := strings.Join(flags, "\n")
		if want := "--platform=linux/" + arch; !strings.Contains(joined, want) {
			t.Errorf("dockers[%s] build_flag_templates %v missing %q", arch, flags, want)
		}
		if !strings.Contains(joined, "--build-arg=RESTIC_VERSION=") {
			t.Errorf("dockers[%s] build_flag_templates %v pin no RESTIC_VERSION", arch, flags)
		}
		const prefix = "--build-arg=RESTIC_SHA256="
		var sum string
		for _, flag := range flags {
			if strings.HasPrefix(flag, prefix) {
				sum = strings.TrimPrefix(flag, prefix)
			}
		}
		if sum == "" {
			t.Errorf("dockers[%s] build_flag_templates %v pass no %s", arch, flags, prefix)
			continue
		}
		if !hex64Re.MatchString(sum) {
			t.Errorf("dockers[%s] RESTIC_SHA256 = %q, want a 64-hex digest", arch, sum)
		}
		if sum != resticSHA256[arch] {
			t.Errorf("dockers[%s] RESTIC_SHA256 = %q, want the published linux_%s digest %q", arch, sum, arch, resticSHA256[arch])
		}
	}
}

// TestGoreleaserDockerManifestsJoinBothArches requires a multi-arch manifest for both the
// version tag and latest, each joining exactly the two per-arch images.
func TestGoreleaserDockerManifestsJoinBothArches(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	block := topLevelBlock(text, "docker_manifests")
	if block == "" {
		t.Fatalf("%s has no top-level docker_manifests: block; per-arch images are never joined", goreleaserFile)
	}
	entries := splitListEntries(block)
	if len(entries) != 2 {
		t.Fatalf("docker_manifests has %d entries, want exactly 2 (version and latest)", len(entries))
	}
	wantImages := []string{
		ghcrImage + ":{{ .Version }}-amd64",
		ghcrImage + ":{{ .Version }}-arm64",
	}
	byName := make(map[string][]string, len(entries))
	for _, entry := range entries {
		name := entryScalar(entry, "name_template")
		if name == "" {
			t.Fatalf("docker_manifests entry has no name_template:\n%s", entry)
		}
		byName[name] = entryList(entry, "image_templates")
	}
	for _, want := range []string{ghcrImage + ":{{ .Version }}", ghcrImage + ":latest"} {
		images, ok := byName[want]
		if !ok {
			t.Errorf("docker_manifests has no name_template %q, got %v", want, keysOf(byName))
			continue
		}
		if strings.Join(images, ",") != strings.Join(wantImages, ",") {
			t.Errorf("docker_manifests[%q] image_templates = %v, want %v", want, images, wantImages)
		}
	}
}

func keysOf(m map[string][]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// TestGoreleaserDockerLatestIsOnlyAManifestTag requires that `:latest` appears only as the
// published manifest tag: never as a build base in Dockerfile.release, and never on a
// per-arch image template, so a rebuild can never float onto a different base.
func TestGoreleaserDockerLatestIsOnlyAManifestTag(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	latest := 0
	for i, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || !strings.Contains(trimmed, ":latest") {
			continue
		}
		latest++
		if !strings.Contains(trimmed, "name_template:") {
			t.Errorf("%s line %d uses :latest outside a docker_manifests name_template: %q", goreleaserFile, i+1, trimmed)
		}
	}
	if latest != 1 {
		t.Errorf("%s has %d :latest tag(s), want exactly 1 (the docker_manifests latest alias)", goreleaserFile, latest)
	}
	for _, line := range instructions(t, readDockerfile(t), "FROM") {
		if strings.Contains(line, ":latest") || !strings.Contains(line, "@sha256:") {
			t.Errorf("%s base is not digest-pinned: %q", dockerfilePath, line)
		}
	}
}

// TestGoreleaserDockerConfigPassesCheck runs `goreleaser check` once the docker stages
// exist, so the image and manifest blocks are validated by goreleaser itself.
func TestGoreleaserDockerConfigPassesCheck(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	for _, key := range []string{"dockers", "docker_manifests"} {
		if topLevelBlock(text, key) == "" {
			t.Fatalf("%s has no top-level %s: block", goreleaserFile, key)
		}
	}
	bin, err := exec.LookPath("goreleaser")
	if err != nil {
		t.Skip("goreleaser not on PATH")
	}
	cmd := exec.Command(bin, "check")
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("goreleaser check failed: %v\n%s", err, out)
	}
}
