package targets

import (
	"context"
	"encoding/json"

	"github.com/SteerMesh/steer/internal/compiler"
)

// Cursor renders the model to Cursor steering format (JSON).
// Output structure follows a minimal .cursor/steering.json-style format.
type Cursor struct{}

func (Cursor) Name() string { return "cursor" }

func (c Cursor) Render(ctx context.Context, model *compiler.Model) (map[string][]byte, error) {
	out := map[string]interface{}{
		"pack":     map[string]string{"name": model.Pack.Name, "version": model.Pack.Version},
		"policies": policiesToMaps(model.Policies),
	}
	if len(model.Rules) > 0 {
		out["rules"] = rulesToMaps(model.Rules)
	}
	raw, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, err
	}
	return map[string][]byte{".cursor/steering.json": raw}, nil
}
