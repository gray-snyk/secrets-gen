package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/gray-snyk/secrets-gen/internal/display"
	"github.com/gray-snyk/secrets-gen/internal/generators"
)

// Version is overridden by goreleaser via -ldflags at build time.
var Version = "dev"

// errSilent signals that a user-facing message has already been printed and
// the top-level handler should exit non-zero without printing anything more.
var errSilent = errors.New("handled")

var (
	noColor          bool
	genCount         int
	genCopy          bool
	genFormat        string
	genID            string
	genListProviders bool
)

var rootCmd = &cobra.Command{
	Use:   "secrets-gen [provider]",
	Short: "Generate fake secrets in real-world formats for testing secret scanners.",
	Long: "Generate fake secrets in real-world formats for testing secret scanners.\n\n" +
		"Run with no arguments to launch an interactive provider picker, or pass a\n" +
		"provider name to generate secrets directly (e.g. `secrets-gen github`).",
	Version: Version,
	Args:    cobra.MaximumNArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			display.DisableColor()
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		rules, err := generators.LoadRules()
		if err != nil {
			return err
		}

		if genListProviders {
			printProviders(rules)
			return nil
		}

		// Resolve the provider: an explicit --id needs none, a positional
		// argument names one directly, and otherwise we fall back to the
		// interactive picker.
		provider := ""
		if len(args) == 1 {
			provider = args[0]
		}
		if genID == "" && provider == "" {
			chosen, ok, err := runPicker(uniqueProviders(rules))
			if err != nil {
				return err
			}
			if !ok {
				// User quit the picker without selecting — exit cleanly.
				return nil
			}
			provider = chosen
		}

		return runGeneration(rules, provider)
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func runGeneration(rules []generators.SecretRule, provider string) error {
	matched, err := selectRules(rules, provider)
	if err != nil {
		return err
	}

	results := generateAll(matched, genCount)
	if len(results) == 0 {
		return fmt.Errorf("no secrets generated (every matching rule failed)")
	}

	if genFormat == "json" {
		return writeJSON(results)
	}

	writeStyled(results)

	if genCopy {
		if err := clipboard.WriteAll(results[0].Value); err != nil {
			fmt.Fprintln(os.Stderr, "warning: clipboard unavailable:", err)
		} else {
			fmt.Println(display.CheckStyle.Render("✓ Copied to clipboard"))
		}
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable ANSI color output")
	rootCmd.Flags().IntVarP(&genCount, "count", "n", 1, "number of secrets to generate per matching rule")
	rootCmd.Flags().BoolVarP(&genCopy, "copy", "c", false, "copy the first generated secret to the clipboard")
	rootCmd.Flags().StringVar(&genFormat, "format", "", "output format: json (omit for styled output)")
	rootCmd.Flags().StringVar(&genID, "id", "", "generate a secret for a specific rule ID, skipping provider matching")
	rootCmd.Flags().BoolVar(&genListProviders, "list-providers", false, "print all unique providers and exit")
	rootCmd.SetVersionTemplate("secrets-gen {{.Version}}\n")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if !errors.Is(err, errSilent) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
