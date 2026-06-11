package generators

import (
	"encoding/json"
	"fmt"

	"github.com/gray-snyk/secrets-gen/assets"
)

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
	return rules, nil
}
