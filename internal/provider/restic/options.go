// Package restic is the restic implementation of the provider seam.
package restic

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
)

// reservedEnv lists child variables that Options.Env may not set.
var reservedEnv = []string{"PATH", "RESTIC_REPOSITORY", "RESTIC_PASSWORD", "RESTIC_PASSWORD_COMMAND"}

// Options configures a restic-backed provider.
type Options struct {
	Binary, Repository, PasswordFile, CacheDir, RcloneBinary string
	NoCache, NoLock                                          bool
	Env                                                      map[string]string
	Runner                                                   Runner
}

// Runner runs restic child processes.
type Runner interface {
	Run(ctx context.Context, name string, args, env []string) (stdout, stderr []byte, err error)
	Start(name string, args, env []string) (Process, error)
}

// Process is a started child process.
type Process interface {
	Wait() error
	Signal(os.Signal) error
	Kill() error
}

// Provider is the restic implementation of the provider seam.
type Provider struct {
	opts   Options
	runner Runner
	lstat  func(string) (os.FileInfo, error)
	goos   string
}

// New validates opts and returns a restic-backed provider.
func New(opts Options) (*Provider, error) {
	switch {
	case !filepath.IsAbs(opts.Binary):
		return nil, errors.New("restic: binary must be an absolute path")
	case opts.Repository == "":
		return nil, errors.New("restic: repository is required")
	case !filepath.IsAbs(opts.PasswordFile):
		return nil, errors.New("restic: password file must be an absolute path")
	case opts.RcloneBinary != "" && !filepath.IsAbs(opts.RcloneBinary):
		return nil, errors.New("restic: rclone binary must be an absolute path")
	case opts.CacheDir != "" && !filepath.IsAbs(opts.CacheDir):
		return nil, errors.New("restic: cache dir must be an absolute path")
	case opts.CacheDir != "" && opts.NoCache:
		return nil, errors.New("restic: cache dir and no cache are mutually exclusive")
	}
	for _, k := range reservedEnv {
		if _, ok := opts.Env[k]; ok {
			return nil, errors.New("restic: env may not set " + k)
		}
	}
	runner := opts.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	return &Provider{opts: opts, runner: runner, lstat: os.Lstat, goos: runtime.GOOS}, nil
}

func (p *Provider) globalArgs() []string {
	args := []string{"--password-file", p.opts.PasswordFile}
	if p.opts.CacheDir != "" {
		args = append(args, "--cache-dir", p.opts.CacheDir)
	} else if p.opts.NoCache {
		args = append(args, "--no-cache")
	}
	if p.opts.NoLock {
		args = append(args, "--no-lock")
	}
	return args
}

func (p *Provider) childEnv() []string {
	env := make([]string, 0, len(p.opts.Env)+3)
	for k, v := range p.opts.Env {
		env = append(env, k+"="+v)
	}
	if home := os.Getenv("HOME"); home != "" {
		env = append(env, "HOME="+home)
	}
	path := "/usr/bin:/bin"
	if p.opts.RcloneBinary != "" {
		path = filepath.Dir(p.opts.RcloneBinary) + ":" + path
	}
	env = append(env, "PATH="+path, "RESTIC_REPOSITORY="+p.opts.Repository)
	slices.Sort(env)
	return env
}

func (p *Provider) secrets() []string {
	s := []string{p.opts.Repository}
	for _, v := range p.opts.Env {
		s = append(s, v)
	}
	return s
}
