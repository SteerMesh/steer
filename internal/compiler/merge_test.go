package compiler

import (
	"testing"
)

func TestMerge(t *testing.T) {
	a := &Model{
		Pack:     PackInfo{Name: "first", Version: "1.0.0"},
		Policies: []Policy{{ID: "p1", Enforcement: "strict"}},
		Targets:  map[string]TargetConfig{"kiro": {Output: "kiro.yaml"}},
	}
	b := &Model{
		Pack:     PackInfo{Name: "second", Version: "2.0.0"},
		Policies: []Policy{{ID: "p2", Enforcement: "advisory"}},
		Targets:  map[string]TargetConfig{"cursor": {Output: ".cursor/steering.json"}},
	}
	merged := Merge([]*Model{a, b})
	if merged.Pack.Name != "first" || merged.Pack.Version != "1.0.0" {
		t.Errorf("pack: got %+v", merged.Pack)
	}
	if len(merged.Policies) != 2 || merged.Policies[0].ID != "p1" || merged.Policies[1].ID != "p2" {
		t.Errorf("policies: %+v", merged.Policies)
	}
	if merged.Targets["kiro"].Output != "kiro.yaml" || merged.Targets["cursor"].Output != ".cursor/steering.json" {
		t.Errorf("targets: %+v", merged.Targets)
	}
}

func TestMerge_empty(t *testing.T) {
	m := Merge(nil)
	if m == nil || len(m.Targets) != 0 {
		t.Errorf("Merge(nil): %+v", m)
	}
}
