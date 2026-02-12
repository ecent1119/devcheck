# devcheck

**Free and open source** — Local project readiness inspector. Find out what's missing before "it doesn't work on my machine."

devcheck scans your repository and tells you what's needed to run the project locally: missing env files, undefined variables, unmet dependencies, and unclear setup steps.

## The Problem

You clone a repo. You run `docker compose up`. It fails.

```
ERROR: environment variable DATABASE_URL is not set
```

Now you're grepping through files, checking READMEs, and asking teammates. devcheck does this for you:

```
devcheck scan

Blocking:
  ✗ docker-compose.yml references ${DATABASE_URL} but it is not defined

Warnings:
  ⚠ .env.example exists but .env is missing
  ⚠ .env.example has API_KEY but .env does not

Info:
  ℹ Detected Node.js project (package.json)
  ℹ Likely entrypoint: pnpm install && pnpm dev (from README.md)
```

## Features

- **Env var analysis** — finds `${VAR}` in compose files and checks if they're defined
- **Missing file detection** — flags missing `.env` when `.env.example` exists
- **Compose validation** — checks depends_on references, undefined services
- **Language detection** — identifies Node, Go, Python, Rust, Java projects
- **Run hints** — scans README for setup instructions
- **Project config file** — `.devcheck.yaml` for custom rules, required vars, ignored checks
- **Tool version checks** — verify docker, docker-compose, node, go, python versions
- **Build context validation** — ensures Dockerfiles exist in build.context paths
- **Fix list generation** — generate actionable markdown checklists
- **Source code scanning** — detect env vars used in code but not defined
- **Multiple outputs** — text, JSON, or Markdown checklist
- **Check profiles** — default, strict, ci, minimal, full

## Quick Start

```bash
# Scan current directory
devcheck scan

# Scan specific path
devcheck scan /path/to/project

# JSON output for CI
devcheck scan --format json

# Fail CI if blocking issues found
devcheck scan --strict

# Use a check profile
devcheck scan --profile ci

# Check tool versions
devcheck scan --check-tools

# Generate fix checklist
devcheck scan --fix-list fixes.md

# Use custom config file
devcheck scan --config .devcheck.yaml
```

## Configuration File

Create `.devcheck.yaml` for project-specific rules:

```yaml
# Custom validation rules
custom_rules:
  - id: "DB_REQUIRED"
    pattern: "^DATABASE_"
    required: true
    description: "Database variables must be defined"
    severity: blocking

# Minimum tool versions
tool_versions:
  docker: "20.10.0"
  docker_compose: "2.0.0"
  node: "18.0.0"

# Finding codes to ignore
ignore_codes:
  - "HINT001"

# Environment variables that must always be defined
required_env_vars:
  - "NODE_ENV"
  - "DATABASE_URL"

# Map service names to expected Dockerfile paths
build_contexts:
  api: "./api"
  web: "./frontend"
```

## Example Output

```
devcheck v1.0.0

Scanning: /Users/dev/myproject

Artifacts Found:
  ✓ docker-compose.yml
  ✓ .env.example
  ✗ .env (missing)
  ✓ package.json (Node.js)
  ✓ pnpm-lock.yaml

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Blocking (2):
  ✗ ENV001  ${DATABASE_URL} referenced in docker-compose.yml but not defined
     Fix: Add DATABASE_URL to .env file

  ✗ ENV002  ${REDIS_URL} referenced in docker-compose.yml but not defined
     Fix: Add REDIS_URL to .env file

Warnings (1):
  ⚠ ENV003  .env.example exists but .env is missing
     Fix: Copy .env.example to .env and fill in values

Info (2):
  ℹ LANG001  Detected Node.js project with pnpm
  ℹ HINT001  Likely entrypoint: pnpm install && pnpm dev
```

## What It Is / What It Isn't

**It is:**
- A readiness checker for local development
- A pre-onboarding validator
- A way to document what's needed to run a project

**It is not:**
- A build tool
- A dependency installer
- A security scanner
- A production validator

## Installation

Download the binary for your platform from [Gumroad](https://example.gumroad.com/l/devcheck).

```bash
# macOS / Linux
chmod +x devcheck
sudo mv devcheck /usr/local/bin/

# Verify
devcheck version
```

## Flags Reference

| Flag | Description |
|------|-------------|
| `--format` | Output format: `text`, `json`, `markdown`, `checklist` |
| `--compose` | Specify compose file path |
| `--env` | Specify env file(s) |
| `--strict` | Exit 1 if blocking findings exist |
| `--profile` | Check profile: `default`, `strict`, `ci`, `minimal`, `full` |
| `--check-tools` | Check tool versions (docker, docker-compose, etc.) |
| `--config` | Custom config file path |
| `--fix-list` | Generate fix checklist to file (markdown) |
| `--no-color` | Disable color output |

## Exit Codes

- `0` — Scan completed successfully
- `1` — Scan completed, `--strict` mode and blocking findings found
- `2` — Parse error or invalid input

## Finding Codes

| Code | Description |
|------|-------------|
| ENV001 | Variable referenced but not defined |
| ENV002 | Variable in .env.example missing from .env |
| ENV003 | .env missing when .env.example exists |
| CMP001 | depends_on references unknown service |
| LANG001 | Language/framework detected |
| HINT001 | Run instructions found |

## Related Tools

devcheck is part of a local development toolchain:

- **[stackgen](https://github.com/ecent1119/stackgen)** — Generate docker-compose.yml for any stack
- **[envgraph](https://github.com/ecent1119/envgraph)** — Visualize environment variable flow
- **[dataclean](https://github.com/ecent1119/dataclean)** — Snapshot and reset Docker volumes
- **[compose-diff](https://github.com/ecent1119/compose-diff)** — Semantic diff for compose files

## Support This Project

**devcheck is free and open source.**

If this tool saved you time, consider sponsoring:

[![Sponsor on GitHub](https://img.shields.io/badge/Sponsor-❤️-red?logo=github)](https://github.com/sponsors/ecent1119)

Your support helps maintain and improve this tool.

## License

MIT License — see [LICENSE](LICENSE) for details.
