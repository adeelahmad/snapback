package restic

import (
	"strconv"
	"strings"
)

// redacted replaces every secret value in the rendered command line.
const redacted = "***"

// passwordFileFlag is the only restic flag whose value is a secret path.
const passwordFileFlag = "--password-file"

// secretEnvSuffixes mark env keys whose value is a credential.
var secretEnvSuffixes = []string{"_KEY", "_SECRET", "_TOKEN"}

// CommandLine renders a restic invocation for logs with every secret redacted.
// It is for logs only and must never be used to build a command: the redacted
// fields and the quoting are not a faithful round-trip of the real argv or env.
//
// The output is a single space-separated line: the env pairs in the order given,
// then the argv in the order given. An argv field is quoted with %q only when it
// contains a space; an env pair quotes only its value, as KEY="with space".
// Nil or empty args and env render as the empty string.
func CommandLine(args, env []string) string {
	fields := make([]string, 0, len(env)+len(args))
	for _, pair := range env {
		fields = append(fields, redactEnvPair(pair))
	}
	redactNext := false
	for _, arg := range args {
		field := arg
		switch {
		case redactNext:
			field = redacted
			redactNext = false
		case arg == passwordFileFlag:
			redactNext = true
		case strings.HasPrefix(arg, passwordFileFlag+"="):
			field = passwordFileFlag + "=" + redacted
		}
		fields = append(fields, quoteField(field))
	}
	return strings.Join(fields, " ")
}

// redactEnvPair redacts the value of a KEY=VALUE pair whose key names a secret,
// and quotes a value that contains a space. An entry without "=" is kept as is.
func redactEnvPair(pair string) string {
	key, value, ok := strings.Cut(pair, "=")
	if !ok {
		return pair
	}
	if isSecretEnvKey(key) {
		return key + "=" + redacted
	}
	if strings.Contains(value, " ") {
		return key + "=" + strconv.Quote(value)
	}
	return key + "=" + value
}

// isSecretEnvKey reports whether the env key names a credential.
func isSecretEnvKey(key string) bool {
	if strings.HasPrefix(key, "RESTIC_PASSWORD") || key == "RCLONE_CONFIG_PASS" {
		return true
	}
	for _, suffix := range secretEnvSuffixes {
		if strings.HasSuffix(key, suffix) {
			return true
		}
	}
	return false
}

// quoteField quotes an argv field only when it contains a space.
func quoteField(arg string) string {
	if strings.Contains(arg, " ") {
		return strconv.Quote(arg)
	}
	return arg
}
