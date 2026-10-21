<p align="center">
  <img src="Assets/KumoApp-Banner-1280x640.png" alt="Kumo app banner" width="640" style="border-radius: 24px;">
</p>

<h1 align="center">Kumo</h1>

<p align="center">
  <strong>A calm desktop client for the <a href="https://github.com/MetaCubeX/mihomo">Mihomo</a> proxy core.</strong>
</p>

<p align="center">
  Go + MyGo app · React UI · Agent-friendly CLI
</p>

<p align="center">
  <a href="https://go.dev">
    <img alt="Go" src="https://img.shields.io/badge/Go-1.27-00ADD8?logo=go&logoColor=white">
  </a>
  <a href="https://www.apple.com/macos/">
    <img alt="macOS 13+" src="https://img.shields.io/badge/macOS-13%2B-000000?logo=apple&logoColor=white">
  </a>
  <a href="https://github.com/MetaCubeX/mihomo">
    <img alt="Powered by Mihomo" src="https://img.shields.io/badge/core-Mihomo-0F766E">
  </a>
  <img alt="Status: active development" src="https://img.shields.io/badge/status-active%20development-2563EB">
</p>

<p align="center">
  <a href="#features">Features</a> ·
  <a href="#quick-start">Quick Start</a> ·
  <a href="#usage">Usage</a> ·
  <a href="#architecture">Architecture</a> ·
  <a href="#documentation">Docs</a> ·
  <a href="#roadmap">Roadmap</a>
</p>

---

## Overview

Kumo connects you through Mihomo without turning proxy management into a
network operations dashboard. The first screen answers four questions: is
Kumo connected, which mode is active, is the system proxy on, and which
profile and nodes are in use. DNS, TUN, rules, logs, overrides and the
routing topology stay one click away.

## Features

| Area | What Kumo provides |
| --- | --- |
| Desktop app | One Go binary built with [MyGo](https://github.com/egoist/mygo): a native window with a React + [coss ui](https://coss.com/ui) frontend, a menu bar item and notifications. |
| Topology | A React Flow canvas of entry → rules → groups → nodes, with selected paths, dynamic members, broken references and a text inspector. |
| Leaves no trace | Quitting, or a crash caught by the guardian process, stops the core and clears the system proxy — only if Kumo still owns it. |
| Plays well with others | Another app taking over the system proxy is shown, not fought; a port held by another program is named, with a free port offered. |
| CLI for humans and agents | The same executable as `kumo`: stable commands, `--json`, dry-run, predictable exit codes. |
| Agent guidance | Each page copies a ready prompt for a coding agent; the Kumo skill installs for Claude Code, Codex, Cursor, Gemini CLI and `.agents`. |
| First run | A guided setup with an animated preview of each step. |

## Quick Start

Download the disk image from [Releases](https://github.com/AstraFoundry/KumoApp/releases),
open Kumo and follow Setup. It finds a Mihomo you have, or downloads one.

### From source

Requires Go 1.27+, Bun and, on Linux, WebKitGTK.

```bash
bun install
make dev          # live reload
make build        # package into build/
make check        # generate, go vet/test, typecheck, FSD and message checks
```

## Usage

| Group | Pages |
| --- | --- |
| Daily | Overview, Profiles, Proxies, Topology |
| Inspect | Connections, Logs, Rules |
| Configure | Core, System Proxy, DNS, TUN, Sniffer, Resources, Overrides |
| Kumo | Agents, Settings, About |

```bash
kumo status --json
kumo start
kumo proxies --json
kumo select Proxy "Tokyo 01"
kumo sysproxy on
kumo topology --issues
kumo prompt proxies           # a ready-made prompt for a coding agent
kumo -l                       # every command
```

## Architecture

```text
internal/domain          rules: runtime, system proxy ownership, profiles, overrides, config, topology
internal/application     use cases shared by the app, the CLI and the guardian
internal/infrastructure  storage, Mihomo client/launcher, system proxy, ports, processes
internal/interfaces      cli (kumo) and desktop (MyGo bindings, window, tray, quit path)
src/                     React frontend in Feature-Sliced Design: app, pages, widgets, features, entities, shared
```

The backend follows Domain-Driven Design; the frontend follows Feature-Sliced
Design. See [docs/README.md](docs/README.md).

## Documentation

- [Product](docs/product/README.md) · [Interfaces](docs/interfaces/README.md) ·
  [Core](docs/core/README.md) · [Operations](docs/operations/README.md) ·
  [Quality](docs/quality/README.md) · [Roadmap](docs/roadmap/README.md)
- [Agent index](docs/agents.md) for coding agents working on Kumo.

## Contributing

1. Read [AGENTS.md](AGENTS.md) and the document under [docs/](docs/) for your change.
2. Keep business rules in `internal/domain` and `internal/application`; surfaces only call services.
3. Run `make check`; for lifecycle changes also `make e2e MIHOMO=<path>`.
4. Update `docs/` in the same change when behavior, architecture, persistence,
   permissions or tests change.

## License

Kumo is licensed under the GNU Affero General Public License v3.0 only
(AGPL-3.0-only). See [LICENSE](LICENSE) for the full text.

Third-party components bundled or used by Kumo may be licensed under
different terms. See [THIRD_PARTY_NOTICES.md](docs/THIRD_PARTY_NOTICES.md).

## Acknowledgements

- [Mihomo](https://github.com/MetaCubeX/mihomo), the proxy core Kumo drives.
- The Clash and Mihomo ecosystem, for the controller API conventions Kumo speaks.
- [MyGo](https://github.com/egoist/mygo), [Base UI](https://base-ui.com), [coss ui](https://coss.com/ui) and [React Flow](https://reactflow.dev).

## Star History

<a href="https://www.star-history.com/?repos=projectkumo%2Fkumoapp&type=date&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=projectkumo/kumoapp&type=date&theme=dark&legend=top-left" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=projectkumo/kumoapp&type=date&legend=top-left" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=projectkumo/kumoapp&type=date&legend=top-left" />
 </picture>
</a>
