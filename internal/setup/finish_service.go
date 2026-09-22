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

// InstallService decides whether setup installs the login service.
func InstallService(ctx context.Context, goos string, supported, noService bool, inst ServiceInstaller, exe, configPath string) (Outcome, error) {
	return Outcome{}, nil
}
