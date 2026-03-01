package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/SteerMesh/steer/internal/resolver"
)

// Index is the registry index format: packs by name with list of versions.
type Index struct {
	Packs []PackMeta `json:"packs"`
}

// PackMeta is a pack entry in the index.
type PackMeta struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
	// Optional: base URL for pack content (default: registry base + /packs/<name>/<version>/pack.yaml)
	ContentURLTemplate string `json:"contentUrlTemplate,omitempty"`
}

// Client fetches registry index and resolves pack versions.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	index      *Index
}

// NewClient returns a registry client. baseURL should not have trailing slash.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    strings.TrimSuffix(baseURL, "/"),
		HTTPClient: http.DefaultClient,
	}
}

// FetchIndex fetches and parses the registry index (e.g. GET baseURL/index.json).
func (c *Client) FetchIndex() (*Index, error) {
	url := c.BaseURL + "/index.json"
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry index: %s", resp.Status)
	}
	var idx Index
	if err := json.NewDecoder(resp.Body).Decode(&idx); err != nil {
		return nil, err
	}
	c.index = &idx
	return &idx, nil
}

// Resolve returns the resolved version and content URL for the given name and constraint.
// FetchIndex must have been called first, or Resolve will return an error.
func (c *Client) Resolve(name, constraint string) (version, contentURL string, err error) {
	if c.index == nil {
		if _, err := c.FetchIndex(); err != nil {
			return "", "", err
		}
	}
	var versions []string
	for _, p := range c.index.Packs {
		if p.Name == name {
			versions = p.Versions
			break
		}
	}
	if len(versions) == 0 {
		return "", "", fmt.Errorf("pack %q not found in registry", name)
	}
	version, err = resolver.ResolveConstraint(constraint, versions)
	if err != nil {
		return "", "", err
	}
	contentURL = c.contentURL(name, version)
	return version, contentURL, nil
}

func (c *Client) contentURL(name, version string) string {
	for _, p := range c.index.Packs {
		if p.Name == name && p.ContentURLTemplate != "" {
			return strings.ReplaceAll(strings.ReplaceAll(p.ContentURLTemplate, "<name>", name), "<version>", version)
		}
	}
	return fmt.Sprintf("%s/packs/%s/%s/pack.yaml", c.BaseURL, name, version)
}
