package webui

import "html/template"

// SetupView is the Setup page model.
type SetupView struct {
	Chrome
	ResticPaths    []string
	RclonePaths    []string
	RepoURI        string
	CredentialFile string
	Roots          []string
	Errors         []string

	// DetectedHost and DetectedPrefix carry what machine detection inferred:
	// the repository's snapshot hostname and the prefix map it derived.
	DetectedHost   Control
	DetectedPrefix Control

	// MountPoint is the directory the repository history is mounted under.
	MountPoint Control

	// Tour is the pre-rendered first-run tour, empty once a config exists.
	Tour template.HTML
	// PasswordControls are the repository password controls: the mode and the
	// typed secret, which is never echoed back.
	PasswordControls []Control
}

// Control is one form control on a page, shaped like the controls.html
// partials: a field path that is also the input name, its label, its value,
// its help text and the error anchored to it.
type Control struct {
	Kind     Kind
	Path     string
	Label    string
	Value    string
	Help     string
	Options  []Option
	Error    string
	Required bool
}

// ControlOption is one option of a select, chips or row control.
type ControlOption = Option

// ConfigSection is one group of controls on the Configuration page.
type ConfigSection struct {
	Title    string
	Controls []Control
}

// ConfigView is the Configuration page model. Sections carries the whole
// configuration as key-path controls; Errors is the page-level banner.
type ConfigView struct {
	Chrome
	Revision        string
	Sections        []ConfigSection
	FormSections    FormSections
	Roots           []string
	Filters         []string
	Exclusions      []string
	SeedPaths       []string
	DiscoveryMode   string
	CacheDir        string
	RefreshInterval string
	Errors          []string
	Saved           bool
}
