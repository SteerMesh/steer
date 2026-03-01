package cli

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/SteerMesh/steer/internal/config"
	"github.com/SteerMesh/steer/internal/sync"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync with SteerMesh Cloud",
	Long:  "Pull latest bundle from Cloud API into .steer/output; requires cloud.apiUrl, cloud.apiKey, cloud.projectId in steer.yaml or STEER_API_URL, STEER_API_KEY, STEER_PROJECT_ID.",
	RunE:  runSync,
}

func runSync(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	proj, _, err := config.LoadFromDir(cwd)
	if err != nil {
		return err
	}
	client, err := sync.NewClientFromConfig(proj)
	if err != nil {
		return err
	}
	outDir := filepath.Join(cwd, defaultOutputDir)
	if err := client.PullLatest(outDir); err != nil {
		return err
	}
	slog.Info("sync done", "output", outDir)
	return nil
}
