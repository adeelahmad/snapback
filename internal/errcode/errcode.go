// Package errcode defines Snapback's machine-readable error codes.
package errcode

// Code is a machine-readable error code.
type Code string

// Error codes. SUB-AGENT-TODO: set each value to its contract string
// (tasks.md § T2).
const (
	InvalidConfig             Code = "SUB-AGENT-TODO: InvalidConfig"
	PrereqMissing             Code = "SUB-AGENT-TODO: PrereqMissing"
	PermissionDenied          Code = "SUB-AGENT-TODO: PermissionDenied"
	LinkConflict              Code = "SUB-AGENT-TODO: LinkConflict"
	RepoUnavailable           Code = "SUB-AGENT-TODO: RepoUnavailable"
	MappingAbsent             Code = "SUB-AGENT-TODO: MappingAbsent"
	MountFailure              Code = "SUB-AGENT-TODO: MountFailure"
	UnsupportedServiceManager Code = "SUB-AGENT-TODO: UnsupportedServiceManager"
	InodeBudgetExceeded       Code = "SUB-AGENT-TODO: InodeBudgetExceeded"
	OnAccessUnavailable       Code = "SUB-AGENT-TODO: OnAccessUnavailable"
	StaleState                Code = "SUB-AGENT-TODO: StaleState"
)

// Error is an error carrying a Code and the operation that failed.
type Error struct {
	Code Code
	Op   string
	Err  error
}

// Error returns "<op>: <code>: <err>", or "<op>: <code>" when Err is nil.
func (e *Error) Error() string {
	panic(`SUB-AGENT-TODO: return "<op>: <code>: <err>", or "<op>: <code>" when Err is nil`)
}

// Unwrap returns the cause.
func (e *Error) Unwrap() error {
	panic("SUB-AGENT-TODO: return e.Err")
}

// New returns an *Error for code, op and cause err.
func New(code Code, op string, err error) *Error {
	panic("SUB-AGENT-TODO: return &Error{Code: code, Op: op, Err: err}")
}

// Of returns the Code of the outermost *Error in err's chain, or "" when none.
func Of(err error) Code {
	panic(`SUB-AGENT-TODO: errors.As for *Error; return its Code, or "" when none (including nil)`)
}
