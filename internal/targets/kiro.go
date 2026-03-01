package targets

import (
	"context"

	"github.com/SteerMesh/steer/internal/compiler"
	"gopkg.in/yaml.v3"
)

// Kiro renders the model to Kiro steering format (YAML).
type Kiro struct{}

func (Kiro) Name() string { return "kiro" }

func (k Kiro) Render(ctx context.Context, model *compiler.Model) (map[string][]byte, error) {
	out := map[string]interface{}{
		"pack":     map[string]string{"name": model.Pack.Name, "version": model.Pack.Version},
		"policies": policiesToMaps(model.Policies),
	}
	if len(model.Rules) > 0 {
		out["rules"] = rulesToMaps(model.Rules)
	}
	raw, err := yaml.Marshal(out)
	if err != nil {
		return nil, err
	}
	return map[string][]byte{"kiro-steering.yaml": raw}, nil
}

func policiesToMaps(p []compiler.Policy) []map[string]interface{} {
	var out []map[string]interface{}
	for _, q := range p {
		out = append(out, map[string]interface{}{
			"id": q.ID, "description": q.Description, "enforcement": q.Enforcement,
		})
	}
	return out
}

func rulesToMaps(r []compiler.Rule) []map[string]interface{} {
	var out []map[string]interface{}
	for _, x := range r {
		out = append(out, x.Fields)
	}
	return out
}
