# Reference standalone strategy

**Only packaging path** for Wisp strategies.

## Contract

```text
fx.Start → runtime.StartStandalone(strategy, configDir, settingsPath) → runtime.Wait()
```

- `settingsPath` empty → `~/.wisp/connectors.yml` (or `WISP_SETTINGS` / project-local migration)
- `Wait()` blocks until **SIGINT/SIGTERM** or monitoring **POST /shutdown**
- Then performs a single clean `Stop` and the process exits

## Green path (from monorepo root)

```bash
make smoke

# Credentials via TUI Settings → ~/.wisp/connectors.yml
./examples/reference-standalone/reference-standalone \
  --config ./examples/reference-standalone
```

## TUI / supervisor

Place a strategy under `strategies/<name>/` with `main.go` + `config.yml`.  
`wisp` Start Live compiles the binary and spawns it. No plugin / `.so` path.

## Private strategies

Copy this package pattern. Keep private alpha out of product git.
