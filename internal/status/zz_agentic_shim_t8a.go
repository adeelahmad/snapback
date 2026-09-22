// agentic:shim

package status

// RecoverySummary is a compile shim for S3-10 T8a; GREEN replaces it.
type RecoverySummary struct {
	Unmounted []string
	Foreign   []string
}
