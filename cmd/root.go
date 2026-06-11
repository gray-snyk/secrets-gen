package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gray-snyk/secrets-gen/internal/display"
)

// Version is overridden by goreleaser via -ldflags at build time.
var Version = "dev"

var noColor bool

var rootCmd = &cobra.Command{
	Use:     "secrets-gen",
	Short:   "Generate fake secrets in real-world formats for testing secret scanners.",
	Version: Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			display.DisableColor()
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(display.Banner())
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable ANSI color output")
	rootCmd.SetVersionTemplate("secrets-gen {{.Version}}\n")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
