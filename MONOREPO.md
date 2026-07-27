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

## Develop in sdk

```bash
cd sdk
git checkout -b fix/foo
# commit & push to wisp-trading/sdk
cd ..
git add sdk
git commit -m "chore: bump sdk submodule"
```
