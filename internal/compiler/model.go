package compiler

// Model is the in-memory representation of a compiled steering pack (or merged packs)
// consumed by target renderers.
type Model struct {
	Pack     PackInfo
	Policies []Policy
	Prompts  []Prompt
	Rules    []Rule
	Targets  map[string]TargetConfig
}

// PackInfo is pack metadata.
type PackInfo struct {
	Name    string
	Version string
}

// Policy is a steering policy.
type Policy struct {
	ID          string
	Description string
	Enforcement string // strict | advisory
}

// Prompt is a prompt template (opaque for compiler; renderers interpret).
type Prompt struct {
	ID     string
	Fields map[string]interface{}
}

// Rule is a rule entry (opaque for compiler; renderers interpret).
type Rule struct {
	Fields map[string]interface{}
}

// TargetConfig is the output path and optional options for a target.
type TargetConfig struct {
	Output  string
	Options map[string]interface{}
}
