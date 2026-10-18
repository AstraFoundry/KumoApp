# Kumo Agent Guidelines

## Development Documentation

Use these documents as the project map before making architectural, runtime, UI, or workflow changes:

- `docs/README.md` — documentation index and source layout.
- `docs/agents.md` — where each kind of task lives and how to check it.
- `docs/product/information-architecture.md` — product scope and information architecture.
- `docs/interfaces/desktop-interface.md` — MyGo shell, bindings, FSD frontend, components, motion and copy.
- `docs/interfaces/cli-agent-control.md` — CLI and agent-control surfaces.
- `docs/core/control-layer.md` — DDD layers, application services and events.
- `docs/core/mihomo-runtime-controller.md` — Mihomo lifecycle, ports, adoption and the guardian.
- `docs/core/profiles-runtime-configuration.md` — profiles, overrides and the generated runtime config.
- `docs/core/topology.md` — the routing graph and canvas.
- `docs/operations/system-integration-permissions.md` — system proxy ownership, TUN permission and system integration.
- `docs/operations/persistence-logging.md` — data directory, state files and logging.
- `docs/operations/release-management.md` — builds, releases and updates.
- `docs/roadmap/service-mode-roadmap.md` — service mode and deferred work.
- `docs/quality/testing-quality.md` — tests, lifecycle simulations, the end-to-end test and lint.

## Documentation Maintenance

When a change meaningfully alters product behavior, architecture, runtime configuration, persistence, permissions, testing expectations, or UI information architecture, update the relevant document in `docs/` in the same change set. Do not let implementation and documentation drift.

## Backend: Domain-Driven Design

- `internal/domain` holds rules and types with no I/O; `internal/application` holds use cases; `internal/infrastructure` implements the domain's ports; `internal/interfaces` (CLI, desktop) parses input, calls a service and formats output.
- The app, the CLI and the guardian share `internal/application`. Never reimplement lifecycle, configuration or system proxy logic in a surface.
- Kumo must never clear a system proxy it does not own or kill a process that is not its Mihomo. Lifecycle changes need a simulation test and, where the window is involved, `scripts/e2e/lifecycle.sh`.

## Frontend: Feature-Sliced Design

- Layers import only downward: `app` → `pages` → `widgets` → `features` → `entities` → `shared`. Slices of one layer don't import each other; import other slices through their `index.ts`. `bun run lint:fsd` checks this.
- Call Go only through the generated client `src/shared/api/mygo.ts`; regenerate it with `bun run generate` after changing bindings or `locales/`.

## UI Copy Constraints

- Avoid redundant copy. Do not repeat information already expressed by a title, metric, selected state, icon, or surrounding section.
- Prefer concise labels over explanatory text when the UI state is self-evident.
- Remove disabled placeholder actions unless they teach a real next step.
- Do not add low-information detail text such as "Current profile", "selected", or repeated counts when nearby UI already communicates the same fact.
- All user-visible copy lives in `locales/en.json` and `locales/zh-Hans.json`; English is the source. `bun run lint:i18n` checks completeness.

## Component Constraints

- Prefer coss ui components in `src/shared/ui` (Button, Menu, Select, Switch, Toggle Group, Dialog, Alert Dialog, Popover, Tooltip, Table, Empty, Toast) before custom markup. Add new ones with `npx shadcn@latest add @coss/<name>`.
- Do not recreate native selection, menu, dialog or button behavior with overlays or hand-rolled hit targets.
- Keep custom views small and compositional. Extract only when it clarifies state, layout, or reuse.

## Motion

Follow Emil Kowalski's design engineering guidance (`emil-design-eng`): ease-out `cubic-bezier(0.23, 1, 0.32, 1)`, 150–250 ms for UI, 0.97 scale on press, blur crossfades, no animation on frequent actions, and respect `prefers-reduced-motion`.
