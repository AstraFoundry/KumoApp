# Kumo Technical Documentation

Kumo is a calm desktop client for the Mihomo proxy core, with a command line
and agent guidance built in. It is one Go program built with
[MyGo](https://github.com/egoist/mygo): a native window that shows a React
frontend, a menu bar item, the `kumo` command and a guardian process.

## Domain Map

- [Product](product/README.md) — scope, daily workflow and information architecture.
- [Interfaces](interfaces/README.md) — the desktop app, the `kumo` CLI and agent control.
- [Core](core/README.md) — the DDD layers, the Mihomo lifecycle, profiles, overrides, runtime configuration and topology.
- [Operations](operations/README.md) — system proxy, TUN permission, persistence, logging and releases.
- [Quality](quality/README.md) — unit tests, lifecycle simulations, the end-to-end app test and lint.
- [Roadmap](roadmap/README.md) — what is deferred and why.
- [Decisions](decisions/) — earlier architecture decisions, still valid for the configuration model.
- [Agent index](agents.md) — where an agent finds the command or document for a task.

## Source Layout

```text
main.go                    one binary: app, kumo CLI (with arguments) or guardian
cmd/kumo/                  the CLI alone, for development
internal/
  domain/                  pure models and rules (DDD domain layer)
  application/             use cases shared by every surface, the event bus
  infrastructure/          adapters: storage, Mihomo, system proxy, ports, processes…
  interfaces/cli/          the kumo command
  interfaces/desktop/      MyGo bindings, window, tray, quit path, guardian spawn
locales/                   message catalogs (en, zh-Hans) shared by Go and the frontend
src/                       React frontend, Feature-Sliced Design
  app/ pages/ widgets/ features/ entities/ shared/
scripts/                   lint scripts and the end-to-end lifecycle test
```

## Architectural Principle

The desktop app, the CLI and the guardian share `internal/application`. A
surface never reimplements the Mihomo lifecycle, profile generation or
system proxy logic; it calls a service and renders what comes back.
