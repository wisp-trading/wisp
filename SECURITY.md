# Security Policy

## Supported versions

| Version | Supported          |
|---------|--------------------|
| `main`  | :white_check_mark: |
| tagged releases (`v*`) | :white_check_mark: latest minor |
| untagged forks / old tags | :x: |

## Reporting a vulnerability

Please **do not** open a public GitHub issue for security problems.

1. Email **security@usewisp.dev** (or the maintainers listed on the GitHub org) with:
   - Description and impact
   - Reproduction steps / PoC (if safe to share)
   - Affected commit or release tag
2. Allow a reasonable window for triage before public disclosure.
3. We will acknowledge receipt, assess severity, and coordinate a fix and credit.

## Scope

In scope:

- The `wisp` CLI/TUI and packaging path in this repository
- Published `github.com/wisp-trading/sdk` APIs used by strategies
- CI supply-chain issues that affect release artifacts

Out of scope (report to the owning repo/vendor):

- Third-party exchanges and venue APIs
- User strategy code and private keys / credential handling mistakes
- Issues only reproducible with intentionally malicious local config

## Hardening notes for operators

- Never commit API keys; credentials live in `~/.wisp/connectors.yml` (mode 0600)
- Run live strategies under least-privilege OS users and process supervisors
- Strategies are standalone binaries only (`StartStandalone` + `Wait`)
- Keep Go toolchain and dependencies current (`make verify`, Dependabot PRs)

Automated checks:

- `govulncheck` on PRs / main / weekly schedule (`.github/workflows/security.yml`)
- Module checksum verification in CI (`go mod verify`)
