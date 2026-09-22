package cli

import (
	"context"
	"flag"
	"io"
	"path/filepath"
)

type linkResult struct {
	Key     string `json:"key"`
	Created bool   `json:"created"`
	Path    string `json:"path"`
}

// LinkCommand returns the link command.
func LinkCommand(d Deps) Command {
	return Command{
		Name:    "link",
		Summary: "create the .snapshot link in a directory",
		Run: func(ctx context.Context, env Env, args []string) int {
			fs := flag.NewFlagSet("link", flag.ContinueOnError)
			fs.SetOutput(io.Discard)
			jsonOut, pos, err := ParseFlags(fs, args)
			if err == nil && len(pos) != 1 {
				err = &UsageError{Msg: "usage: snapback link [--json] <dir>"}
			}
			if err != nil {
				return WriteError(env, "link", jsonOut, err)
			}
			dir := pos[0]
			if !filepath.IsAbs(dir) {
				wd, err := d.Getwd()
				if err != nil {
					return WriteError(env, "link", jsonOut, err)
				}
				dir = filepath.Join(wd, dir)
			}
			res, err := d.Linker.Ensure(ctx, dir)
			if err != nil {
				return WriteError(env, "link", jsonOut, err)
			}
			if jsonOut {
				return WriteOK(env, true, linkResult{Key: res.Key, Created: res.Created, Path: res.Path})
			}
			return WriteOK(env, false, res.Path)
		},
	}
}
