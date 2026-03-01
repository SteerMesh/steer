package bundle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const ManifestVersion = "1.0"

// PackRef is a pack name and version in the manifest.
type PackRef struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// FileEntry is a generated file with path and SHA256.
type FileEntry struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// Manifest is the bundle manifest (matches spec bundle-manifest.schema.json).
type Manifest struct {
	Version        string     `json:"version"`
	BundleVersion  string     `json:"bundleVersion,omitempty"`
	Packs          []PackRef `json:"packs"`
	Files          []FileEntry `json:"files"`
	Signature      *Signature `json:"signature,omitempty"`
}

// Signature is optional signing metadata.
type Signature struct {
	Algorithm string `json:"algorithm,omitempty"`
	KeyID     string `json:"keyId,omitempty"`
	Value     string `json:"value,omitempty"`
}

// Build produces the manifest from a set of generated files and pack refs.
// File paths are sorted for deterministic output.
func Build(files map[string][]byte, packRefs []PackRef) *Manifest {
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var entries []FileEntry
	for _, p := range paths {
		sum := sha256.Sum256(files[p])
		entries = append(entries, FileEntry{Path: p, SHA256: hex.EncodeToString(sum[:])})
	}
	return &Manifest{
		Version: ManifestVersion,
		Packs:   packRefs,
		Files:   entries,
	}
}

// Write writes files and the manifest to outDir. Manifest is written as bundle-manifest.json.
func Write(outDir string, files map[string][]byte, manifest *Manifest) error {
	for path, data := range files {
		full := filepath.Join(outDir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(full, data, 0644); err != nil {
			return err
		}
	}
	manifestPath := filepath.Join(outDir, "bundle-manifest.json")
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath, raw, 0644)
}

// SHA256Hex returns the SHA256 hex digest of data.
func SHA256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// canonicalManifest is the manifest without signature for deterministic signing/verification.
type canonicalManifest struct {
	Version       string     `json:"version"`
	BundleVersion string     `json:"bundleVersion,omitempty"`
	Packs         []PackRef  `json:"packs"`
	Files         []FileEntry `json:"files"`
}

// CanonicalBytes returns JSON bytes of the manifest without the signature field (for signing/verification).
func CanonicalBytes(m *Manifest) ([]byte, error) {
	c := canonicalManifest{
		Version:       m.Version,
		BundleVersion: m.BundleVersion,
		Packs:         m.Packs,
		Files:         m.Files,
	}
	return json.Marshal(c)
}

// LoadManifest reads and parses a bundle manifest from path.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Verify checks that files on disk match the manifest checksums.
func Verify(outDir string, manifest *Manifest) error {
	for _, e := range manifest.Files {
		full := filepath.Join(outDir, e.Path)
		data, err := os.ReadFile(full)
		if err != nil {
			return fmt.Errorf("file %s: %w", e.Path, err)
		}
		if SHA256Hex(data) != e.SHA256 {
			return fmt.Errorf("file %s: checksum mismatch", e.Path)
		}
	}
	return nil
}
