package version

// SUB-AGENT-TODO: initialize to "dev"; must stay a package-level var so -ldflags -X can set it.
var Version string

// SUB-AGENT-TODO: initialize to "none"; must stay a package-level var so -ldflags -X can set it.
var Commit string

// SUB-AGENT-TODO: initialize to runtime.GOOS + "/" + runtime.GOARCH; must stay a package-level var so -ldflags -X can set it.
var Target string

func Format(version, commit, target string) string {
	panic(`SUB-AGENT-TODO: pure, no globals; return "snapback <version> (commit <commit>, target <target>)\n"`)
}

func String() string {
	panic("SUB-AGENT-TODO: return Format(Version, Commit, Target)")
}
