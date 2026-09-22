package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

type linkRecord struct {
	Key    string       `json:"key"`
	RootID string       `json:"root_id"`
	Dir    rawpath.Path `json:"dir"`
	Target string       `json:"target"`
}

type repairEntry struct {
	Key  string       `json:"key"`
	Dir  rawpath.Path `json:"dir"`
	Code string       `json:"code,omitempty"`
}

type repairReport struct {
	Completed []repairEntry `json:"completed"`
	Repaired  []repairEntry `json:"repaired"`
	Removed   []repairEntry `json:"removed"`
	Preserved []repairEntry `json:"preserved"`
}

// linksUsage is the help text of the links command.
var linksUsage = Usage{
	Synopsis: "links <list|repair|remove> [flags]",
	Args: "list    print the managed .snapshot links\n" +
		"repair  restore the managed links that are missing or stale\n" +
		"remove  remove the managed links; requires --managed",
	Example: "snapback links remove --managed --json",
}

// LinksCommand returns the links command.
func LinksCommand(d Deps) Command {
	return Command{
		Name:    "links",
		Summary: "list, repair or remove managed .snapshot links",
		Run: func(ctx context.Context, env Env, args []string) int {
			usage := &UsageError{Msg: "usage: snapback " + linksUsage.Synopsis}
			verb, rest := "", args
			if len(args) != 0 && !strings.HasPrefix(args[0], "-") {
				verb, rest = args[0], args[1:]
			}
			fs := NewFlagSet(env, linksUsage)
			managed := fs.Bool("managed", false, "with remove, remove every registry-owned link")
			jsonOut := fs.Bool("json", false, "write a JSON envelope")
			help, err := ParseWithUsage(fs, rest)
			if help {
				return 0
			}
			if err == nil && (verb == "" || fs.NArg() != 0) {
				err = usage
			}
			if err != nil {
				return WriteError(env, "links", *jsonOut, err)
			}
			switch verb {
			case "list":
				recs, err := d.Linker.List()
				if err != nil {
					return WriteError(env, "links", *jsonOut, err)
				}
				return writeRecords(env, *jsonOut, recs)
			case "repair":
				rep, err := d.Linker.Repair(ctx)
				if err != nil {
					return WriteError(env, "links", *jsonOut, err)
				}
				return writeReport(env, *jsonOut, rep)
			case "remove":
				if !*managed {
					return WriteError(env, "links", *jsonOut, usage)
				}
				rep, err := d.Linker.RemoveManaged(ctx)
				if err != nil {
					return WriteError(env, "links", *jsonOut, err)
				}
				return writeReport(env, *jsonOut, rep)
			}
			return WriteError(env, "links", *jsonOut, usage)
		},
	}
}

// writeRecords reports recs; Dir marshals as a string or {"b64":...}.
func writeRecords(env Env, jsonOut bool, recs []links.Record) int {
	out := make([]linkRecord, 0, len(recs))
	var b strings.Builder
	for _, r := range recs {
		out = append(out, linkRecord{Key: r.Key, RootID: r.RootID, Dir: r.Dir, Target: r.Target})
		fmt.Fprintf(&b, "%s  %q\n", r.Key, string(r.Dir))
	}
	if jsonOut {
		return WriteOK(env, true, out)
	}
	return WriteOK(env, false, strings.TrimSuffix(b.String(), "\n"))
}

// writeReport reports what a Repair or RemoveManaged pass did.
func writeReport(env Env, jsonOut bool, rep links.RepairReport) int {
	var b strings.Builder
	conv := func(label string, es []links.RepairEntry) []repairEntry {
		out := make([]repairEntry, 0, len(es))
		for _, e := range es {
			out = append(out, repairEntry{Key: e.Key, Dir: e.Dir, Code: string(e.Code)})
			fmt.Fprintf(&b, "%s  %s  %q  %s\n", label, e.Key, string(e.Dir), e.Code)
		}
		return out
	}
	out := repairReport{
		Completed: conv("completed", rep.Completed),
		Repaired:  conv("repaired", rep.Repaired),
		Removed:   conv("removed", rep.Removed),
		Preserved: conv("preserved", rep.Preserved),
	}
	if jsonOut {
		return WriteOK(env, true, out)
	}
	return WriteOK(env, false, strings.TrimRight(b.String(), " \n"))
}
