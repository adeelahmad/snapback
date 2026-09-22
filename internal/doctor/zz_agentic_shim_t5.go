// agentic:shim
package doctor

import (
	"context"
	"io"
	"io/fs"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
)

// Manager stands in for service.Manager until S3-15 T1 lands on this chain.
type Manager string

// Runner stands in for service.Runner until S3-15 T3 lands on this chain.
type Runner func(ctx context.Context, name string, args ...string) ([]byte, error)

// Probes are the doctor's injected system probes.
type Probes struct {
	LookPath   func(string) (string, error)
	Run        Runner
	Stat       func(string) (fs.FileInfo, error)
	Mountinfo  func() (io.Reader, error)
	Repos      func(config.Repository) (provider.Validator, provider.Lister)
	Detect     func() (Manager, error)
	Statfs     func(path string) (freeInodes, totalInodes uint64, err error)
	DialStatus func(ctx context.Context) (string, error)
	MountTest  func(ctx context.Context) error
}

// Check is one doctor check result.
type Check struct {
	Name   string       `json:"name"`
	Status string       `json:"status"`
	Code   errcode.Code `json:"code"`
	Detail string       `json:"detail"`
	Fix    string       `json:"fix"`
}

// Run is a deliberately wrong shim.
func Run(context.Context, *config.Config, error, Probes) []Check {
	return nil
}
