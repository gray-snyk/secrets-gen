package generators

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/lucasjones/reggen"
)

const (
	reggenRepeatLimit = 10
	maxSafeAttempts   = 25
)

// Character sets for the manual / custom generators.
const (
	alphanumericChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	hexChars          = "0123456789abcdef"
	// safeChars mirrors the OpenAI key charset [A-Za-z0-9_-]; none of these are
	// rejected by IsSafeSecret.
	safeChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
)

// customGenerators maps a rule ID to a hand-written generator for rules whose
// source regex reggen cannot handle well, or where the realistic format is
// simpler/more consistent than the raw regex. These take precedence over
// GenerateFromRule and are authoritative for the rule's format (see
// IsValidForRule).
var customGenerators = map[string]func() string{
	"datadog-application-key": func() string {
		// Datadog app key: 40 hex chars.
		return randomHex(40)
	},
	"aws-secret-key": func() string {
		// AWS secret: 40 chars, alphanumeric + _ only (no base64 +,=,/).
		return randomAlphanumericWithChars(40, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_")
	},
	"aws-amazon-bedrock-long-term-api-key": func() string {
		// ABSK + 36 alphanumeric.
		return "ABSK" + randomAlphanumeric(36)
	},
	"aws-amazon-bedrock-short-term-api-key": func() string {
		// Fixed prefix + 16 alphanumeric.
		return "bedrock-api-key-YmVkcm9jay5hbWF6b25hd3MuY29t" + randomAlphanumeric(16)
	},
	"openai-api-key": func() string {
		if randomBool() {
			// Legacy format: sk- + 74 alphanumeric (~77 chars).
			return "sk-" + randomAlphanumeric(74)
		}
		// Project format: sk-proj- + 74 chars + T3BlbkFJ + 74 chars (~164 chars).
		return "sk-proj-" + randomAlphanumericWithChars(74, safeChars) +
			"T3BlbkFJ" + randomAlphanumericWithChars(74, safeChars)
	},
	// _gitlab_session=<id> would embed a literal '=', which is now a forbidden
	// character; emit only the realistic 32-char session id (matches the rule's
	// [0-9a-z]{32} value portion).
	"gitlab-session-cookie": func() string {
		return randomAlphanumericWithChars(32, "0123456789abcdefghijklmnopqrstuvwxyz")
	},
}

func GenerateFromRule(rule SecretRule) (string, error) {
	pattern := preprocessRegex(rule.Regex)
	value, err := reggen.Generate(pattern, reggenRepeatLimit)
	if err != nil {
		return "", fmt.Errorf("generate %s: %w", rule.ID, err)
	}
	return value, nil
}

// GenerateSafe retries generation until the output passes IsSafeSecret and
// IsValidForRule, or maxSafeAttempts is exhausted. Rules with a custom
// generator use it directly; every other rule goes through reggen, which
// occasionally emits values containing regex metacharacters or values that
// don't actually match the source pattern.
func GenerateSafe(rule SecretRule) (string, error) {
	if gen, ok := customGenerators[rule.ID]; ok {
		for i := 0; i < maxSafeAttempts; i++ {
			value := gen()
			if IsSafeSecret(value) && IsValidForRule(value, rule) {
				return value, nil
			}
		}
		return "", fmt.Errorf("could not produce a safe secret for %s after %d attempts", rule.ID, maxSafeAttempts)
	}

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

// preprocessRegex turns a source regex into the pattern reggen should generate
// from. It strips the (?i) flag and ^/$ anchors, and — for "key_name + value"
// rules with multiple top-level capture groups — narrows generation to just the
// last capture group (the actual secret), so context like `client_secret` does
// not leak into the output.
func preprocessRegex(pattern string) string {
	pattern = strings.ReplaceAll(pattern, "(?i)", "")
	pattern = strings.TrimPrefix(pattern, "^")
	pattern = strings.TrimSuffix(pattern, "$")

	// When there are 2+ top-level capture groups, the leading ones are context
	// (key names, separators) and the last one is the secret value itself.
	if groups := topLevelCaptureGroups(pattern); len(groups) >= 2 {
		last := groups[0]
		for _, g := range groups[1:] {
			if g[0] > last[0] {
				last = g
			}
		}
		return pattern[last[0]:last[1]]
	}

	if wrapsWhole(pattern) {
		pattern = pattern[1 : len(pattern)-1]
	}
	return pattern
}

// topLevelCaptureGroups returns the [innerStart, innerEnd) spans of capturing
// groups that are not nested inside another group. Non-capturing groups (`(?:`,
// `(?i)`, …) and groups inside character classes are ignored.
func topLevelCaptureGroups(pattern string) [][2]int {
	type frame struct {
		capturing  bool
		innerStart int
	}
	var groups [][2]int
	var stack []frame
	inClass := false
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		if c == '\\' {
			i++ // skip the escaped character
			continue
		}
		if inClass {
			if c == ']' {
				inClass = false
			}
			continue
		}
		switch c {
		case '[':
			inClass = true
		case '(':
			capturing := i+1 >= len(pattern) || pattern[i+1] != '?'
			topLevel := len(stack) == 0
			stack = append(stack, frame{capturing: capturing && topLevel, innerStart: i + 1})
		case ')':
			if len(stack) > 0 {
				f := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if f.capturing {
					groups = append(groups, [2]int{f.innerStart, i})
				}
			}
		}
	}
	return groups
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

// randomFromCharset returns n characters chosen from charset using crypto/rand.
func randomFromCharset(n int, charset string) string {
	if n <= 0 || len(charset) == 0 {
		return ""
	}
	b := make([]byte, n)
	_, _ = rand.Read(b)
	out := make([]byte, n)
	for i := range b {
		out[i] = charset[int(b[i])%len(charset)]
	}
	return string(out)
}

func randomAlphanumeric(n int) string { return randomFromCharset(n, alphanumericChars) }

func randomHex(n int) string { return randomFromCharset(n, hexChars) }

func randomAlphanumericWithChars(n int, charset string) string {
	return randomFromCharset(n, charset)
}

func randomBool() bool {
	b := make([]byte, 1)
	_, _ = rand.Read(b)
	return b[0]&1 == 0
}
