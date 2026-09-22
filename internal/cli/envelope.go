package cli

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// fixes maps each error code to the corrective action shown to the user.
var fixes = map[errcode.Code]string{
	errcode.InvalidConfig:             "run 'snapback config validate' and correct the reported fields",
	errcode.PrereqMissing:             "install the missing prerequisite or start 'snapback run'",
	errcode.PermissionDenied:          "check ownership and permissions of the path, then retry",
	errcode.LinkConflict:              "move the existing .snapshot entry aside, then retry",
	errcode.RepoUnavailable:           "check the Restic repository location and credentials, then retry",
	errcode.MappingAbsent:             "add a root covering the path or pass --repository and --prefix",
	errcode.MountFailure:              "check that FUSE is installed and the mount point is free, then retry",
	errcode.UnsupportedServiceManager: "start 'snapback run' under your own supervisor",
	errcode.InodeBudgetExceeded:       "raise the inode budget in the configuration or narrow the roots",
	errcode.OnAccessUnavailable:       "enable on-access support or rely on scheduled refresh",
	errcode.StaleState:                "wait for the next refresh or restart 'snapback run'",
}

type okEnvelope struct {
	OK   bool `json:"ok"`
	Data any  `json:"data"`
}

type errEnvelope struct {
	OK    bool   `json:"ok"`
	Code  string `json:"code,omitempty"`
	Error string `json:"error"`
	Fix   string `json:"fix,omitempty"`
}

// WriteOK writes a success result.
func WriteOK(env Env, jsonOut bool, data any) int {
	if jsonOut {
		if err := json.NewEncoder(env.Stdout).Encode(okEnvelope{OK: true, Data: data}); err != nil {
			return 1
		}
		return 0
	}
	_, _ = fmt.Fprintln(env.Stdout, data)
	return 0
}

// WriteError writes a failure result.
func WriteError(env Env, cmd string, jsonOut bool, err error) int {
	code := errcode.Of(err)
	fix := fixes[code]
	if jsonOut {
		if encErr := json.NewEncoder(env.Stdout).Encode(errEnvelope{Code: string(code), Error: err.Error(), Fix: fix}); encErr != nil {
			return 1
		}
		return ExitCode(err)
	}
	_, _ = fmt.Fprintf(env.Stderr, "snapback %s: %v\n", cmd, err)
	if fix != "" {
		_, _ = fmt.Fprintf(env.Stderr, "fix: %s\n", fix)
	}
	return ExitCode(err)
}

// ParseFlags parses args with fs, adding --json.
func ParseFlags(fs *flag.FlagSet, args []string) (jsonOut bool, pos []string, err error) {
	j := fs.Bool("json", false, "write a JSON envelope")
	if err := fs.Parse(args); err != nil {
		return false, nil, &UsageError{Msg: err.Error()}
	}
	return *j, fs.Args(), nil
}
