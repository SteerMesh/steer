package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuild(t *testing.T) {
	files := map[string][]byte{
		"a.txt": []byte("hello"),
		"b.txt": []byte("world"),
	}
	packRefs := []PackRef{{Name: "security-core", Version: "1.0.0"}}
	m := Build(files, packRefs)
	if m.Version != ManifestVersion || len(m.Packs) != 1 || len(m.Files) != 2 {
		t.Fatalf("manifest: %+v", m)
	}
	// Deterministic order (paths sorted)
	if m.Files[0].Path > m.Files[1].Path {
		t.Errorf("files should be sorted by path")
	}
	if len(m.Files[0].SHA256) != 64 {
		t.Errorf("sha256 hex length: got %d", len(m.Files[0].SHA256))
	}
}

func TestWriteAndLoadAndVerify(t *testing.T) {
	dir := t.TempDir()
	files := map[string][]byte{"out/foo.txt": []byte("content")}
	manifest := Build(files, []PackRef{{Name: "p", Version: "1.0.0"}})
	if err := Write(dir, files, manifest); err != nil {
		t.Fatalf("Write: %v", err)
	}
	loaded, err := LoadManifest(filepath.Join(dir, "bundle-manifest.json"))
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if loaded.Version != manifest.Version || len(loaded.Files) != 1 {
		t.Errorf("loaded manifest: %+v", loaded)
	}
	if err := Verify(dir, loaded); err != nil {
		t.Fatalf("Verify: %v", err)
	}
	// Tamper file and verify fails
	os.WriteFile(filepath.Join(dir, "out/foo.txt"), []byte("x"), 0644)
	if err := Verify(dir, loaded); err == nil {
		t.Fatal("expected Verify to fail after tampering")
	}
}
