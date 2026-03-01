package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SteerMesh/steer/internal/bundle"
	"github.com/SteerMesh/steer/internal/sign"
	"github.com/spf13/cobra"
)

var (
	verifyBundleManifest string
	verifyBundlePubKey   string
)

var verifyBundleCmd = &cobra.Command{
	Use:   "verify-bundle",
	Short: "Verify bundle manifest signature",
	Long:  "Reads bundle-manifest.json; if it has a signature, verifies it with the given public key.",
	RunE:  runVerifyBundle,
}

func init() {
	verifyBundleCmd.Flags().StringVar(&verifyBundleManifest, "manifest", "", "Path to bundle-manifest.json (default: .steer/output/bundle-manifest.json)")
	verifyBundleCmd.Flags().StringVar(&verifyBundlePubKey, "public-key", "", "Path to Ed25519 public key PEM (required if manifest is signed)")
}

func runVerifyBundle(cmd *cobra.Command, args []string) error {
	manifestPath := verifyBundleManifest
	if manifestPath == "" {
		cwd, _ := os.Getwd()
		manifestPath = filepath.Join(cwd, defaultOutputDir, "bundle-manifest.json")
	}
	m, err := bundle.LoadManifest(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("manifest not found: %s", manifestPath)
		}
		return err
	}
	if m.Signature == nil {
		fmt.Println("Manifest has no signature; nothing to verify.")
		return nil
	}
	if verifyBundlePubKey == "" {
		return usageErr("manifest is signed; provide --public-key to verify")
	}
	canonical, err := bundle.CanonicalBytes(m)
	if err != nil {
		return err
	}
	if err := sign.Verify(canonical, m.Signature, verifyBundlePubKey); err != nil {
		return err
	}
	fmt.Println("Signature valid.")
	return nil
}
