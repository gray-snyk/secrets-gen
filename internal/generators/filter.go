package generators

import (
	"math"
	"regexp"
	"strings"
)

// forbiddenChars rejects reggen output that contains regex metacharacters or
// punctuation that wouldn't appear in real-world secrets. `^` is included
// alongside `$` because some source rules use unintended ranges like
// `[0-9A-za-z]` that legitimately match it.
const forbiddenChars = "][)($^!%&*#@{}|\\`~<>?,;'\" \n\t\r"

const minEntropy = 2.8

// IsSafeSecret returns true when the generated value looks like a plausible
// secret: no regex metacharacters, no whitespace, and enough Shannon entropy
// to rule out degenerate / highly repetitive output from reggen.
func IsSafeSecret(value string) bool {
	if value == "" {
		return false
	}
	if strings.ContainsAny(value, forbiddenChars) {
		return false
	}
	if shannonEntropy(value) <= minEntropy {
		return false
	}
	return true
}

// IsValidForRule re-validates the generated value against the rule's original
// regex. If the regex can't be compiled by Go's RE2 engine, we don't block on
// our own parser limitation and return true.
func IsValidForRule(value string, rule SecretRule) bool {
	re, err := regexp.Compile(rule.Regex)
	if err != nil {
		return true
	}
	return re.MatchString(value)
}

func shannonEntropy(s string) float64 {
	counts := make(map[rune]int)
	total := 0
	for _, r := range s {
		counts[r]++
		total++
	}
	if total == 0 {
		return 0
	}
	length := float64(total)
	var entropy float64
	for _, c := range counts {
		p := float64(c) / length
		entropy -= p * math.Log2(p)
	}
	return entropy
}
