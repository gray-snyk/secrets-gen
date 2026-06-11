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
const forbiddenChars = "][)($^!%&*#@{}|\\`~<>?,;'\" \n\t\r+=/"

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

// IsValidForRule re-validates the generated value against the pattern it was
// actually generated from — the preprocessed generation target, anchored — not
// the raw source regex. This matters for "key_name + value" rules (e.g. Azure)
// where we deliberately emit only the secret portion, which would never match
// the full source regex that also requires the key-name prefix.
//
// Rules backed by a custom generator are authoritative for their realistic
// format (which may intentionally diverge from the baroque source regex), so
// they are considered valid by construction. If the target can't be compiled
// by Go's RE2 engine, we don't block on our own parser limitation.
func IsValidForRule(value string, rule SecretRule) bool {
	if _, ok := customGenerators[rule.ID]; ok {
		return true
	}
	target := preprocessRegex(rule.Regex)
	re, err := regexp.Compile("^(?:" + target + ")$")
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
