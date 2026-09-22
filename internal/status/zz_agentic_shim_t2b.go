// agentic:shim

package status

import "github.com/adeelahmad/snapback/internal/refresh"

// FromRefresh is a RED compile shim; it deliberately drops every field.
func FromRefresh(refresh.Result) refresh.Result {
	return refresh.Result{}
}
