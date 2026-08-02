# Wisp monorepo

This repository (`github.com/wisp-trading/wisp`) is the **star-facing entry point**.

## In this repo

| Path | Module | Notes |
|------|--------|--------|
| `.` | `github.com/wisp-trading/wisp` | CLI + TUI |
| `sdk/` | `github.com/wisp-trading/sdk` | git submodule — runtime / lifecycle / domains |

## Not in this repo (by design)

| Module | Why separate |
|--------|----------------|
| `github.com/wisp-trading/connectors` | Heavy, venue-specific; clone only when changing exchanges |
| `github.com/wisp-trading/backtest` | Experimental sim |
| `polymarket-go-sdk` | Third-party fork; pulled as a normal Go module dependency |
| Private strategies | Local only — never product git |

## Clone

```bash
git clone --recurse-submodules git@github.com:wisp-trading/wisp.git
cd wisp
# Go 1.26+ (Green Tea GC default)
go test ./internal/services/live/manager/ ./sdk/pkg/lifecycle/ ./sdk/pkg/runtime/
go build -o bin/wisp .
```

Without submodules (CLI module only):

```bash
git clone git@github.com:wisp-trading/wisp.git
cd wisp
go build -o bin/wisp .
```

## Working on connectors (optional)

Keep connectors as a sibling checkout when needed:

```bash
# next to monorepo, or anywhere:
git clone git@github.com:wisp-trading/connectors.git
cd wisp
go work use ../connectors   # temporary; do not commit unless intentional
# or in connectors/go.mod while developing:
# replace github.com/wisp-trading/sdk => ../wisp/sdk
```

Default product builds use **released** `connectors` module versions from `go.mod` — no submodule required.

### Composition (avoid loading every venue)

Strategies should import **only the venue modules they use** (e.g. `hyperliquid.Module`),
not `connectors.Module` (the bag of every exchange). Settings/CLI may keep the bag for discovery.

Live-exchange integration tests live next to each venue package (`//go:build integration`)
and share harnesses from `connectors/pkg/testing/connector`. Run one venue:

```bash
cd ../connectors
make test-integration PKG=./pkg/connectors/hyperliquid
```

## Develop in sdk

```bash
cd sdk
git checkout -b fix/foo
# commit & push to wisp-trading/sdk
cd ..
git add sdk
git commit -m "chore: bump sdk submodule"
```

## Reference strategy

See [`examples/reference-standalone`](./examples/reference-standalone) for the blessed
`StartStandalone` + `Wait` process host. Copy that pattern for private strategies.

## Green path

```bash
make help    # developer targets
make smoke   # builds CLI + examples/reference-standalone binary
make ci      # local stand-in for GitHub Actions (verify, vet, lint, test, smoke)
```

TUI Start Live compiles `strategies/<name>/` as a **standalone binary** when `main.go` exists.

## CI

GitHub Actions (`.github/workflows/`):

| Workflow | Purpose |
|----------|---------|
| `ci.yml` | modules verify + tidy, golangci-lint, race tests, packaging smoke |
| `security.yml` | `govulncheck` on PR/main + weekly |
| `release.yml` | multi-platform binaries on `v*` tags |

Always runs with `GOWORK=off` and the public module proxy so checksums match the sumdb.

