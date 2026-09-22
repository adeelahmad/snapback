package crawler

import (
	"errors"
	"fmt"
	"strings"
)

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

// Argv returns the argument vector that runs the tool over root. Only the
// rsync row uses dest. The first element is the program name; callers replace
// it with the path Resolve returns.
func (t Tool) Argv(root, dest string) []string {
	var argv []string
	switch t.Candidates[0] {
	case "rg":
		argv = []string{"rg", "--files", "--hidden", "--no-ignore"}
	case "fd":
		argv = []string{"fd", "-H", "-I"}
	case "find":
		argv = []string{"find"}
	case "rsync":
		return []string{"rsync", "-a", root, dest + "/"}
	}
	if t.Follows {
		argv = append(argv, "-L")
	}
	if t.Candidates[0] == "fd" {
		argv = append(argv, ".")
	}
	return append(argv, root)
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
	rg := []string{"rg"}
	fd := []string{"fd", "fdfind"}
	find := []string{"find"}
	version := []string{"--version"}
	return []Tool{
		{Name: "rg", Candidates: rg, VersionArgs: version},
		{Name: "rg -L", Candidates: rg, Follows: true, VersionArgs: version},
		{Name: "fd", Candidates: fd, VersionArgs: version},
		{Name: "fd -L", Candidates: fd, Follows: true, VersionArgs: version},
		{Name: "find", Candidates: find, VersionArgs: version},
		{Name: "find -L", Candidates: find, Follows: true, VersionArgs: version},
		{Name: "rsync -a", Candidates: []string{"rsync"}, VersionArgs: version},
	}
}

// Resolve returns the path of the first candidate binary lookPath finds.
// Callers normally pass exec.LookPath.
func Resolve(tool Tool, lookPath func(string) (string, error)) (string, error) {
	for _, name := range tool.Candidates {
		if path, err := lookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("%w: %s (tried %s)", ErrToolMissing, tool.Name, strings.Join(tool.Candidates, ", "))
}

// VSCodeRow returns the fixed VS Code search row.
func VSCodeRow() FixedRow {
	return FixedRow{
		Name:   "vscode search",
		Status: StatusNotTestedHere,
		Reason: "VS Code search needs a GUI session; it follows symlinks only when search.followSymlinks is true (the default)",
	}
}
