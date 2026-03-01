package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/SteerMesh/steer/internal/config"
)

const apiKeyHeader = "X-API-Key"

// Client calls the SteerMesh Cloud API.
type Client struct {
	BaseURL    string
	APIKey     string
	ProjectID  string
	HTTPClient *http.Client
}

// NewClientFromConfig builds a client from project config and env.
// Env STEER_API_URL, STEER_API_KEY, STEER_PROJECT_ID override config.
func NewClientFromConfig(proj *config.Project) (*Client, error) {
	c := &Client{HTTPClient: http.DefaultClient}
	if proj != nil && proj.Cloud != nil {
		c.BaseURL = strings.TrimSuffix(proj.Cloud.APIURL, "/")
		c.APIKey = proj.Cloud.APIKey
		c.ProjectID = proj.Cloud.ProjectID
	}
	if v := os.Getenv("STEER_API_URL"); v != "" {
		c.BaseURL = strings.TrimSuffix(v, "/")
	}
	if v := os.Getenv("STEER_API_KEY"); v != "" {
		c.APIKey = v
	}
	if v := os.Getenv("STEER_PROJECT_ID"); v != "" {
		c.ProjectID = v
	}
	if c.BaseURL == "" || c.APIKey == "" {
		return nil, fmt.Errorf("sync requires apiUrl and apiKey (config or STEER_API_URL, STEER_API_KEY)")
	}
	return c, nil
}

// BundleManifest is the API response for a bundle.
type BundleManifest struct {
	ID       string        `json:"id"`
	Manifest *ManifestBody `json:"manifest"`
	Files    []FileRef     `json:"files,omitempty"`
}

type ManifestBody struct {
	Version string     `json:"version"`
	Packs   []PackRef  `json:"packs"`
	Files   []FileEntry `json:"files"`
}

type PackRef struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type FileEntry struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type FileRef struct {
	Path string `json:"path"`
	URL  string `json:"url,omitempty"`
}

// PullLatest fetches the project's latest bundle and writes it to outDir.
// If the API returns no bundles, returns without error (no-op).
func (c *Client) PullLatest(outDir string) error {
	if c.ProjectID == "" {
		return fmt.Errorf("projectId required for sync (config or STEER_PROJECT_ID)")
	}
	// GET /projects/{id}/bundles, pick latest, then GET /bundles/{id}
	req, _ := http.NewRequest("GET", c.BaseURL+"/projects/"+c.ProjectID+"/bundles", nil)
	req.Header.Set(apiKeyHeader, c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("projects/bundles: %s %s", resp.Status, string(body))
	}
	var list struct {
		Bundles []struct {
			ID string `json:"id"`
		} `json:"bundles"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return err
	}
	if len(list.Bundles) == 0 {
		return nil
	}
	latestID := list.Bundles[len(list.Bundles)-1].ID
	return c.PullBundle(latestID, outDir)
}

// PullBundle fetches a bundle by ID and writes manifest and files to outDir.
func (c *Client) PullBundle(bundleID, outDir string) error {
	req, _ := http.NewRequest("GET", c.BaseURL+"/bundles/"+bundleID, nil)
	req.Header.Set(apiKeyHeader, c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bundles/%s: %s %s", bundleID, resp.Status, string(body))
	}
	var b BundleManifest
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	if b.Manifest != nil {
		manifestBytes, _ := json.MarshalIndent(b.Manifest, "", "  ")
		if err := os.WriteFile(filepath.Join(outDir, "bundle-manifest.json"), manifestBytes, 0644); err != nil {
			return err
		}
	}
	for _, f := range b.Files {
		if f.URL == "" {
			continue
		}
		if err := c.downloadFile(f.URL, filepath.Join(outDir, f.Path)); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) downloadFile(url, path string) error {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set(apiKeyHeader, c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: %s", url, resp.Status)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	return err
}
