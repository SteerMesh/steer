package lock

import (
	"path/filepath"
	"testing"
)

func TestLoadSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "steer.lock")
	l := &Lockfile{Packs: []Entry{
		{Name: "b", Version: "1.0.0", Source: "https://registry.steermesh.dev"},
		{Name: "a", Version: "2.0.0", Source: "file:///packs"},
	}}
	if err := Save(path, l); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Version != LockfileVersion || len(loaded.Packs) != 2 {
		t.Fatalf("loaded: %+v", loaded)
	}
	// Saved order should be sorted by name
	if loaded.Packs[0].Name != "a" || loaded.Packs[1].Name != "b" {
		t.Errorf("packs not sorted: %v %v", loaded.Packs[0].Name, loaded.Packs[1].Name)
	}
}

func TestAddOrUpdate(t *testing.T) {
	l := &Lockfile{}
	l.AddOrUpdate("p1", "1.0.0", "https://reg", "abc")
	l.AddOrUpdate("p1", "1.0.1", "https://reg", "def")
	if len(l.Packs) != 1 || l.Packs[0].Version != "1.0.1" || l.Packs[0].Checksum != "def" {
		t.Errorf("AddOrUpdate: %+v", l.Packs)
	}
	l.AddOrUpdate("p2", "2.0.0", "file:///local", "")
	if len(l.Packs) != 2 {
		t.Errorf("expected 2 packs, got %d", len(l.Packs))
	}
}

func TestDrift(t *testing.T) {
	l := &Lockfile{Packs: []Entry{
		{Name: "a", Version: "1.0.0", Source: "x"},
		{Name: "b", Version: "1.0.0", Source: "x"},
	}}
	report := l.Drift([]string{"a", "c"})
	if len(report.MissingInLock) != 1 || report.MissingInLock[0] != "c" {
		t.Errorf("MissingInLock: %v", report.MissingInLock)
	}
	if len(report.MissingInConfig) != 1 || report.MissingInConfig[0] != "b" {
		t.Errorf("MissingInConfig: %v", report.MissingInConfig)
	}
}
