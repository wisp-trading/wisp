---
name: wisp-code-discipline
description: >
  Hard constraints for work on wisp / sdk / connectors. Prefer existing types and
  call paths; ban parallel abstractions and recreated logic. Use whenever editing
  this monorepo, designing credentials/settings, packaging, lifecycle, or when
  the user says "don't reinvent", "existing API", "stay within boundaries",
  or "/wisp-code-discipline".
---

# Wisp code discipline

## Non-negotiable

1. **Use what already exists.** Before adding a type, interface, or package, grep for
   the same job (`GetRequiredCredentialFields`, `NewConfig`, `Validate`, `MapToSDKConfig`,
   `StartStandalone`, `Wait`, `ResolveSettingsPath`, etc.).
2. **Banned: parallel abstractions.** Do not invent `CredentialSchemaProvider`, second
   config loaders, second packaging paths, or “better” copies of SDK surfaces the CLI
   already calls.
3. **Only add new code when absolutely needed** — bugfix, test, or thin glue. Prefer
   delete/demote over extend.
4. **Extend existing functions** if discovery/validation is weak; do not replace them
   with a new public API of the same purpose.
5. **Connectors own exchange config structs + `Validate()`.** CLI/SDK discover fields via
   `NewConfig()` + existing `GetRequiredCredentialFields` / map+validate — not a new schema DSL.
6. **No drive-by docs** unless the user asked. No new markdown files for design notes.
7. **Mocks via mockery**, not hand stubs.
8. **No DEPI (or other client/project) material in this codebase.** Delete on sight. Never copy,
   generate, or leave `DEPI*` / depi.co.id / Sanctum-unrelated sales specs under wisp/sdk/connectors.

## Credentials / Settings (current law)

| Concern | Existing path |
|---------|----------------|
| Global keys file | `~/.wisp/connectors.yml` via `config.ResolveSettingsPath` / Settings service |
| Field list for forms | `ConnectorService.GetRequiredCredentialFields(exchange)` |
| Config object | `registry.Connector.NewConfig()` then JSON map from credentials |
| Validate | `MapToSDKConfig` + `Config.Validate()` |
| Available venues | `connectors/types.AllConnectors` |

Do **not** add `CredentialField` / `CredentialSchemaProvider` unless the user
explicitly orders a redesign after rejecting the existing path.

## Packaging (current law)

- Standalone only: `main.go` + `StartStandalone` + `Wait`
- Plugins / `run-strategy` / `.so` are **removed** — do not reintroduce

## Trading signals (current law)

| Path | Purpose |
|------|---------|
| `wisp.Spot().Emit` / `Perp().Emit` / `Predict().Emit` / `Options().Emit` | **Places orders** (type-safe by market) |
| `wisp.Emit(signal)` | Untyped convenience — prefer domain Emit |
| `BaseStrategy.Publish` | Observability channel only (`Signals()`) — **not** trading |

## Config boot flow (current law)

```
strategyDir/config.yml     → StrategyConfig (exchanges/assets, no secrets)
ResolveSettingsPath        → ~/.wisp/connectors.yml (Configuration)
GetConnectorConfigsForStrategy → MapToSDKConfig + Validate → connector.Config
StartStandalone            → Initialize connectors + lifecycle
```

## Clarity backlog (parked — not blocking)

- `instruments` in strategy YAML unused for routing (by connector MarketType)
- Soft-skip unregistered connectors in runtime.initializeConnectors
- Zombie `pkg/signal` module; Predict `PredictionSignal` vs `Signal` naming
- Nested `spot/spot` package geography; `package packages`
- Executor/router naming pile-up; profiling measures dispatch
- Spot/perp/options copy-paste shells; dual analytics surfaces
- No release tags unless ship-worthy change

## When tempted to invent

1. Name the existing type/function that already does it.
2. If incomplete, improve **that** implementation + tests.
3. If truly missing, state the gap in one sentence and get confirmation before a new public type.

## Tests

Add tests against existing packages (compile, settings, manager, connector service).
Do not invent test-only APIs that production never uses.
