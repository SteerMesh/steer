package cli

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/SteerMesh/steer/internal/config"
	"github.com/SteerMesh/steer/internal/lock"
	"github.com/SteerMesh/steer/internal/packload"
	"github.com/SteerMesh/steer/internal/registry"
	"github.com/SteerMesh/steer/internal/resolver"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var addCmd = &cobra.Command{
	Use:   "add [pack@version]",
	Short: "Add a pack and update lockfile",
	Long:  "Add pack to config, resolve version, update lockfile. Example: steer add security-core@1.0.0",
	RunE:  runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return usageErr("usage: steer add <pack@version>")
	}
	cwd, _ := os.Getwd()
	configPath := filepath.Join(cwd, "steer.yaml")
	proj, _, err := config.LoadFromDir(cwd)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(configPath, []byte(defaultSteerYAML), 0644); err != nil {
				return err
			}
			proj = &config.Project{}
		} else {
			return err
		}
	}

	name, constraint, err := parsePackRef(args[0])
	if err != nil {
		return err
	}
	proj.Packs = append(proj.Packs, config.PackRef{Name: name, Version: constraint})

	// Resolve version and source
	registryURL := proj.RegistryURL
	if registryURL == "" {
		registryURL = os.Getenv("STEER_REGISTRY_URL")
	}
	var resolvedVersion, source string
	if registryURL != "" {
		reg := registry.NewClient(registryURL)
		resolvedVersion, contentURL, err := reg.Resolve(name, constraint)
		if err != nil {
			return err
		}
		// Download to packs/<name>/pack.yaml (option B: reproducible, offline compile)
		packDir := filepath.Join(cwd, defaultPacksDir, name)
		if err := os.MkdirAll(packDir, 0755); err != nil {
			return err
		}
		data, err := packload.Load(contentURL, cwd, "")
		if err != nil {
			return err
		}
		packPath := filepath.Join(packDir, "pack.yaml")
		if err := os.WriteFile(packPath, data, 0644); err != nil {
			return err
		}
		slog.Info("pack downloaded", "name", name, "version", resolvedVersion)
		source = "file://" + filepath.Join(defaultPacksDir, name)
	} else {
		// Local: resolve constraint against local pack if present
		resolvedVersion = constraint
		localPath := filepath.Join(cwd, defaultPacksDir, name)
		if data, err := packload.Load("file://"+localPath, cwd, ""); err == nil {
			// Parse pack to get version for semver resolution
			var raw struct {
				Pack struct {
					Version string `yaml:"version"`
				} `yaml:"pack"`
			}
			if err := yaml.Unmarshal(data, &raw); err == nil && raw.Pack.Version != "" {
				versions := []string{raw.Pack.Version}
				if v, err := resolver.ResolveConstraint(constraint, versions); err == nil {
					resolvedVersion = v
				}
			}
		}
		source = "file://" + filepath.Join(defaultPacksDir, name)
	}

	// Write config back
	out, err := yaml.Marshal(proj)
	if err != nil {
		return err
	}
	if err := os.WriteFile(configPath, out, 0644); err != nil {
		return err
	}

	lockPath := filepath.Join(cwd, defaultLockfile)
	lf, _ := lock.Load(lockPath)
	if lf == nil {
		lf = &lock.Lockfile{}
	}
	lf.AddOrUpdate(name, resolvedVersion, source, "")
	if err := lock.Save(lockPath, lf); err != nil {
		return err
	}
	return nil
}

func parsePackRef(s string) (name, version string, err error) {
	i := strings.Index(s, "@")
	if i < 0 {
		return "", "", fmt.Errorf("invalid pack ref %q (expected name@version)", s)
	}
	name, version = strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:])
	if name == "" || version == "" {
		return "", "", fmt.Errorf("invalid pack ref %q", s)
	}
	return name, version, nil
}
