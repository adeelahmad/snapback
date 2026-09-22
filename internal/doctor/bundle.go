package doctor

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// bundleIssuesURL is where the user attaches the bundle by hand; Snapback
// never uploads anything itself.
const bundleIssuesURL = "https://github.com/adeelahmad/snapback/issues"

const bundleReadme = `This is a Snapback diagnostic bundle: a local snapshot of the doctor report,
the build versions, the daemon log and the redacted configuration, collected on
this machine only. Secrets and repository credentials are redacted before they
are written here, nothing is sent anywhere automatically, and you can inspect
every file before sharing it. Review it, then attach it to an issue at
` + bundleIssuesURL + ` so the problem can be reproduced.
`

// BundleInput carries the already-collected, already-redacted material that
// goes into a diagnostic bundle. Nothing here is gathered or sent by the
// bundle writer itself.
type BundleInput struct {
	DoctorJSON     []byte
	Version        string
	Commit         string
	GOOS           string
	GOARCH         string
	DaemonLog      []byte
	ConfigRedacted []byte
}

// WriteBundle writes a gzipped tar of the diagnostic material into dir and
// returns the path it wrote. The name is derived from now in UTC, so the same
// inputs and instant always name the same file.
func WriteBundle(dir string, in BundleInput, now time.Time) (string, error) {
	path := filepath.Join(dir, "snapback-bundle-"+now.UTC().Format("20060102-150405")+".tar.gz")

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	members := []struct {
		name string
		body []byte
	}{
		{"doctor.json", in.DoctorJSON},
		{"versions.txt", []byte(bundleVersions(in))},
		{"daemon.log", in.DaemonLog},
		{"config.yaml", in.ConfigRedacted},
		{"README.txt", []byte(bundleReadme)},
	}
	for _, m := range members {
		if len(m.body) == 0 {
			continue
		}
		hdr := &tar.Header{
			Name:    m.name,
			Mode:    0o600,
			Size:    int64(len(m.body)),
			ModTime: now.UTC(),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return "", fmt.Errorf("write bundle header %s: %w", m.name, err)
		}
		if _, err := tw.Write(m.body); err != nil {
			return "", fmt.Errorf("write bundle member %s: %w", m.name, err)
		}
	}
	if err := tw.Close(); err != nil {
		return "", fmt.Errorf("close bundle tar: %w", err)
	}
	if err := gz.Close(); err != nil {
		return "", fmt.Errorf("close bundle gzip: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		return "", fmt.Errorf("write bundle %s: %w", path, err)
	}
	return path, nil
}

func bundleVersions(in BundleInput) string {
	return fmt.Sprintf("version: %s\ncommit: %s\ntarget: %s/%s\ngo: %s\n",
		in.Version, in.Commit, in.GOOS, in.GOARCH, runtime.Version())
}
