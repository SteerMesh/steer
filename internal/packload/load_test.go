package packload

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_file(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pack.yaml")
	if err := os.WriteFile(path, []byte("pack:\n  name: test\n  version: 1.0.0\ntargets: {}"), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := Load(path, dir, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected content")
	}
}

func TestLoad_dir(t *testing.T) {
	dir := t.TempDir()
	packDir := filepath.Join(dir, "security-core")
	if err := os.MkdirAll(packDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packDir, "pack.yaml"), []byte("pack:\n  name: security-core\n  version: 1.0.0\ntargets: {}"), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := Load(packDir, dir, "")
	if err != nil {
		t.Fatalf("Load dir: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected content")
	}
}

func TestLoad_fileScheme(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pack.yaml")
	if err := os.WriteFile(path, []byte("pack:\n  name: x\n  version: 1.0.0\ntargets: {}"), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := Load("file://"+path, dir, "")
	if err != nil {
		t.Fatalf("Load file://: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected content")
	}
}
