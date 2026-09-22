// agentic:shim

// Package restic is the restic implementation of the provider seam.
package restic

import (
	"context"
	"os"

	"github.com/adeelahmad/snapback/internal/provider"
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
}

// ExecRunner runs real child processes.
type ExecRunner struct{}

// Run is a compile shim.
func (ExecRunner) Run(context.Context, string, []string, []string) ([]byte, []byte, error) {
	return nil, nil, nil
}

// Start is a compile shim.
func (ExecRunner) Start(string, []string, []string) (Process, error) { return nil, nil }

// New is a compile shim that performs no validation.
func New(opts Options) (*Provider, error) { return &Provider{opts: opts}, nil }

func (p *Provider) globalArgs() []string                   { return []string{"--shim"} }
func (p *Provider) childEnv() []string                     { return []string{"SHIM=1"} }
func (p *Provider) validateArgs() []string                 { return []string{"--shim"} }
func (p *Provider) listArgs() []string                     { return []string{"--shim"} }
func (p *Provider) mountArgs(string) []string              { return []string{"--shim"} }
func (p *Provider) lsArgs(provider.SnapshotID) []string    { return []string{"--shim"} }
func (p *Provider) snapArgs(provider.SnapRequest) []string { return []string{"--shim"} }
