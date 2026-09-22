package doctor

import (
	"regexp"
	"sort"
	"strings"

	"github.com/adeelahmad/snapback/internal/config"
)

// bundleRedactedMarker replaces every value removed from bundle material.
const bundleRedactedMarker = "<redacted>"

// bundlePasswordValue matches a password-like key followed by its value, so a
// secret that never appeared in the config is masked too.
var bundlePasswordValue = regexp.MustCompile(`(?i)("?[A-Za-z_]*password[A-Za-z_]*"?\s*[:=]\s*"?)([^"\s,}]+)`)

// redactForBundle prepares the configuration, the doctor report and the daemon
// log for a diagnostic bundle. The configuration goes through config.Redact and
// the config layer's marshaller; all three then lose every repository URI and
// password-file path the configuration names, and the doctor report and the log
// lose any password-like value as well. Repository ids survive, and hostnames
// are left in place: they are part of the doctor output the user consents to
// attach.
func redactForBundle(cfg *config.Config, doctorJSON, daemonLog []byte) ([]byte, []byte, []byte, error) {
	configYAML, err := config.Marshal(config.Redact(cfg))
	if err != nil {
		return nil, nil, nil, err
	}
	secrets := bundleSecrets(cfg)
	return bundleReplaceSecrets(configYAML, secrets), bundleScrub(doctorJSON, secrets), bundleScrub(daemonLog, secrets), nil
}

// bundleSecrets lists the repository URIs and password-file paths of cfg,
// longest first so a value that contains another is replaced whole.
func bundleSecrets(cfg *config.Config) []string {
	if cfg == nil {
		return nil
	}
	var secrets []string
	for _, repo := range cfg.Repositories {
		for _, v := range []string{repo.Repository, repo.PasswordFile} {
			if v != "" {
				secrets = append(secrets, v)
			}
		}
	}
	sort.SliceStable(secrets, func(i, j int) bool { return len(secrets[i]) > len(secrets[j]) })
	return secrets
}

// bundleReplaceSecrets replaces every exact secret value in data.
func bundleReplaceSecrets(data []byte, secrets []string) []byte {
	if len(data) == 0 {
		return data
	}
	out := string(data)
	for _, s := range secrets {
		out = strings.ReplaceAll(out, s, bundleRedactedMarker)
	}
	return []byte(out)
}

// bundleScrub replaces every secret and every password-like value in data.
func bundleScrub(data []byte, secrets []string) []byte {
	if len(data) == 0 {
		return data
	}
	out := bundlePasswordValue.ReplaceAllString(string(bundleReplaceSecrets(data, secrets)), "${1}"+bundleRedactedMarker)
	return []byte(out)
}
