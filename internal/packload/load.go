// Package packload loads pack YAML from a source (file path or URL).
package packload

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Load reads pack content from source. Source may be:
//   - "https://..." — HTTP GET; optional checksum verification if wantChecksum is non-empty
//   - "file://<path>" — path relative to cwd or absolute
//   - "<path>" — relative to cwd (e.g. ./packs/security-core)
//
// For file sources, a pack is either a directory containing pack.yaml (or pack.yml), or a single .yaml/.yml file.
func Load(source string, cwd string, wantChecksum string) ([]byte, error) {
	if strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "http://") {
		return loadHTTP(source, wantChecksum)
	}
	return loadFile(source, cwd)
}

func loadFile(source string, cwd string) ([]byte, error) {
	path := source
	if strings.HasPrefix(source, "file://") {
		path = strings.TrimPrefix(source, "file://")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(cwd, path)
	}
	path = filepath.Clean(path)

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		for _, name := range []string{"pack.yaml", "pack.yml"} {
			p := filepath.Join(path, name)
			if data, err := os.ReadFile(p); err == nil {
				return data, nil
			}
		}
		return nil, os.ErrNotExist
	}
	return os.ReadFile(path)
}

func loadHTTP(url string, wantChecksum string) ([]byte, error) {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, &httpError{status: resp.StatusCode, url: url}
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if wantChecksum != "" {
		sum := sha256.Sum256(data)
		got := hex.EncodeToString(sum[:])
		if got != wantChecksum {
			return nil, &checksumMismatchErr{expected: wantChecksum, got: got}
		}
	}
	return data, nil
}

type httpError struct {
	status int
	url    string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("http %d %s", e.status, e.url)
}

type checksumMismatchErr struct {
	expected, got string
}

func (e *checksumMismatchErr) Error() string {
	return "checksum mismatch"
}
