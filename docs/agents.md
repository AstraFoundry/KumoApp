# Agent Path Index

Quick reference for AI agents working on this codebase: where each kind of
task lives and which command checks it.

| Task | Start here | Check with |
| --- | --- | --- |
| Lifecycle, ports, quit, guardian | `internal/application/lifecycle.go`, `guardian.go`; [runtime controller](core/mihomo-runtime-controller.md) | `KUMO_TEST_MIHOMO=… go test ./internal/application`, `make e2e` |
| System proxy ownership | `internal/application/systemproxy.go`, `internal/domain/systemproxy` | same |
| Generated config, overrides | `internal/domain/runtimeconfig`, `internal/domain/override` | `go test ./internal/domain/...` |
| Topology graph | `internal/domain/topology`, `src/widgets/topology-canvas` | `go test ./internal/domain/topology` |
| A CLI command | `internal/interfaces/cli/commands.go`; [CLI](interfaces/cli-agent-control.md) | `go run ./cmd/kumo <cmd> --json` |
| A desktop binding or event | `internal/interfaces/desktop/{services,app,bridge}.go` | `make generate` (regenerates `src/shared/api/mygo.ts`) |
| A page or UI flow | `src/pages/<page>`, then its widgets/features; [desktop interface](interfaces/desktop-interface.md) | `make lint` |
| Copy | `locales/en.json` and `locales/zh-Hans.json`, then `make generate` | `bun run lint:i18n` |
| Agent prompts per page | `internal/application/guidance.go` | `kumo prompt <page>` |
| The bundled agent skill | `internal/infrastructure/skills/bundle` (bump `bundleVersion`) | `kumo skills status --json` |
| Releases | [release management](operations/release-management.md) | `make check` |

## Common Commands

```bash
make dev        # run the app with live reload
make generate   # TypeScript client + message catalogs
make test       # go vet + go test
make lint       # tsc, FSD layering, message catalogs
make build      # package for this platform into build/
make e2e MIHOMO=/path/to/mihomo
```

## Decision Records

- [ADR-001](decisions/ADR-001-dns-sniffer-decoupling.md) — DNS/Sniffer decoupling
- [ADR-002](decisions/ADR-002-policy-value-types.md) — Policy value types
- [ADR-003](decisions/ADR-003-hosts-top-level-vs-nested.md) — Hosts top-level vs nested
- [ADR-004](decisions/ADR-004-restart-vs-patch-for-dns-sniffer.md) — Restart vs patch for DNS/Sniffer

The ADRs predate the Go rewrite; their file paths refer to the Swift
sources, their decisions still hold for `internal/domain/runtimeconfig`.
