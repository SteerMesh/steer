package cli

import (
	"log/slog"
	"os"

	"github.com/SteerMesh/steer/internal/compiler"
	"github.com/SteerMesh/steer/internal/config"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [pack.yaml]",
	Short: "Validate project config and pack YAML against spec",
	Long:  "Loads project config and pack YAML(s), validates against spec schemas. Exit 0 only if valid.",
	RunE:  runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	if len(args) >= 1 {
		for _, path := range args {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if _, err := compiler.ParsePack(data); err != nil {
				slog.Error("validation failed", "file", path, "error", err)
				return err
			}
		}
		return nil
	}
	_, configPath, err := config.LoadFromDir(cwd)
	if err != nil {
		return err
	}
	slog.Debug("config loaded", "path", configPath)
	// If no pack paths given, try default pack.yaml in cwd
	packPath := "pack.yaml"
	if _, err := os.Stat(packPath); err != nil {
		return nil // no pack file to validate
	}
	data, err := os.ReadFile(packPath)
	if err != nil {
		return err
	}
	_, err = compiler.ParsePack(data)
	return err
}
