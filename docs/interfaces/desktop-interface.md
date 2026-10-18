# Desktop Interface

## Shell (`internal/interfaces/desktop`)

- **Run** takes the single-instance lock (a second launch focuses the first
  window), registers `kumo://` and `clash://install-config?url=…` (opens
  Profiles with the import prefilled) and binds the services below.
- **Window**: one main window, 1180×780 (min 940×620), state remembered
  under `main`. On macOS the title bar is hidden and the traffic lights sit
  at (18, 20) over the sidebar; on Linux and Windows the system title bar
  keeps its own controls, so page headers never reserve room for them. It opens
  at `/setup` until setup completes. Closing it keeps Kumo in the menu bar
  unless *Quit when the window closes* is on.
- **Tray**: a template cloud icon (filled while the core runs) with status,
  start/stop, outbound mode, system proxy, TUN, proxy groups, Open and Quit.
  *Hide menu bar icon* applies at the next launch.
- **Quit path**: every quit (menu, tray, Cmd+Q, SIGINT/SIGTERM) goes through
  `OnBeforeQuit`, which holds the quit, broadcasts `quitting`, releases the
  system proxy Kumo set, stops the core it started and then lets the app
  exit. A crash or SIGKILL skips this; the guardian (see
  [the runtime controller](../core/mihomo-runtime-controller.md#guardian))
  covers it.
- **Background work** while the app runs: `Lifecycle.Monitor` (state and
  proxy takeover every 3 s), traffic streaming, subscription refresh every
  10 min, an update check 30 s after launch when enabled, and
  notifications for core exits and proxy takeovers.

## Bindings

`desktop.Bind` registers services that MyGo turns into the typed client
`src/shared/api/mygo.ts` (`bun run generate`): `Runtime`, `SystemProxy`,
`Profiles`, `Overrides`, `Proxies`, `Inspect`, `Settings`, `Topology`,
`Preferences` and `App`. Events: `status`, `traffic`, `contention`,
`coreExited`, `profilesChanged`, `navigate`, `quitting`, `updateProgress`.
Logs stream over a channel (`Logs.follow`).

## Frontend (`src/`, Feature-Sliced Design)

| Layer | Holds | Examples |
| --- | --- | --- |
| `app` | entry, router, providers, global styles | `router.tsx` lazy-loads every page but Overview |
| `pages` | one slice per route | `overview`, `topology`, `setup`, `about` |
| `widgets` | composed blocks | `app-sidebar`, `page-header`, `topology-canvas`, `showcase` |
| `features` | one user action each | `toggle-core`, `toggle-system-proxy`, `copy-agent-prompt`, `import-profile` |
| `entities` | domain views and stores | `status`, `proxy`, `profile` |
| `shared` | API client, i18n, ui kit, lib | `shared/ui` holds the coss ui components |

A layer imports only from layers below it; slices of one layer do not
import each other; other slices are imported through their `index.ts`.
`bun run lint:fsd` checks this.

### Components

UI is built from [coss ui](https://coss.com/ui) (Base UI) components in
`src/shared/ui`, added with `npx shadcn@latest add @coss/<name>`. Prefer a
coss component (Button, Menu, Select, Switch, Toggle Group, Dialog, Alert
Dialog, Popover, Tooltip, Table, Empty, Toast) over custom markup.
Kumo's own shared pieces are `PageBody`/`Section`, `SettingsGroup`/
`SettingRow`, `ListEditor`, `MapEditor`, `CodeEditor`, `ApplyBar` and
`SearchField`.

### Layout and Alignment

- **One gutter**: 20px. Page header (title and actions), page body and
  canvas overlays share it, so the title, the first card and the first
  topology column start on the same line. The body stays left-aligned
  under the title at any window width (About is the one centered page).
- **Header actions** are the page's controls in one row, all 28px high
  (search, select, `sm` buttons, ghost icon buttons), with the agent
  button last. Pages that read the running core show their actions only
  while it runs (`WhenRunning`); no disabled placeholders.
- **Canvas overlays** (topology) share one panel style, 36px high: view
  controls top-left, legend bottom-left, zoom bottom-right. Search and
  issues belong in the page header. The graph fits once React Flow has
  measured the nodes.
- **Setup** keeps its stepper and language, the step, and Back/Skip and
  Continue in one 28rem column: every row shares both edges. Ghost
  buttons at a column edge are offset by their padding so their label,
  not their box, meets the edge.

### Settings that reload the core

Core, DNS, TUN and Sniffer edit a draft. `ApplyBar` appears with Discard
and Apply once the draft differs; Apply writes the settings and reloads
the running core (`PUT /configs?force=true`), restarting it only when TUN
turns on.

### Motion

Motion follows Emil Kowalski's design engineering guidance:

- ease-out `cubic-bezier(0.23, 1, 0.32, 1)` for things that appear, 150–250 ms;
- buttons scale to 0.97 while pressed;
- crossfades blur slightly (`filter: blur`) instead of sliding;
- frequent actions (switching nodes, toggling rules) don't animate;
- `prefers-reduced-motion` turns animation off.

### Copy

English and Simplified Chinese live in `locales/*.json` (ICU messages) and
generate `locales/catalog_gen.go` and `src/shared/api/locales.ts`. Pages
read `t.key()`. `bun run lint:i18n` lists keys pages use that a catalog
lacks. Copy stays short and never repeats what a title, value or icon
already says.

## Agent Prompt Button

`features/copy-agent-prompt` puts a ghost icon button in each page header.
Its menu lists the page's intents (from `application.Guides`) and copies
the prompt for one, the same text `kumo prompt <page> [intent]` prints.
