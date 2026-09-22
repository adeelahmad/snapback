// agentic:shim

package crawler

import "errors"

// Status is the evidence status of one crawler row.
type Status string

// ErrToolMissing reports that no candidate binary of a tool is on PATH.
var ErrToolMissing = errors.New("shim")

// Tool is one row of the crawler tool table.
type Tool struct {
	Name        string
	Candidates  []string
	Follows     bool
	VersionArgs []string
}

// Argv returns the argument vector that runs the tool over root.
func (t Tool) Argv(root, dest string) []string { return nil }

// FixedRow is a row that is recorded but never run.
type FixedRow struct {
	Name   string
	Status Status
	Reason string
	Argv   []string
}

// Tools returns the ordered crawler tool table.
func Tools() []Tool { return nil }

// Resolve returns the path of the first candidate binary lookPath finds.
func Resolve(tool Tool, lookPath func(string) (string, error)) (string, error) { return "", nil }

// VSCodeRow returns the fixed VS Code search row.
func VSCodeRow() FixedRow { return FixedRow{Status: "tested", Argv: []string{"code"}} }
