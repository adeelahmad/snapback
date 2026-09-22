package main

import (
	"context"
	"errors"
	"net"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/daemon"
	"github.com/adeelahmad/snapback/internal/errcode"
)

// daemonBuilder builds the production daemon.Deps. It is an interim builder
// that S3-09 T7c replaces with the real dependencies.
func daemonBuilder(context.Context, *config.Config, net.Listener) (daemon.Deps, error) {
	return daemon.Deps{}, errcode.New(errcode.PrereqMissing, "run",
		errors.New("daemon dependencies are wired in S3-09 T7c"))
}
