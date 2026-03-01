package cli

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/SteerMesh/steer/internal/bundle"
	"github.com/SteerMesh/steer/internal/compiler"
	"github.com/SteerMesh/steer/internal/config"
	"github.com/SteerMesh/steer/internal/lock"
	"github.com/SteerMesh/steer/internal/packload"
	"github.com/SteerMesh/steer/internal/registry"
	"github.com/SteerMesh/steer/internal/resolver"
	"github.com/SteerMesh/steer/internal/sign"
	"github.com/SteerMesh/steer/internal/targets"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	compileSign    bool
	compileSignKey string
)

func init() {
	compileCmd.Flags().BoolVar(&compileSign, "sign", false, "Sign the bundle manifest after compile (requires --sign-key)")
	compileCmd.Flags().StringVar(&compileSignKey, "sign-key", "", "Path to Ed25519 private key (PEM PKCS#8) for signing")
}

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile steering packs into target artifacts",
	Long:  "Loads config, resolves packs (from lock or resolve), runs compiler and target renderers, writes bundle and manifest.",
	RunE:  runCompile,
}

func runCompile(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	proj, _, err := config.LoadFromDir(cwd)
	if err != nil {
		return err
	}

	if len(proj.Packs) == 0 {
		// Backward compat: single pack.yaml in cwd
		packPath := filepath.Join(cwd, defaultPackPath)
		if data, err := os.ReadFile(packPath); err == nil {
			model, err := compiler.ParsePack(data)
			if err != nil {
				return err
			}
			return runCompileOne(cwd, proj, model, []bundle.PackRef{{Name: model.Pack.Name, Version: model.Pack.Version}})
		}
		return usageErr("no packs in steer.yaml and no pack.yaml in current directory")
	}

	lockPath := filepath.Join(cwd, defaultLockfile)
	lf, err := lock.Load(lockPath)
	if err != nil {
		lf = &lock.Lockfile{}
	}

	var models []*compiler.Model
	var packRefs []bundle.PackRef
	lockUpdated := false

	registryURL := proj.RegistryURL
	if registryURL == "" {
		registryURL = os.Getenv("STEER_REGISTRY_URL")
	}

	for _, ref := range proj.Packs {
		entry := lf.Get(ref.Name)
		source := ""
		version := ref.Version
		if entry != nil {
			source = entry.Source
			version = entry.Version
		} else {
			if registryURL != "" {
				reg := registry.NewClient(registryURL)
				resolvedVersion, contentURL, err := reg.Resolve(ref.Name, ref.Version)
				if err != nil {
					return err
				}
				packDir := filepath.Join(cwd, defaultPacksDir, ref.Name)
				if err := os.MkdirAll(packDir, 0755); err != nil {
					return err
				}
				data, err := packload.Load(contentURL, cwd, "")
				if err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(packDir, "pack.yaml"), data, 0644); err != nil {
					return err
				}
				source = "file://" + filepath.Join(defaultPacksDir, ref.Name)
				version = resolvedVersion
			} else {
				// Local: resolve constraint against local pack if present
				source = "file://" + filepath.Join(defaultPacksDir, ref.Name)
				if data, err := packload.Load(source, cwd, ""); err == nil {
					var raw struct {
						Pack struct {
							Version string `yaml:"version"`
						} `yaml:"pack"`
					}
					if err := yaml.Unmarshal(data, &raw); err == nil && raw.Pack.Version != "" {
						if v, err := resolver.ResolveConstraint(ref.Version, []string{raw.Pack.Version}); err == nil {
							version = v
						}
					}
				}
			}
			lf.AddOrUpdate(ref.Name, version, source, "")
			lockUpdated = true
		}
		checksum := ""
		if entry != nil {
			checksum = entry.Checksum
		}
		data, err := packload.Load(source, cwd, checksum)
		if err != nil {
			return err
		}
		model, err := compiler.ParsePack(data)
		if err != nil {
			return err
		}
		models = append(models, model)
		packRefs = append(packRefs, bundle.PackRef{Name: model.Pack.Name, Version: model.Pack.Version})
	}

	if lockUpdated {
		if err := lock.Save(lockPath, lf); err != nil {
			slog.Warn("could not save lockfile", "error", err)
		}
	}

	merged := compiler.Merge(models)
	return runCompileOne(cwd, proj, merged, packRefs)
}

func runCompileOne(cwd string, proj *config.Project, model *compiler.Model, packRefs []bundle.PackRef) error {
	registry := targets.Registry()
	selected := make(map[string]targets.Renderer)
	for _, r := range registry {
		if len(proj.Targets) == 0 {
			selected[r.Name()] = r
		} else {
			for _, t := range proj.Targets {
				if t == r.Name() {
					selected[r.Name()] = r
					break
				}
			}
		}
	}

	ctx := context.Background()
	allFiles := make(map[string][]byte)
	for _, r := range selected {
		files, err := r.Render(ctx, model)
		if err != nil {
			return err
		}
		for k, v := range files {
			allFiles[k] = v
		}
	}

	manifest := bundle.Build(allFiles, packRefs)
	outDir := filepath.Join(cwd, defaultOutputDir)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	if err := bundle.Write(outDir, allFiles, manifest); err != nil {
		return err
	}
	if compileSign {
		if compileSignKey == "" {
			return usageErr("--sign requires --sign-key (path to Ed25519 private key PEM)")
		}
		canonical, err := bundle.CanonicalBytes(manifest)
		if err != nil {
			return err
		}
		algo, keyID, value, err := sign.Sign(canonical, compileSignKey)
		if err != nil {
			return err
		}
		manifest.Signature = &bundle.Signature{Algorithm: algo, KeyID: keyID, Value: value}
		manifestPath := filepath.Join(outDir, "bundle-manifest.json")
		raw, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(manifestPath, raw, 0644); err != nil {
			return err
		}
		slog.Info("bundle manifest signed", "keyId", keyID)
	}
	slog.Info("compile done", "output", outDir)
	return nil
}
