// Command smoketest exercises every rule's generator and reports how many of
// 5 attempts produce a value that passes both IsSafeSecret and IsValidForRule.
//
//	go run cmd/smoketest/main.go
//
// Exits 1 if any rule fails completely (0/5), 0 otherwise.
package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/gray-snyk/secrets-gen/internal/generators"
)

const attempts = 5

func main() {
	rules, err := generators.LoadRules()
	if err != nil {
		fmt.Fprintln(os.Stderr, "load rules:", err)
		os.Exit(1)
	}

	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })

	var failures int
	for _, rule := range rules {
		passed := 0
		for i := 0; i < attempts; i++ {
			value, err := generators.GenerateSafe(rule)
			if err != nil {
				continue
			}
			if generators.IsSafeSecret(value) && generators.IsValidForRule(value, rule) {
				passed++
			}
		}

		status := "PASS"
		note := ""
		switch {
		case passed == 0:
			status = "FAIL"
			failures++
		case passed < attempts:
			status = "WARN"
			note = fmt.Sprintf("  (%d skipped)", attempts-passed)
		}

		fmt.Printf("%-4s  %-45s %d/%d%s\n", status, rule.ID, passed, attempts, note)
	}

	fmt.Printf("\n%d rules checked, %d failed\n", len(rules), failures)
	if failures > 0 {
		os.Exit(1)
	}
}
