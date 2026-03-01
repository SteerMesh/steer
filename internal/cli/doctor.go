package cli

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/SteerMesh/steer/internal/bundle"
	"github.com/SteerMesh/steer/internal/config"
	"github.com/SteerMesh/steer/internal/lock"
	"github.com/SteerMesh/steer/internal/sign"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check project and environment",
	Long:  "Checks env, config, lockfile, and bundle consistency; reports issues (missing deps, drift, invalid paths).",
	RunE:  runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	proj, configPath, err := config.LoadFromDir(cwd)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No steer.yaml or steering.yaml found in current directory.")
			return nil
		}
		return err
	}
	fmt.Printf("Config: %s\n", configPath)

	lockPath := filepath.Join(cwd, defaultLockfile)
	lf, err := lock.Load(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No steer.lock found.")
			return nil
		}
		return err
	}

	names := make([]string, 0, len(proj.Packs))
	for _, p := range proj.Packs {
		names = append(names, p.Name)
	}
	report := lf.Drift(names)
	ok := true
	if len(report.MissingInLock) > 0 {
		slog.Warn("packs in config but not in lockfile", "packs", report.MissingInLock)
		fmt.Printf("Drift: packs not in lockfile: %v\n", report.MissingInLock)
		ok = false
	}
	if len(report.MissingInConfig) > 0 {
		slog.Info("packs in lockfile but not in config", "packs", report.MissingInConfig)
		fmt.Printf("Note: lockfile has extra packs: %v\n", report.MissingInConfig)
	}
	// Optional: verify bundle manifest signature if present and STEER_SIGNATURE_PUBLIC_KEY is set
	manifestPath := filepath.Join(cwd, defaultOutputDir, "bundle-manifest.json")
	if pubKeyPath := os.Getenv("STEER_SIGNATURE_PUBLIC_KEY"); pubKeyPath != "" {
		m, err := bundle.LoadManifest(manifestPath)
		if err == nil && m.Signature != nil {
			canonical, err := bundle.CanonicalBytes(m)
			if err != nil {
				slog.Warn("doctor: could not get canonical manifest", "error", err)
			} else if err := sign.Verify(canonical, m.Signature, pubKeyPath); err != nil {
				slog.Warn("bundle signature verification failed", "error", err)
				fmt.Printf("Bundle signature: invalid (%v)\n", err)
				ok = false
			} else {
				fmt.Println("Bundle signature: valid.")
			}
		}
	}
	if ok {
		fmt.Println("Doctor: no issues found.")
	}
	return nil
}
