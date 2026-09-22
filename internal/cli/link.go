package cli

import (
	"context"
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
			fs := NewFlagSet(env, Usage{
				Synopsis: "link [flags] DIR",
				Args:     "DIR  the directory to give a .snapshot entry",
				Example:  "snapback link ~/work",
			})
			jsonOut := fs.Bool("json", false, "write a JSON envelope")
			help, err := ParseWithUsage(fs, args)
			switch {
			case help:
				return 0
			case err != nil:
				return WriteError(env, "link", false, err)
			case fs.NArg() != 1:
				return WriteError(env, "link", *jsonOut, &UsageError{Msg: "link takes exactly one DIR"})
			}
			dir := fs.Arg(0)
			if !filepath.IsAbs(dir) {
				wd, err := d.Getwd()
				if err != nil {
					return WriteError(env, "link", *jsonOut, err)
				}
				dir = filepath.Join(wd, dir)
			}
			res, err := d.Linker.Ensure(ctx, dir)
			if err != nil {
				return WriteError(env, "link", *jsonOut, err)
			}
			if *jsonOut {
				return WriteOK(env, true, linkResult{Key: res.Key, Created: res.Created, Path: res.Path})
			}
			if code := WriteOK(env, false, res.Path); code != 0 {
				return code
			}
			if err := WriteNext(env.Stdout, "ls "+res.Path); err != nil {
				return WriteError(env, "link", false, err)
			}
			return 0
		},
	}
}
