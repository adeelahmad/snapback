// Package errcode defines Snapback's machine-readable error codes.
package errcode

import "errors"

// Code is a machine-readable error code.
type Code string

// Error codes.
const (
	InvalidConfig             Code = "invalid_configuration"
	PrereqMissing             Code = "prerequisite_missing"
	PermissionDenied          Code = "permission_denied"
	LinkConflict              Code = "link_conflict"
	RepoUnavailable           Code = "repository_unavailable"
	MappingAbsent             Code = "mapping_absent"
	MountFailure              Code = "mount_failure"
	UnsupportedServiceManager Code = "unsupported_service_manager"
	InodeBudgetExceeded       Code = "inode_budget_exceeded"
	OnAccessUnavailable       Code = "on_access_unavailable"
	StaleState                Code = "stale_state"
)

// Error is an error carrying a Code and the operation that failed.
type Error struct {
	Code Code
	Op   string
	Err  error
}

// Error returns "<op>: <code>: <err>", or "<op>: <code>" when Err is nil.
func (e *Error) Error() string {
	if e.Err == nil {
		return e.Op + ": " + string(e.Code)
	}
	return e.Op + ": " + string(e.Code) + ": " + e.Err.Error()
}

// Unwrap returns the cause.
func (e *Error) Unwrap() error {
	return e.Err
}

// New returns an *Error for code, op and cause err.
func New(code Code, op string, err error) *Error {
	return &Error{Code: code, Op: op, Err: err}
}

// Of returns the Code of the outermost *Error in err's chain, or "" when none.
func Of(err error) Code {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}
