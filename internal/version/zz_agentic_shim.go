// agentic:shim
package version

var (
	Version = ""
	Commit  = ""
	Target  = ""
)

func Format(version, commit, target string) string { return "" }

func String() string { return "shim" }
