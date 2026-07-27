# Reference standalone strategy

**Blessed packaging path** for Wisp strategies.

## Contract

```text
fx.Start → runtime.StartStandalone(strategy, configDir, wispYml) → runtime.Wait()
```

- `Wait()` blocks until **SIGINT/SIGTERM** or monitoring **POST /shutdown**
- Then performs a single clean `Stop` and the process exits
- Do **not** write your own signal loop that ignores `/shutdown`

## Green path (from monorepo root)

```bash
# Go 1.26+
make smoke

# With credentials (copy example and enable an exchange):
cp wisp.yml.example wisp.yml
# edit wisp.yml → enable connector + keys

./examples/reference-standalone/reference-standalone \
  --config ./examples/reference-standalone \
  --wisp ./wisp.yml
```

## TUI / supervisor

Place a strategy under `strategies/<name>/` with `main.go` + `config.yml`.  
`wisp` Start Live will **compile a binary** and spawn it (not a `.so` plugin).

Legacy plugin path is only used if no binary can be built.

## Private strategies

Copy this package pattern. Keep private alpha out of product git.
