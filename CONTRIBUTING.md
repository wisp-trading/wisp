# Contributing to Wisp

Thanks for investing time in a premium open-source trading stack. This guide keeps local work and CI identical.

## Prerequisites

- **Go 1.26+** (toolchain in `go.mod` pins `go1.26.5`; Green Tea GC is default)
- `git`, `make`
- Optional: [golangci-lint](https://golangci-lint.run/) v2.12+

```bash
git clone --recurse-submodules https://github.com/wisp-trading/wisp.git
cd wisp
```

## Module hygiene

Public modules should resolve via the Go module proxy so checksums match CI:

```bash
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org
export GOPRIVATE=github.com/wisp-trading/*   # private org modules only
export GOWORK=off                            # CI always uses GOWORK=off
```

`GOPROXY=direct` against retagged third-party modules can write a `go.sum` that fails on Actions. Prefer the proxy for public deps.

## Day-to-day

| Command        | Purpose                                      |
|----------------|----------------------------------------------|
| `make help`    | List targets                                 |
| `make build`   | CLI → `bin/wisp`                             |
| `make test`    | Unit tests (`-race`, shuffled)               |
| `make lint`    | golangci-lint (see `.golangci.yml`)          |
| `make smoke`   | CLI + `examples/reference-standalone` build  |
| `make ci`      | Local stand-in for GitHub CI                 |

Mocks are generated with [mockery](https://github.com/vektra/mockery) (`.mockery.yaml`). Commit regenerated mocks with the interface change — do not hand-write stubs.

## Pull requests

1. Branch from `main`.
2. Keep PRs focused (one concern per PR when practical).
3. Fill in the PR template (summary, test plan).
4. CI must be green: **Modules**, **Lint**, **Test**, **Smoke** (and **Security** on schedule/PRs).
5. Prefer squash merge with a clear commit subject.

### Scope notes

| In this repo              | Separate                                          |
|---------------------------|---------------------------------------------------|
| CLI + TUI (`github.com/wisp-trading/wisp`) | `connectors` (venue adapters)          |
| `sdk/` submodule          | Private strategies / alpha — **never** product git |

See [MONOREPO.md](./MONOREPO.md) for layout details. Blessed strategy packaging: [`examples/reference-standalone`](./examples/reference-standalone) (`StartStandalone` + `Wait`).

## Code style

- Idiomatic Go; `gofmt` / `goimports` required
- Errors wrapped with `%w` where the chain matters
- No secrets in git (credentials in `~/.wisp/connectors.yml`; `.env` ignored)
- Prefer interfaces + mockery over hand-rolled test doubles

## Security

Report vulnerabilities privately — see [SECURITY.md](./SECURITY.md). Do not open public issues for undisclosed vulns.

## License

By contributing, you agree that your contributions are licensed under the MIT License (see [LICENSE](./LICENSE)).
