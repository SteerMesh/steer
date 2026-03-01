package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "steer.yaml")
	if err := os.WriteFile(path, []byte(`
packs:
  - name: security-core
    version: "1.0.0"
targets:
  - kiro
  - cursor
`), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(p.Packs) != 1 || p.Packs[0].Name != "security-core" || p.Packs[0].Version != "1.0.0" {
		t.Errorf("packs: got %+v", p.Packs)
	}
	if len(p.Targets) != 2 || p.Targets[0] != "kiro" || p.Targets[1] != "cursor" {
		t.Errorf("targets: got %v", p.Targets)
	}
}

func TestLoadFromDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "steer.yaml")
	if err := os.WriteFile(path, []byte("packs: []\ntargets: []"), 0644); err != nil {
		t.Fatal(err)
	}
	p, found, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	if found != path || p == nil {
		t.Errorf("LoadFromDir: got path=%q p=%v", found, p)
	}
}
