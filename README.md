# devcheck

Local project readiness inspector — know if your dev environment will actually work.

---

## The problem

- Clone repo → run commands → something fails → debug for 30 minutes
- "Works on my machine" is always the answer
- Missing dependencies discovered at runtime
- Docker version mismatches cause silent failures
- README instructions are outdated

---

## What it does

- Scans project for requirements (Docker, Node, Go, Python, etc.)
- Checks if dependencies are installed and correct versions
- Validates Docker Compose can start
- Verifies ports aren't already in use
- Reports what's missing before you waste time

---

## Example output

```bash
$ devcheck

Project: my-app
─────────────────

Runtime Requirements:
  ✅ Docker 24.0.0 (required: >=20.0.0)
  ✅ docker-compose 2.24.0
  ✅ Node.js 20.11.0 (required: >=18.0.0)
  ❌ Go not installed (required: >=1.21.0)

Port Availability:
  ✅ 3000 available
  ✅ 5432 available
  ❌ 6379 in use by redis-server (PID 1234)

Docker Compose:
  ✅ docker-compose.yml valid
  ⚠️  .env.example exists but .env missing
  ✅ All referenced images pullable

Files:
  ✅ package.json present
  ❌ go.mod missing (expected for Go project)

Summary: 3 issues found
  • Install Go 1.21+
  • Stop redis-server on port 6379 or change compose port
  • Copy .env.example to .env
```

---

## What it checks

| Category | Checks |
|----------|--------|
| **Runtimes** | Docker, Node, Go, Python, Java, Rust, .NET |
| **Ports** | Compose-defined ports availability |
| **Files** | Required config files, .env setup |
| **Docker** | Compose validity, image availability |
| **Versions** | Semver compatibility with project requirements |
| **Build contexts** | Dockerfile exists in build.context paths |
| **Source code** | Env vars used in code but not defined |

---

## New in v2.0

- **Project config file** — `.devcheck.yaml` for custom rules, required vars, ignored checks
- **Tool version checks** — verify docker, docker-compose, node, go, python versions
- **Build context validation** — ensures Dockerfiles exist in build.context paths  
- **Fix list generation** — generate actionable markdown checklists
- **Check profiles** — default, strict, ci, minimal, full

---

## Output formats

| Format | Use case |
|--------|----------|
| `--format text` | Human-readable terminal output |
| `--format json` | CI integration, scripts |
| `--format markdown` | Documentation |

---

## Scope

- Read-only inspection
- No modifications to system
- No installations
- No telemetry

---

## Get it

**$29** — one-time purchase, standalone macOS/Linux/Windows binary.

👉 [Download on Gumroad](https://ecent.gumroad.com/l/rafogb)

---

## Related tools

| Tool | Purpose |
|------|---------|
| **[stackgen](https://github.com/stackgen-cli/stackgen)** | Generate local dev Docker Compose stacks |
| **[envgraph](https://github.com/stackgen-cli/envgraph)** | Scan and validate environment variable usage |
| **[dataclean](https://github.com/stackgen-cli/dataclean)** | Reset local dev data safely |
| **[compose-diff](https://github.com/stackgen-cli/compose-diff)** | Semantic Docker Compose diff |

---

## License

MIT — this repository contains documentation and examples only.
