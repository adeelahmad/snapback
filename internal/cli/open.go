package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// openRepoFix replaces the generic repository_unavailable fix: for open the
// usual cause is a daemon that is down or has no history yet.
const openRepoFix = "run 'snapback status' to check the daemon and repository, then retry"

// OpenCommand returns the open command.
func OpenCommand(d Deps) Command {
	return Command{
		Name:    "open",
		Summary: "open a directory's .snapshot history in the file manager",
		Run: func(ctx context.Context, env Env, args []string) int {
			fs := flag.NewFlagSet("open", flag.ContinueOnError)
			fs.SetOutput(io.Discard)
			jsonOut, pos, err := ParseFlags(fs, args)
			if err == nil && len(pos) != 1 {
				err = &UsageError{Msg: "usage: snapback open [--json] <dir>"}
			}
			if err != nil {
				return WriteError(env, "open", jsonOut, err)
			}
			ctx, cancel := context.WithTimeout(ctx, d.OpenTimeout)
			defer cancel()
			if err := runOpen(ctx, d, pos[0]); err != nil {
				return writeOpenError(env, jsonOut, err)
			}
			return 0
		},
	}
}

func runOpen(ctx context.Context, d Deps, dir string) error {
	if !filepath.IsAbs(dir) {
		wd, err := d.Getwd()
		if err != nil {
			return err
		}
		dir = filepath.Join(wd, dir)
	}
	res, err := d.Linker.Ensure(ctx, dir)
	if err != nil {
		return err
	}
	daemon, err := d.Daemon(ctx)
	if err != nil {
		return errcode.New(errcode.RepoUnavailable, "open", err)
	}
	ok, err := daemon.HistoryAvailable(ctx)
	if err != nil {
		return errcode.New(errcode.RepoUnavailable, "open", err)
	}
	if !ok {
		return errcode.New(errcode.RepoUnavailable, "open", errors.New("history is not available"))
	}
	opener, err := d.LookPath("xdg-open")
	if err != nil {
		return errcode.New(errcode.PrereqMissing, "open", err)
	}
	if err := d.Exec(ctx, opener, []string{res.Path}); err != nil {
		return errcode.New(errcode.PrereqMissing, "open", err)
	}
	return nil
}

// writeOpenError is WriteError with the open-specific repository_unavailable fix.
func writeOpenError(env Env, jsonOut bool, err error) int {
	if errcode.Of(err) != errcode.RepoUnavailable {
		return WriteError(env, "open", jsonOut, err)
	}
	if jsonOut {
		e := errEnvelope{Code: string(errcode.RepoUnavailable), Error: err.Error(), Fix: openRepoFix}
		if encErr := json.NewEncoder(env.Stdout).Encode(e); encErr != nil {
			return 1
		}
		return ExitCode(err)
	}
	_, _ = fmt.Fprintf(env.Stderr, "snapback open: %v\nfix: %s\n", err, openRepoFix)
	return ExitCode(err)
}
