package compiler

import (
	"testing"
)

const minimalPackYAML = `
pack:
  name: security-core
  version: 1.0.0

policies:
  - id: no-secrets
    description: Do not commit secrets
    enforcement: strict

targets:
  kiro:
    output: kiro-steering.yaml
  cursor:
    output: .cursor/steering.json
`

func TestParsePack(t *testing.T) {
	m, err := ParsePack([]byte(minimalPackYAML))
	if err != nil {
		t.Fatalf("ParsePack: %v", err)
	}
	if m.Pack.Name != "security-core" || m.Pack.Version != "1.0.0" {
		t.Errorf("pack: got name=%q version=%q", m.Pack.Name, m.Pack.Version)
	}
	if len(m.Policies) != 1 {
		t.Fatalf("policies: got %d", len(m.Policies))
	}
	if m.Policies[0].ID != "no-secrets" || m.Policies[0].Enforcement != "strict" {
		t.Errorf("policy: got %+v", m.Policies[0])
	}
	if m.Targets["kiro"].Output != "kiro-steering.yaml" || m.Targets["cursor"].Output != ".cursor/steering.json" {
		t.Errorf("targets: got kiro=%q cursor=%q", m.Targets["kiro"].Output, m.Targets["cursor"].Output)
	}
}

func TestParsePack_InvalidYAML(t *testing.T) {
	_, err := ParsePack([]byte("not: valid: yaml: ["))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParsePack_InvalidSchema(t *testing.T) {
	_, err := ParsePack([]byte(`
pack:
  name: bad name
  version: 1.0.0
targets: {}
`))
	if err == nil {
		t.Fatal("expected validation error for invalid pack (name pattern)")
	}
}
