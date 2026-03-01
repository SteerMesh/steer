package lock

import (
	"encoding/json"
	"os"
	"sort"
)

const LockfileVersion = "1.0"

// Lockfile is the resolved pack list (matches spec lockfile.schema.json).
type Lockfile struct {
	Version string  `json:"version"`
	Packs   []Entry `json:"packs"`
}

// Entry is a resolved pack entry.
type Entry struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Source   string `json:"source"`
	Checksum string `json:"checksum,omitempty"`
}

// Load reads the lockfile from path.
func Load(path string) (*Lockfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var l Lockfile
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

// Save writes the lockfile to path. Packs are sorted by name for determinism.
func Save(path string, l *Lockfile) error {
	l.Version = LockfileVersion
	sort.Slice(l.Packs, func(i, j int) bool { return l.Packs[i].Name < l.Packs[j].Name })
	raw, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0644)
}

// AddOrUpdate sets or updates the entry for the given name (exact version, source, optional checksum).
func (l *Lockfile) AddOrUpdate(name, version, source, checksum string) {
	for i := range l.Packs {
		if l.Packs[i].Name == name {
			l.Packs[i].Version = version
			l.Packs[i].Source = source
			l.Packs[i].Checksum = checksum
			return
		}
	}
	l.Packs = append(l.Packs, Entry{Name: name, Version: version, Source: source, Checksum: checksum})
}

// Get returns the entry for the given pack name, or nil.
func (l *Lockfile) Get(name string) *Entry {
	for i := range l.Packs {
		if l.Packs[i].Name == name {
			return &l.Packs[i]
		}
	}
	return nil
}

// DriftReport describes a difference between config constraints and lockfile.
type DriftReport struct {
	MissingInLock []string // pack names in config but not in lock
	MissingInConfig []string // pack names in lock but not in config (optional to report)
	VersionMismatch []string // pack name where resolved version differs from lock
}

// Drift compares config pack names (and optionally version constraints) to the lockfile.
// It reports packs that are in config but not in lock, and optionally version mismatches.
func (l *Lockfile) Drift(configPackNames []string) DriftReport {
	namesInLock := make(map[string]string)
	for _, e := range l.Packs {
		namesInLock[e.Name] = e.Version
	}
	namesInConfig := make(map[string]bool)
	for _, n := range configPackNames {
		namesInConfig[n] = true
	}
	var missingInLock, missingInConfig, versionMismatch []string
	for _, n := range configPackNames {
		if _, ok := namesInLock[n]; !ok {
			missingInLock = append(missingInLock, n)
		}
	}
	for n := range namesInLock {
		if !namesInConfig[n] {
			missingInConfig = append(missingInConfig, n)
		}
	}
	sort.Strings(missingInLock)
	sort.Strings(missingInConfig)
	return DriftReport{
		MissingInLock:   missingInLock,
		MissingInConfig: missingInConfig,
		VersionMismatch: versionMismatch,
	}
}
