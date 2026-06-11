package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/gray-snyk/secrets-gen/internal/display"
	"github.com/gray-snyk/secrets-gen/internal/generators"
)

var (
	genCount         int
	genCopy          bool
	genFormat        string
	genID            string
	genListProviders bool
)

type generatedSecret struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Provider string `json:"provider"`
	Value    string `json:"value"`
	Severity string `json:"severity"`
}

var generateCmd = &cobra.Command{
	Use:   "generate [provider]",
	Short: "Generate fake secrets matching a provider's known formats.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rules, err := generators.LoadRules()
		if err != nil {
			return err
		}

		if genListProviders {
			printProviders(rules)
			return nil
		}

		matched, err := selectRules(rules, args)
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
	},
}

func init() {
	generateCmd.Flags().IntVarP(&genCount, "count", "n", 1, "number of secrets to generate per matching rule")
	generateCmd.Flags().BoolVarP(&genCopy, "copy", "c", false, "copy the first generated secret to the clipboard")
	generateCmd.Flags().StringVar(&genFormat, "format", "", "output format: json (omit for styled output)")
	generateCmd.Flags().StringVar(&genID, "id", "", "generate a secret for a specific rule ID, skipping provider matching")
	generateCmd.Flags().BoolVar(&genListProviders, "list-providers", false, "print all unique providers and exit")
	rootCmd.AddCommand(generateCmd)
}

func selectRules(rules []generators.SecretRule, args []string) ([]generators.SecretRule, error) {
	if genID != "" {
		for _, r := range rules {
			if r.ID == genID {
				return []generators.SecretRule{r}, nil
			}
		}
		return nil, fmt.Errorf("no rule with id %q", genID)
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("provider argument required (or use --id / --list-providers)")
	}
	needle := strings.ToLower(strings.TrimSpace(args[0]))
	if needle == "" {
		return nil, fmt.Errorf("provider argument cannot be empty")
	}

	var matched []generators.SecretRule
	for _, r := range rules {
		if strings.Contains(strings.ToLower(r.Provider), needle) {
			matched = append(matched, r)
		}
	}
	if len(matched) == 0 {
		return nil, fmt.Errorf("no rules match provider %q", args[0])
	}
	return matched, nil
}

func generateAll(rules []generators.SecretRule, count int) []generatedSecret {
	if count < 1 {
		count = 1
	}
	out := make([]generatedSecret, 0, len(rules)*count)
	for _, r := range rules {
		for i := 0; i < count; i++ {
			value, err := generators.GenerateSafe(r)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", r.ID, err)
				break
			}
			out = append(out, generatedSecret{
				ID:       r.ID,
				Title:    r.Title,
				Provider: r.Provider,
				Value:    value,
				Severity: r.Severity,
			})
		}
	}
	return out
}

func writeJSON(results []generatedSecret) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func writeStyled(results []generatedSecret) {
	for _, s := range results {
		header := fmt.Sprintf("%s   %s",
			display.TitleStyle.Render(s.Title),
			display.SeverityStyle(s.Severity).Render(strings.ToUpper(s.Severity)),
		)
		fmt.Println(header)
		fmt.Println(display.SecretStyle.Render(s.Value))
		fmt.Println()
	}
}

func printProviders(rules []generators.SecretRule) {
	seen := make(map[string]struct{}, len(rules))
	for _, r := range rules {
		if r.Provider == "" {
			continue
		}
		seen[r.Provider] = struct{}{}
	}
	providers := make([]string, 0, len(seen))
	for p := range seen {
		providers = append(providers, p)
	}
	sort.Strings(providers)
	for _, p := range providers {
		fmt.Println(p)
	}
}
