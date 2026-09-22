// agentic:shim

package ipc

// Link maintenance ops (S3-10 LNK-D1); the daemon owns links.db while it runs.
const (
	OpLinksList          = "links_list"
	OpLinksRepair        = "links_repair"
	OpLinksRemoveManaged = "links_remove_managed"
)
