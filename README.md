# secrets-gen

A CLI that generates fake secrets in the exact formats used by major services
(GitHub, AWS, Azure, Stripe, Slack, and 200+ others) — useful for exercising
secret-scanning tools like Snyk without ever touching a real credential. Every
value is synthesized from the same regex rules a scanner would match against,
so the output is structurally indistinguishable from a real secret but
cryptographically meaningless.

## Install

### Homebrew (recommended)

```sh
brew install gray-snyk/tap/secrets-gen
```

### Direct download

Grab a prebuilt binary for your platform from the
[latest release](https://github.com/gray-snyk/secrets-gen/releases/latest):

- `secrets-gen_<version>_linux_amd64.tar.gz`
- `secrets-gen_<version>_linux_arm64.tar.gz`
- `secrets-gen_<version>_darwin_amd64.tar.gz`
- `secrets-gen_<version>_darwin_arm64.tar.gz`
- `secrets-gen_<version>_windows_amd64.zip`

Extract and drop `secrets-gen` somewhere on your `$PATH`.

### From source

```sh
go install github.com/gray-snyk/secrets-gen@latest
```

## Usage

### Generate

```sh
secrets-gen generate github               # one fake secret per matching GitHub rule
secrets-gen generate aws --count 3        # 3 fake secrets per matching AWS rule
secrets-gen generate stripe --copy        # copy the first result to the clipboard
secrets-gen generate --id github-personal-access-token
secrets-gen generate --list-providers     # print every provider known to the tool
secrets-gen generate aws --format json    # machine-readable output, no styling
secrets-gen generate aws --no-color | pbcopy   # plain text for piping
```

Provider matching is case-insensitive substring on the rule's provider field —
`secrets-gen generate github` matches every GitHub rule and emits one secret
per rule.

### List

```sh
secrets-gen list                                       # every rule
secrets-gen list --provider github                     # filter by provider
secrets-gen list --severity CRITICAL                   # filter by severity
secrets-gen list --provider aws --severity CRITICAL    # combine filters
```

### Flags

| Flag | Command | Default | Description |
| --- | --- | --- | --- |
| `--count, -n` | `generate` | `1` | Secrets to generate per matching rule |
| `--copy, -c` | `generate` | `false` | Copy the first generated secret to the clipboard |
| `--format` | `generate` | `` | Set to `json` for raw JSON output |
| `--id` | `generate` | `` | Generate a secret for a specific rule ID |
| `--list-providers` | `generate` | `false` | Print every unique provider and exit |
| `--provider` | `list` | `` | Filter rules by provider (case-insensitive substring) |
| `--severity` | `list` | `` | Filter rules by severity |
| `--no-color` | global | `false` | Disable ANSI colour output |

## A note on safety

Every secret this tool emits is **synthetic** — generated from a regex with a
pseudo-random match. They are not real credentials, will not authenticate
against any service, and cannot be used to access anything. They exist solely
to give secret-scanning tools something realistic-looking to detect.
