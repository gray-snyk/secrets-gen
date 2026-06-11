package generators

import (
	"fmt"
	"strings"

	"github.com/lucasjones/reggen"
)

const (
	reggenRepeatLimit = 10
	maxSafeAttempts   = 10
)

func GenerateFromRule(rule SecretRule) (string, error) {
	pattern := preprocessRegex(rule.Regex)
	value, err := reggen.Generate(pattern, reggenRepeatLimit)
	if err != nil {
		return "", fmt.Errorf("generate %s: %w", rule.ID, err)
	}
	return value, nil
}

// GenerateSafe retries GenerateFromRule until the output passes IsSafeSecret
// and IsValidForRule, or maxSafeAttempts is exhausted. It exists because
// reggen occasionally emits values containing regex metacharacters or values
// that don't actually match the source pattern.
func GenerateSafe(rule SecretRule) (string, error) {
	var lastErr error
	for i := 0; i < maxSafeAttempts; i++ {
		value, err := GenerateFromRule(rule)
		if err != nil {
			lastErr = err
			continue
		}
		if !IsSafeSecret(value) {
			continue
		}
		if !IsValidForRule(value, rule) {
			continue
		}
		return value, nil
	}
	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("could not produce a safe secret for %s after %d attempts", rule.ID, maxSafeAttempts)
}

func preprocessRegex(pattern string) string {
	pattern = strings.ReplaceAll(pattern, "(?i)", "")
	pattern = strings.TrimPrefix(pattern, "^")
	pattern = strings.TrimSuffix(pattern, "$")
	if wrapsWhole(pattern) {
		pattern = pattern[1 : len(pattern)-1]
	}
	return pattern
}

func wrapsWhole(s string) bool {
	if len(s) < 2 || s[0] != '(' || s[len(s)-1] != ')' {
		return false
	}
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++ // skip the escaped character
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 && i != len(s)-1 {
				return false
			}
		}
	}
	return depth == 0
}
