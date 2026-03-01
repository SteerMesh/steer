package targets

import (
	"context"
	"encoding/json"

	"github.com/SteerMesh/steer/internal/compiler"
)

// AmazonQ is a stub renderer for Amazon Q format.
type AmazonQ struct{}

func (AmazonQ) Name() string { return "amazonq" }

func (a AmazonQ) Render(ctx context.Context, model *compiler.Model) (map[string][]byte, error) {
	// Stub: produce a minimal placeholder file so bundle has an artifact for this target.
	out := map[string]interface{}{
		"pack":     map[string]string{"name": model.Pack.Name, "version": model.Pack.Version},
		"policies": policiesToMaps(model.Policies),
	}
	raw, err := marshalAmazonQ(out)
	if err != nil {
		return nil, err
	}
	return map[string][]byte{"amazonq-steering.json": raw}, nil
}

func marshalAmazonQ(v map[string]interface{}) ([]byte, error) {
	// Minimal JSON for now; can be replaced with Amazon Q-specific format.
	return json.MarshalIndent(v, "", "  ")
}
