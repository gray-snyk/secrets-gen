package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"

	"github.com/gray-snyk/secrets-gen/internal/display"
	"github.com/gray-snyk/secrets-gen/internal/generators"
)

var (
	listProvider string
	listSeverity string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all known secret rules with optional filtering.",
	RunE: func(cmd *cobra.Command, args []string) error {
		rules, err := generators.LoadRules()
		if err != nil {
			return err
		}

		filtered := filterRules(rules, listProvider, listSeverity)
		if len(filtered) == 0 {
			fmt.Fprintln(os.Stderr, "no rules match the given filters")
			return nil
		}

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(display.LabelStyle).
			Headers("ID", "TITLE", "PROVIDER", "SEVERITY", "PREFIX").
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return display.LabelStyle.Bold(true).Padding(0, 1)
				}
				if col == 3 && row >= 0 && row < len(filtered) {
					return display.SeverityStyle(filtered[row].Severity).Padding(0, 1)
				}
				return lipgloss.NewStyle().Padding(0, 1)
			})

		for _, r := range filtered {
			t.Row(r.ID, r.Title, r.Provider, strings.ToUpper(r.Severity), firstPrefix(r.Prefixes))
		}

		fmt.Println(t.Render())
		fmt.Println(display.LabelStyle.Render(fmt.Sprintf("%d rule(s)", len(filtered))))
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listProvider, "provider", "", "filter by provider (case-insensitive substring)")
	listCmd.Flags().StringVar(&listSeverity, "severity", "", "filter by severity (CRITICAL, HIGH, MEDIUM, LOW)")
	rootCmd.AddCommand(listCmd)
}

func filterRules(rules []generators.SecretRule, provider, severity string) []generators.SecretRule {
	provider = strings.ToLower(strings.TrimSpace(provider))
	severity = strings.ToUpper(strings.TrimSpace(severity))

	out := make([]generators.SecretRule, 0, len(rules))
	for _, r := range rules {
		if provider != "" && !strings.Contains(strings.ToLower(r.Provider), provider) {
			continue
		}
		if severity != "" && strings.ToUpper(r.Severity) != severity {
			continue
		}
		out = append(out, r)
	}
	return out
}

func firstPrefix(prefixes []string) string {
	if len(prefixes) == 0 {
		return ""
	}
	return prefixes[0]
}
