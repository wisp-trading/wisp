# Wisp monorepo

This repository (`github.com/wisp-trading/wisp`) is the **star-facing entry point** for the stack.

## Layout

| Path | Module | Notes |
|------|--------|--------|
| `.` | `github.com/wisp-trading/wisp` | CLI + TUI |
| `sdk/` | `github.com/wisp-trading/sdk` | submodule |
| `connectors/` | `github.com/wisp-trading/connectors` | submodule (heavy) |
| `backtest/` | `github.com/wisp-trading/backtest` | submodule, experimental |
| `polymarket-go-sdk/` | third-party fork | submodule |

Private strategies are **not** in this repo.

## Clone

```bash
git clone --recurse-submodules git@github.com:wisp-trading/wisp.git
cd wisp
# Go 1.26+ (Green Tea GC is default)
go test ./internal/services/live/manager/ ./sdk/pkg/lifecycle/ ./sdk/pkg/runtime/
go build -o bin/wisp .
```

## Develop in a submodule

```bash
cd sdk
git checkout -b fix/foo
# commit & push to wisp-trading/sdk
cd ..
git add sdk
git commit -m "chore: bump sdk"
```

## Go version

- `go 1.26` + `toolchain go1.26.5`
- Green Tea GC on by default; opt out: `GOEXPERIMENT=nogreenteagc`
