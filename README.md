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

Run `secrets-gen` with no arguments to launch an interactive provider picker,
or pass a provider name to generate secrets for it directly.

```sh
secrets-gen                          # interactive provider picker
secrets-gen github                   # one fake secret per matching GitHub rule
secrets-gen aws --count 3            # 3 fake secrets per matching AWS rule
secrets-gen stripe --copy            # copy the first result to the clipboard
secrets-gen --id github-personal-access-token
secrets-gen --list-providers         # print every provider known to the tool
secrets-gen aws --format json        # machine-readable output, no styling
secrets-gen aws --no-color | pbcopy  # plain text for piping
```

In the interactive picker, use ↑/↓ to navigate, `/` to filter, `Enter` to
select a provider and generate its secrets, and `q` or `Ctrl+C` to quit.

Provider matching is case-insensitive substring on the rule's provider field —
`secrets-gen github` matches every GitHub rule and emits one secret per rule.

### Supported providers

The tool surfaces a curated set of 14 providers:

| | | | |
| --- | --- | --- | --- |
| AWS | Anthropic | Azure | Bitbucket |
| ClickUp | Cloudflare | Datadog | Docker |
| GitHub | GitLab | OpenAI | Square |
| Stripe | npm | | |

Any other provider in the underlying metadata is excluded — it won't appear in
the picker or `--list-providers`, and requesting it directly returns no rules.

### Flags

| Flag | Default | Description |
| --- | --- | --- |
| `--count, -n` | `1` | Secrets to generate per matching rule |
| `--copy, -c` | `false` | Copy the first generated secret to the clipboard |
| `--format` | `` | Set to `json` for raw JSON output |
| `--id` | `` | Generate a secret for a specific rule ID |
| `--list-providers` | `false` | Print every unique provider and exit |
| `--no-color` | `false` | Disable ANSI colour output |

## A note on safety

Every secret this tool emits is **synthetic** — generated from a regex with a
pseudo-random match. They are not real credentials, will not authenticate
against any service, and cannot be used to access anything. They exist solely
to give secret-scanning tools something realistic-looking to detect.
