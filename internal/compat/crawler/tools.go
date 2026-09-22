package crawler

import "errors"

// Status is the evidence status of one crawler row.
type Status string

// ErrToolMissing reports that no candidate binary of a tool is on PATH.
var ErrToolMissing = errors.New("crawler: tool missing")

// Tool is one row of the crawler tool table.
type Tool struct {
	Name        string
	Candidates  []string
	Follows     bool
	VersionArgs []string
}

// Argv returns the argument vector that runs the tool over root.
func (t Tool) Argv(root, dest string) []string {
	panic("SUB-AGENT-TODO: T2 build the argv for root (rsync also uses dest); rg/fd pass --hidden --no-ignore (fd: -H -I); argument arrays only, no shell")
}

// FixedRow is a row that is recorded but never run.
type FixedRow struct {
	Name   string
	Status Status
	Reason string
	Argv   []string
}

// Tools returns the ordered crawler tool table.
func Tools() []Tool {
	panic("SUB-AGENT-TODO: T2 return the seven ordered rows rg, rg -L, fd, fd -L, find, find -L, rsync -a; fd rows try fd then fdfind")
}

// Resolve returns the path of the first candidate binary lookPath finds.
func Resolve(tool Tool, lookPath func(string) (string, error)) (string, error) {
	panic("SUB-AGENT-TODO: T2 return the first candidate lookPath finds, else an error wrapping ErrToolMissing naming every candidate tried")
}

// VSCodeRow returns the fixed VS Code search row.
func VSCodeRow() FixedRow {
	panic("SUB-AGENT-TODO: T2 return the VS Code search row with status not-tested-here, a reason and no argv")
}
