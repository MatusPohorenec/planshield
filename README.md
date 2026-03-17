# PlanShield

**Render a Terraform plan into a minimal, safe-to-share change list.**

PlanShield parses Terraform plan output (JSON or binary) and produces a clean, deterministic summary grouped by action — with sensitive values redacted so the output is safe for PR comments and CI logs.

## Features

- **Action grouping** — Resources sorted into Create / Update / Delete / Replace buckets, with destructive changes surfaced first
- **Sensitive-field redaction** — Values marked sensitive by Terraform are automatically replaced with `REDACTED`
- **Regex-based secret detection** — Built-in patterns catch common secret shapes (AWS keys, GitHub tokens, private keys, connection strings) even when providers fail to mark them sensitive
- **CI-friendly exit codes** — Exit 0 for clean plans, exit 2 when destroys/replaces are present, exit 1 on error
- **Multiple output formats** — `--format md` (default) for PR comments, `--format json` for programmatic consumption

## Installation

```bash
go install github.com/planshield/planshield/cmd/planshield@latest
```

Or build from source:

```bash
git clone https://github.com/planshield/planshield.git
cd planshield
go build -o planshield ./cmd/planshield
```

## Usage

```bash
# From a pre-generated JSON plan
terraform show -json tfplan > plan.json
planshield plan.json

# From a binary plan file (requires terraform on PATH)
planshield tfplan

# From stdin
terraform show -json tfplan | planshield -

# JSON output for CI pipelines
planshield --format json plan.json

# Custom redaction patterns
planshield --redact-patterns "my-org-secret-\d+,internal-token-[a-f0-9]+" plan.json

# Disable built-in secret detection (rely only on Terraform sensitive markers)
planshield --no-default-redact plan.json
```

## Example Output

### Markdown (default)

```
## PlanShield Summary

Terraform version: `1.7.4`

**1 Delete** · **1 Replace** · **1 Update** · **2 Create**

> **⚠ Destructive changes detected.** Review carefully before applying.

### - Delete (1)

| Resource Address |
|------------------|
| `aws_iam_role.old_role` |

### ! Replace (1)

| Resource Address |
|------------------|
| `aws_db_instance.main` |

### ~ Update (1)

| Resource Address |
|------------------|
| `aws_security_group.allow_tls` |

### + Create (2)

| Resource Address |
|------------------|
| `aws_instance.web` |
| `aws_s3_bucket.data` |
```

### JSON

```json
{
  "terraform_version": "1.7.4",
  "has_destroys": true,
  "total_changes": 5,
  "counts": {
    "Create": 2,
    "Update": 1,
    "Delete": 1,
    "Replace": 1
  },
  "groups": [
    {"action": "Delete", "count": 1, "addresses": ["aws_iam_role.old_role"]},
    {"action": "Replace", "count": 1, "addresses": ["aws_db_instance.main"]},
    {"action": "Update", "count": 1, "addresses": ["aws_security_group.allow_tls"]},
    {"action": "Create", "count": 2, "addresses": ["aws_instance.web", "aws_s3_bucket.data"]}
  ]
}
```

## CI Integration

### GitHub Actions

```yaml
- name: Render plan summary
  run: |
    terraform show -json tfplan | planshield - --format md > plan-summary.md

- name: Comment on PR
  uses: marocchino/sticky-pull-request-comment@v2
  with:
    path: plan-summary.md

- name: Check for destructive changes
  run: |
    terraform show -json tfplan | planshield - --format json
    # Exit code 2 = destructive changes present
    # Use this to gate applies or require extra approval
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `md` | Output format: `md` or `json` |
| `--redact-patterns` | _(none)_ | Comma-separated extra regex patterns for redaction |
| `--no-default-redact` | `false` | Disable built-in secret-detection patterns |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success — no destructive changes |
| 1 | Error (invalid input, parse failure, etc.) |
| 2 | Success — plan contains destroys or replaces |

## Built-in Redaction Patterns

PlanShield detects these common secret shapes by default:

- AWS access key IDs (`AKIA...`)
- Generic secret/password/token assignments
- Database connection strings with embedded credentials
- PEM private key headers
- GitHub personal access tokens (`ghp_...`)
- GitLab personal access tokens (`glpat-...`)
- Bearer tokens

Disable with `--no-default-redact` and supply your own via `--redact-patterns`.

## Project Structure

```
planshield/
├── cmd/planshield/       # CLI entry point
│   └── main.go
├── internal/
│   ├── plan/             # Data model, parser, grouping logic
│   │   ├── model.go
│   │   ├── parse.go
│   │   ├── group.go
│   │   └── plan_test.go
│   ├── redact/           # Sensitive value redaction
│   │   ├── patterns.go
│   │   ├── redact.go
│   │   └── redact_test.go
│   ├── render/           # Output formatters
│   │   ├── markdown.go
│   │   ├── json.go
│   │   └── render_test.go
│   └── runner/           # Terraform CLI invocation
│       └── terraform.go
├── testdata/             # Test fixtures
│   ├── mixed_plan.json
│   └── no_changes.json
├── go.mod
└── README.md
```

## License

MIT
