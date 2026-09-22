package service

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// UnitOptions describes the systemd unit to render.
type UnitOptions struct {
	Exe    string
	Config string
	Scope  string
	User   string
}

const unitHeader = `# Managed by snapback; do not edit.
[Unit]
Description=Snapback .snapshot history
StartLimitIntervalSec=300
StartLimitBurst=5

`

// SystemdUnit renders the systemd unit file for o.
func SystemdUnit(o UnitOptions) (string, error) {
	if !filepath.IsAbs(o.Exe) {
		return "", errcode.New(errcode.InvalidConfig, "service unit", errors.New("executable path must be absolute"))
	}
	if strings.ContainsAny(o.Exe+o.Config+o.User, "\n\r") {
		return "", errcode.New(errcode.InvalidConfig, "service unit", errors.New("unit values must not contain newlines"))
	}

	var b strings.Builder
	b.WriteString(unitHeader)
	b.WriteString("[Service]\nType=simple\nKillMode=mixed\n")
	wantedBy := "default.target"
	if o.Scope == "system" {
		wantedBy = "multi-user.target"
		if o.User != "" {
			b.WriteString("User=" + o.User + "\n")
		}
	}
	b.WriteString("ExecStart=" + systemdQuote(o.Exe) + " run --config " + systemdQuote(o.Config) + "\n")
	b.WriteString("Restart=on-failure\nRestartSec=5s\nTimeoutStopSec=30s\n\n[Install]\nWantedBy=" + wantedBy + "\n")
	return b.String(), nil
}

// systemdQuote escapes s as one ExecStart argument: % and $ specifiers are
// doubled, and values with whitespace, quotes or backslashes are double-quoted.
func systemdQuote(s string) string {
	s = strings.NewReplacer("%", "%%", "$", "$$").Replace(s)
	if !strings.ContainsAny(s, " \t\"'\\;") {
		return s
	}
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}
