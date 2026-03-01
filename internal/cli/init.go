package cli

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const defaultSteerYAML = `# SteerMesh project config
# https://steermesh.dev

packs: []
# Example:
#   - name: security-core
#     version: "1.0.0"

targets: []
# Optional: list of target ids to build (empty = all)
#   - kiro
#   - cursor
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a SteerMesh project",
	Long:  "Creates default steer.yaml and optionally .steer/ in the current directory. Idempotent.",
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	path := filepath.Join(cwd, "steer.yaml")
	if _, err := os.Stat(path); err == nil {
		// Idempotent: already exists, do not overwrite
		return nil
	}
	if err := os.WriteFile(path, []byte(defaultSteerYAML), 0644); err != nil {
		return err
	}
	// Optional .steer directory
	steerDir := filepath.Join(cwd, ".steer")
	_ = os.MkdirAll(steerDir, 0755)
	return nil
}

