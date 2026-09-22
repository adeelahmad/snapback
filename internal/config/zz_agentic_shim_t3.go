// agentic:shim

package config

import (
	"os"
	"strings"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// FieldError is a compile shim for T4's type; the scaffolder replaces it.
type FieldError struct {
	Path string
	Msg  string
	Code errcode.Code
}

// within is deliberately wrong: it compares by string prefix, not by path
// component.
func within(parent, child string) bool {
	return strings.HasPrefix(child, parent)
}

// checkTopology is deliberately wrong: it flags both mounts unconditionally
// with messages that name no other field, and never checks state_dir.
func checkTopology(c *Config) []FieldError {
	return []FieldError{
		{Path: "history_mount", Msg: "agentic shim"},
		{Path: "backend_mount_dir", Msg: "agentic shim"},
	}
}

// checkCredentials is deliberately wrong: it always reports an error and
// leaks the file contents into the message.
func checkCredentials(c *Config) []FieldError {
	var errs []FieldError
	for _, r := range c.Repositories {
		b, _ := os.ReadFile(r.PasswordFile)
		errs = append(errs, FieldError{Path: "repositories[0].password_file", Msg: "agentic shim " + string(b)})
	}
	return errs
}
