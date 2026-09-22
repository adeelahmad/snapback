// Package restic is the restic implementation of the provider seam.
package restic

import (
	"context"
	"os"
)

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
	panic("SUB-AGENT-TODO: validate per tasks.md decisions (Binary, PasswordFile absolute; RcloneBinary, CacheDir absolute when set; Repository non-empty; not CacheDir with NoCache; Env may not set PATH, RESTIC_REPOSITORY, RESTIC_PASSWORD, RESTIC_PASSWORD_COMMAND; errors never contain the repository); nil Runner becomes ExecRunner{}; lstat defaults to os.Lstat, goos to runtime.GOOS")
}

func (p *Provider) globalArgs() []string {
	panic("SUB-AGENT-TODO: --password-file <PasswordFile>, then --cache-dir <CacheDir> when set or --no-cache when NoCache, then --no-lock when NoLock")
}

func (p *Provider) childEnv() []string {
	panic("SUB-AGENT-TODO: sorted K=V list of Env entries, HOME from the process only when non-empty, PATH=<dir(RcloneBinary)>:/usr/bin:/bin (or /usr/bin:/bin), RESTIC_REPOSITORY=<repo>; never os.Environ()")
}

func (p *Provider) secrets() []string {
	panic("SUB-AGENT-TODO: the repository plus every Env value, for classify redaction")
}
