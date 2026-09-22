// agentic:shim
package seed

import (
	"context"

	"github.com/adeelahmad/snapback/internal/links"
)

// BatchLinker ensures the .snapshot links of many directories at once.
type BatchLinker interface {
	EnsureBatch(ctx context.Context, dirs []string) ([]links.Result, error)
}

// BatchSize is the number of directories Run passes to one EnsureBatch call.
const BatchSize = 1
