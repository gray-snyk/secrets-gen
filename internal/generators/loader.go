package generators

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gray-snyk/secrets-gen/assets"
)

// ApprovedProviders is the allow-list of providers the tool will surface.
// Rules for any provider not in this list are dropped at load time.
var ApprovedProviders = []string{
	"AWS", "Anthropic", "Azure", "Bitbucket", "ClickUp",
	"Cloudflare", "Datadog", "Docker", "GitHub", "GitLab",
	"OpenAI", "Stripe", "npm",
}

type SecretRule struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Provider    string   `json:"provider"`
	Prefixes    []string `json:"prefixes"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Regex       string   `json:"regex"`
	Severity    string   `json:"severity"`
}

func LoadRules() ([]SecretRule, error) {
	var rules []SecretRule
	if err := json.Unmarshal(assets.SecretTypeMetadata, &rules); err != nil {
		return nil, fmt.Errorf("parse secret metadata: %w", err)
	}

	approved := rules[:0]
	for _, r := range rules {
		if isApprovedProvider(r.Provider) {
			approved = append(approved, r)
		}
	}
	return approved, nil
}

func isApprovedProvider(provider string) bool {
	for _, p := range ApprovedProviders {
		if strings.EqualFold(p, provider) {
			return true
		}
	}
	return false
}
