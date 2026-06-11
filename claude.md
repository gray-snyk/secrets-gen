# secrets-gen

## What this is
A CLI tool that generates fake secrets in the exact formats used by major services
(GitHub, AWS, Azure, Stripe, Slack, etc.) — useful for testing secret scanning tools
like Snyk without using real credentials.

## Tech stack
- Language: Go 1.21+
- CLI framework: github.com/spf13/cobra
- Terminal styling: github.com/charmbracelet/lipgloss
- Regex-based generation: github.com/lucasjones/reggen
- Clipboard support: github.com/atotto/clipboard
- Release pipeline: Goreleaser

## Project structure
secrets-gen/
├── main.go
├── assets/
│   └── secret_type_metadata.json   ← embed this, DO NOT modify it
├── cmd/
│   ├── root.go                     ← root cobra command + lipgloss banner
│   ├── generate.go                 ← generate subcommand
│   └── list.go                     ← list subcommand
├── internal/
│   ├── generators/
│   │   ├── loader.go               ← load + parse secret_type_metadata.json
│   │   ├── generate.go             ← generate secret value from regex via reggen
│   │   └── filter.go               ← entropy + safety checks
│   └── display/
│       └── styles.go               ← all lipgloss styles defined here
├── .goreleaser.yaml
└── README.md

## The metadata file
`assets/secret_type_metadata.json` is Snyk's real secret detection engine metadata.
667 rules. Each entry looks like:
```json
{
  "id": "github-personal-access-token",
  "title": "GitHub Personal Access Token",
  "provider": "GitHub",
  "prefixes": ["ghp_", "github"],
  "regex": "^(ghp_[a-zA-Z0-9]{36})$",
  "severity": "CRITICAL"
}
```

Fields used:
- `id` — machine identifier, used for lookups
- `title` — human-readable name shown in output
- `provider` — used for filtering (e.g. `generate github` matches all GitHub rules)
- `prefixes` — shown in output as context hints
- `regex` — used to generate the secret value via reggen
- `severity` — shown in output (color-coded)

## Secret generation approach
- Embed the JSON using go:embed — no external file dependency, works offline
- Use `reggen` to generate a string matching each rule's regex
- DO NOT hardcode any secret formats — everything is driven from the JSON
- Strip regex anchors (^ and $) and outer capture groups before passing to reggen

## The SecretRule struct
```go
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
```

## CLI commands

### generate
secrets-gen generate <provider>
secrets-gen generate github
secrets-gen generate aws --count 3
secrets-gen generate stripe --copy
secrets-gen generate github --id github-fine-grained-personal-access-token
secrets-gen generate --list-providers

Flags:
- `--count / -n` (default 1) — how many secrets to generate
- `--copy / -c` — copy first result to clipboard
- `--no-color` — plain output for piping
- `--format json` — output as JSON (id, title, provider, value, severity)
- `--id` — target a specific rule ID directly
- `--list-providers` — print all unique providers and exit

Provider matching: case-insensitive substring match on the `provider` field.
If multiple rules match (e.g. "github" matches 7 rules), generate one secret
per matching rule and display them grouped.

### list
secrets-gen list
secrets-gen list --provider github
secrets-gen list --severity CRITICAL

Shows a styled table: ID | Title | Severity | Prefix
Flags: `--provider` (filter), `--severity` (filter)

## Output design
Target look for `generate`:
╭──────────────────────────────────────────────────────╮
│  ⚡ secrets-gen                                        │
╰──────────────────────────────────────────────────────╯
GitHub Personal Access Token          CRITICAL
ghp_aB3xK9mNpQ2rS7tUvW4yZ1cD6eF8gH0j
GitHub Fine-Grained Personal Access Token   CRITICAL
github_pat_11AB3xK9m...
✓ Copied to clipboard

Severity color coding (lipgloss):
- CRITICAL → red `#FF4444`
- HIGH     → orange `#FF8C00`
- MEDIUM   → yellow `#FFD700`
- LOW      → blue `#4A9EFF`

## Lipgloss styles (define all in internal/display/styles.go)
```go
var (
    TitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED"))
    LabelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
    SecretStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true)
    CriticalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true)
    HighStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Bold(true)
    MediumStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
    LowStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#4A9EFF"))
    BoxStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
    CheckStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
)
```

## go:embed setup
In loader.go:
```go
//go:embed ../../assets/secret_type_metadata.json
var metadataBytes []byte
```
Adjust the relative path based on the actual package location.

## Regex preprocessing before passing to reggen
Some regexes in the metadata use patterns reggen struggles with.
Strip these before generating:
1. Leading `^` and trailing `$` anchors
2. Outer capture group: `^(pattern)$` → `pattern`
3. Case-insensitive flag `(?i)` — reggen doesn't support it; strip it and
   generate uppercase only if the pattern has [A-Z] in it
If reggen returns an error for a rule, skip it and note it in a debug log —
don't crash.

## Distribution
- Homebrew tap: gray-snyk/tap/secrets-gen
- Goreleaser handles: multi-platform binaries, Homebrew tap update, GitHub release
- Target platforms: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64

## Build order
1. go:embed + loader.go — parse the JSON, expose LoadRules() []SecretRule
2. generate.go — GenerateFromRule(rule SecretRule) (string, error) using reggen
3. cmd/root.go — cobra root with lipgloss banner
4. cmd/list.go — table output of all rules with filtering
5. cmd/generate.go — full generate command with all flags
6. display/styles.go — all styles in one place
7. .goreleaser.yaml
8. README.md with install instructions

## Key constraints
- Use crypto/rand where random choices are made manually (not needed if reggen handles it)
- When --no-color is set, skip all lipgloss styling
- When --format json is set, output raw JSON only, no decorative chrome
- The tool should work as a pipe: `secrets-gen generate aws --no-color | pbcopy`
- Never print a real secret — these are always synthetically generated from regex
