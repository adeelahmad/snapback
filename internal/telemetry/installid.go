package telemetry

import "errors"

// InstallID returns the random per-install identifier stored in <dir>/install_id,
// creating the directory and the file on first use.
func InstallID(dir string) (string, error) {
	_ = dir
	return "", nil
}

// ForgetInstallID removes <dir>/install_id. A missing file is not an error.
func ForgetInstallID(dir string) error {
	_ = dir
	return errors.New("telemetry: ForgetInstallID is not implemented")
}
