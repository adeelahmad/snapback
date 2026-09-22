// Package service installs and controls Snapback as a background service.
//
// It detects the running service manager, renders its unit file and drives
// install, start, stop, status and uninstall through a single Installer.
// Only systemd is supported for v0.1; other managers get foreground
// instructions and ErrUnsupportedManager.
package service
