# Reference standalone strategy

**Blessed packaging path** for Wisp strategies.

## Contract

```text
fx.Start → runtime.StartStandalone(strategy, configDir, wispYml) → runtime.Wait()
```

- `Wait()` blocks until **SIGINT/SIGTERM** or monitoring **POST /shutdown**
- Then performs a single clean `Stop` and the process exits
- Do **not** write your own signal loop that ignores `/shutdown`

## Run (when connectors build is green)

```bash
# from monorepo root (wisp + sdk submodule)
cd examples/reference-standalone
# copy/adapt config as needed
go run . -config . -wisp ../../wisp.yml
```

Until `connectors` hyperliquid is fixed ([connectors#52](https://github.com/wisp-trading/connectors/issues/52)),
full `go run` may fail to compile the live Module. The source below is still the
canonical pattern for strategy binaries (including private alpha).

## Files

| File | Role |
|------|------|
| `main.go` | Process host: fx + StartStandalone + Wait |
| `strategy.go` | Minimal self-directed strategy |

Private strategies should copy this pattern; they stay out of this repo.
