package setup

import "context"

// ServiceInstaller installs the Snapback login service for one machine.
type ServiceInstaller interface {
	Install(ctx context.Context, exe, configPath string) error
}

// Outcome reports what setup decided about the login service: whether it was
// installed, the named reason when it was not, and how to remove it again.
type Outcome struct {
	Installed bool
	Reason    string
	Removal   string
}

// InstallService decides whether setup installs the login service, and
// installs it when it can. The service is on by default, so only --no-service,
// a non-Linux host or a host without a supported service manager stops it;
// each of those returns the named reason and leaves the installer untouched.
func InstallService(ctx context.Context, goos string, supported, noService bool, inst ServiceInstaller, exe, configPath string) (Outcome, error) {
	switch {
	case noService:
		return Outcome{Reason: "skipped (--no-service)"}, nil
	case goos != "linux":
		return Outcome{Reason: "login service is Linux-only for now"}, nil
	case !supported:
		return Outcome{Reason: "no supported service manager detected"}, nil
	}
	if err := inst.Install(ctx, exe, configPath); err != nil {
		return Outcome{}, err
	}
	return Outcome{Installed: true, Removal: "snapback service uninstall"}, nil
}
