package main

import (
	"context"
	"io/fs"

	"github.com/adeelahmad/snapback/internal/links"
)

// mountLinker adapts the links engine to daemon.MountLinker: it supplies the
// directory mode the engine needs and drops the context the engine has no use
// for.
type mountLinker struct {
	eng  *links.Engine
	mode fs.FileMode
}

func (m mountLinker) EnsureMountLink(_ context.Context, dir, target string) (links.Result, error) {
	return m.eng.EnsureMountLink(dir, target, m.mode)
}

func (m mountLinker) RemoveMountLink(_ context.Context, dir string) error {
	return m.eng.RemoveMountLink(dir)
}
