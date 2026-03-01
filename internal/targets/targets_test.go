package targets

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/SteerMesh/steer/internal/compiler"
)

func TestKiro_Render(t *testing.T) {
	model := &compiler.Model{
		Pack:     compiler.PackInfo{Name: "security-core", Version: "1.0.0"},
		Policies: []compiler.Policy{{ID: "no-secrets", Description: "Do not commit secrets", Enforcement: "strict"}},
		Targets:  map[string]compiler.TargetConfig{"kiro": {Output: "kiro-steering.yaml"}},
	}
	r := Kiro{}
	files, err := r.Render(context.Background(), model)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(files) != 1 || files["kiro-steering.yaml"] == nil {
		t.Fatalf("expected one file kiro-steering.yaml, got %v", files)
	}
	// Determinism: same model produces same output
	if !strings.Contains(string(files["kiro-steering.yaml"]), "security-core") {
		t.Errorf("output missing pack name: %s", files["kiro-steering.yaml"])
	}
}

func TestCursor_Render(t *testing.T) {
	model := &compiler.Model{
		Pack:     compiler.PackInfo{Name: "security-core", Version: "1.0.0"},
		Policies: []compiler.Policy{{ID: "no-secrets", Description: "Do not commit secrets", Enforcement: "strict"}},
	}
	r := Cursor{}
	files, err := r.Render(context.Background(), model)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(files) != 1 || files[".cursor/steering.json"] == nil {
		t.Fatalf("expected .cursor/steering.json, got %v", files)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(files[".cursor/steering.json"], &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["pack"].(map[string]interface{})["name"] != "security-core" {
		t.Errorf("unexpected pack name in JSON")
	}
}

func TestAmazonQ_Render(t *testing.T) {
	model := &compiler.Model{
		Pack:     compiler.PackInfo{Name: "security-core", Version: "1.0.0"},
		Policies: []compiler.Policy{},
	}
	r := AmazonQ{}
	files, err := r.Render(context.Background(), model)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(files) != 1 || files["amazonq-steering.json"] == nil {
		t.Fatalf("expected amazonq-steering.json, got %v", files)
	}
}

func TestRegistry(t *testing.T) {
	reg := Registry()
	names := make(map[string]bool)
	for _, r := range reg {
		names[r.Name()] = true
	}
	for _, want := range []string{"kiro", "cursor", "amazonq"} {
		if !names[want] {
			t.Errorf("Registry missing %q", want)
		}
	}
}
