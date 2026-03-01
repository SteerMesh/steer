package compiler

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/SteerMesh/steer/internal/spec/schemas"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// ParsePack parses pack YAML bytes, validates against the pack schema, and returns a Model.
// Deterministic: same input produces same model (no timestamps or random data).
func ParsePack(yamlBytes []byte) (*Model, error) {
	var raw map[string]interface{}
	if err := yaml.Unmarshal(yamlBytes, &raw); err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	// Validate with embedded schema
	schemaBytes := schemas.PackSchemaJSON
	schemaLoader := gojsonschema.NewBytesLoader(schemaBytes)
	docLoader := gojsonschema.NewBytesLoader(jsonBytes)
	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return nil, err
	}
	if !result.Valid() {
		errs := result.Errors()
		for _, e := range errs {
			slog.Debug("schema validation error", "error", e.String())
		}
		return nil, fmt.Errorf("pack schema validation failed: %s", errs[0].String())
	}

	return mapToModel(raw), nil
}

func mapToModel(raw map[string]interface{}) *Model {
	m := &Model{
		Targets: make(map[string]TargetConfig),
	}
	if pack, ok := raw["pack"].(map[string]interface{}); ok {
		if n, ok := pack["name"].(string); ok {
			m.Pack.Name = n
		}
		if v, ok := pack["version"].(string); ok {
			m.Pack.Version = v
		}
	}
	if policies, ok := raw["policies"].([]interface{}); ok {
		for _, p := range policies {
			if pm, ok := p.(map[string]interface{}); ok {
				m.Policies = append(m.Policies, Policy{
					ID:          str(pm, "id"),
					Description: str(pm, "description"),
					Enforcement: str(pm, "enforcement"),
				})
			}
		}
	}
	if prompts, ok := raw["prompts"].([]interface{}); ok {
		for _, pr := range prompts {
			if pm, ok := pr.(map[string]interface{}); ok {
				m.Prompts = append(m.Prompts, Prompt{ID: str(pm, "id"), Fields: pm})
			}
		}
	}
	if rules, ok := raw["rules"].([]interface{}); ok {
		for _, r := range rules {
			if rm, ok := r.(map[string]interface{}); ok {
				m.Rules = append(m.Rules, Rule{Fields: rm})
			}
		}
	}
	if targets, ok := raw["targets"].(map[string]interface{}); ok {
		for id, t := range targets {
			if tm, ok := t.(map[string]interface{}); ok {
				m.Targets[id] = TargetConfig{
					Output:  str(tm, "output"),
					Options: tm,
				}
			}
		}
	}
	return m
}

func str(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
