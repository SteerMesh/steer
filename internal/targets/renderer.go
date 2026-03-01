package targets

import (
	"context"

	"github.com/SteerMesh/steer/internal/compiler"
)

// Renderer produces tool-specific output from a compiled model.
// Implementations must be deterministic (no timestamps or random data in output).
type Renderer interface {
	Name() string
	Render(ctx context.Context, model *compiler.Model) (files map[string][]byte, err error)
}

// Registry returns all built-in renderers.
func Registry() []Renderer {
	return []Renderer{
		&Kiro{},
		&Cursor{},
		&AmazonQ{},
	}
}
