package main

import (
	"context"
	"log/slog"
	"net"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/daemon"
	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/provider"
)

// daemonWiring carries the production Deps plus the collaborators the daemon
// installs but does not expose: the per-repository listers, the reader-policy
// gate handed to the FUSE adapter and the wrapper the history view applies to
// a catalog before publishing it. Naming them lets the debug decorators be
// driven without mounting anything.
type daemonWiring struct {
	Deps    daemon.Deps
	Listers map[string]provider.Lister
	Gate    mount.Gate
	Catalog func(mount.Catalog) mount.Catalog
}

// daemonBuilderWithLog builds the production daemon wiring for cfg, serving
// IPC on ln, with log as the logger the debug decorators write to.
//
// SUB-AGENT-TODO(S5-36/T16 GREEN): decorate the wiring with the S5-36 seams.
func daemonBuilderWithLog(ctx context.Context, cfg *config.Config, ln net.Listener, _ *slog.Logger) (daemonWiring, error) {
	deps, err := daemonBuilder(ctx, cfg, ln)
	if err != nil {
		return daemonWiring{}, err
	}
	return daemonWiring{
		Deps:    deps,
		Listers: undecoratedListers(cfg),
		Gate:    undecoratedGate{},
		Catalog: func(c mount.Catalog) mount.Catalog { return c },
	}, nil
}

// undecoratedListers returns one plain restic lister per configured
// repository, with no LogRunner in front of it.
func undecoratedListers(cfg *config.Config) map[string]provider.Lister {
	out := make(map[string]provider.Lister, len(cfg.Repositories))
	for _, r := range cfg.Repositories {
		p, err := daemonProvider(r)
		if err != nil {
			continue
		}
		out[r.ID] = p
	}
	return out
}

// undecoratedGate denies every event and logs nothing.
type undecoratedGate struct{}

func (undecoratedGate) Allow(mount.Event) bool { return false }
