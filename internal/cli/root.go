package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "steer",
	Short: "SteerMesh CLI — steering compiler and sync client",
	Long:  "SteerMesh enables tool-agnostic AI steering packs and compiles them into tool-specific formats (Kiro, Cursor, Amazon Q, etc.).",
}

// Execute runs the root command. Returns an error for programmatic use; exit code is set via ExitCode().
func Execute() error {
	return rootCmd.Execute()
}

// Exit codes: 0 success, 1 validation/user error, 2 runtime/internal error.
const (
	ExitSuccess       = 0
	ExitValidation    = 1
	ExitRuntime       = 2
)

// ErrRuntime is a sentinel for runtime errors (exit 2).
type ErrRuntime struct{ Err error }

func (e ErrRuntime) Error() string { return e.Err.Error() }
func (e ErrRuntime) Unwrap() error { return e.Err }

// ExitCode maps known error types to exit codes.
func ExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}
	if _, ok := err.(ErrRuntime); ok {
		return ExitRuntime
	}
	return ExitValidation
}

func init() {
	level := slog.LevelInfo
	if os.Getenv("STEER_DEBUG") != "" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(compileCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(verifyBundleCmd)
}

func usageErr(msg string) error {
	return fmt.Errorf("%s", msg)
}
