package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gray-snyk/secrets-gen/internal/display"
	"github.com/gray-snyk/secrets-gen/internal/generators"
)

type generatedSecret struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Provider string `json:"provider"`
	Value    string `json:"value"`
	Severity string `json:"severity"`
}

// selectRules resolves which rules to generate from, based on --id or a
// provider name. The provider is matched case-insensitively as a substring of
// the rule's provider field.
func selectRules(rules []generators.SecretRule, provider string) ([]generators.SecretRule, error) {
	if genID != "" {
		for _, r := range rules {
			if r.ID == genID {
				return []generators.SecretRule{r}, nil
			}
		}
		return nil, fmt.Errorf("no rule with id %q", genID)
	}

	needle := strings.ToLower(strings.TrimSpace(provider))
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
		return nil, fmt.Errorf("no rules match provider %q", provider)
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
	prevID := ""
	for i, s := range results {
		if s.ID != prevID {
			if i > 0 {
				fmt.Println()
			}
			header := fmt.Sprintf("%s   %s",
				display.TitleStyle.Render(s.Title),
				display.SeverityStyle(s.Severity).Render(strings.ToUpper(s.Severity)),
			)
			fmt.Println(header)
			prevID = s.ID
		}
		fmt.Println(display.SecretStyle.Render(s.Value))
	}
	fmt.Println()
}

// uniqueProviders returns the sorted set of non-empty provider names.
func uniqueProviders(rules []generators.SecretRule) []string {
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
	return providers
}

func printProviders(rules []generators.SecretRule) {
	for _, p := range uniqueProviders(rules) {
		fmt.Println(p)
	}
}
