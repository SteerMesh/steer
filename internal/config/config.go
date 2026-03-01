// Package config loads and parses project configuration (steer.yaml).
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Project is the root project configuration.
type Project struct {
	Packs       []PackRef `yaml:"packs"`
	Targets     []string  `yaml:"targets,omitempty"`     // enabled target ids (empty = all from packs)
	Overrides   string    `yaml:"overrides,omitempty"`  // optional path to overrides
	RegistryURL string    `yaml:"registryUrl,omitempty"` // optional pack registry base URL
	Cloud       *Cloud    `yaml:"cloud,omitempty"`       // optional Cloud API for sync
}

// Cloud holds Cloud API connection settings.
type Cloud struct {
	APIURL    string `yaml:"apiUrl,omitempty"`    // e.g. https://api.steermesh.dev
	APIKey    string `yaml:"apiKey,omitempty"`    // X-API-Key value
	ProjectID string `yaml:"projectId,omitempty"` // project to sync
}

// PackRef references a pack by name and version constraint.
type PackRef struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"` // semver constraint, e.g. 1.0.0, ^1.0.0
}

// DefaultConfigFilenames are tried in order when loading.
var DefaultConfigFilenames = []string{"steer.yaml", "steering.yaml"}

// Load reads and parses the project config from the given path.
func Load(path string) (*Project, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p Project
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// LoadFromDir finds steer.yaml or steering.yaml in dir and loads it.
func LoadFromDir(dir string) (*Project, string, error) {
	for _, name := range DefaultConfigFilenames {
		path := dir + string(os.PathSeparator) + name
		if _, err := os.Stat(path); err == nil {
			p, err := Load(path)
			return p, path, err
		}
	}
	return nil, "", os.ErrNotExist
}
