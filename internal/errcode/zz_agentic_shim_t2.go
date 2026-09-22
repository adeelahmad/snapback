// agentic:shim

// Package errcode is a RED compile shim with deliberately wrong bodies.
package errcode

// Code is a shim.
type Code string

// Shim constants: deliberately wrong values.
const (
	InvalidConfig             Code = "shim_01"
	PrereqMissing             Code = "shim_02"
	PermissionDenied          Code = "shim_03"
	LinkConflict              Code = "shim_04"
	RepoUnavailable           Code = "shim_05"
	MappingAbsent             Code = "shim_06"
	MountFailure              Code = "shim_07"
	UnsupportedServiceManager Code = "shim_08"
	InodeBudgetExceeded       Code = "shim_09"
	OnAccessUnavailable       Code = "shim_10"
	StaleState                Code = "shim_11"
)

// Error is a shim.
type Error struct {
	Code Code
	Op   string
	Err  error
}

// Error is a shim returning a wrong string.
func (e *Error) Error() string { return "shim" }

// Unwrap is a shim that hides the cause.
func (e *Error) Unwrap() error { return nil }

// New is a shim that drops its arguments.
func New(code Code, op string, err error) *Error { return &Error{} }

// Of is a shim that never finds a code.
func Of(err error) Code { return "" }
