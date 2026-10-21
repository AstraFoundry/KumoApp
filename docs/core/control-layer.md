# Control Layer

Kumo follows Domain-Driven Design. Dependencies point inward: interfaces →
application → domain, with infrastructure implementing the domain's ports.

## Domain (`internal/domain`)

Pure types and rules, no I/O.

| Package | Holds |
| --- | --- |
| `core` | `Runtime` aggregate (state, pid, owner, readiness, mode, endpoint, settings, system proxy, last exit) and its transitions (`BeginStart`, `Launched`, `Ready`, `Fail`, `Stopped`, `Exited`); `Mode`, `Owner`, `RuntimeSettings` (geo data, TUN, DNS, sniffer); `PortConflict`; the `Store` port |
| `systemproxy` | settings, PAC rendering, `Applied` and `OwnershipOf` (owned / taken over / cleared), the `Platform` port |
| `profile` | profile metadata (compatible with the Swift app's files), usage parsing, refresh due dates, `Repository` and `Fetcher` ports |
| `override` | YAML deep merge with `+key`, `key+`, `prepend-`, `append-`, `delete-`, `key!` and `<key>`; JavaScript overrides through a `Transformer` port |
| `runtimeconfig` | `Build`: profile + overrides + Kumo's controlled keys → the config Mihomo runs |
| `proxy`, `inspect`, `topology` | proxy groups and grades, country flags, connection/rule views, the routing graph |
| `preferences` | app preferences and their store port |

## Application (`internal/application`)

`application.New(deps)` wires one `Kumo` with these services, shared by the
desktop app, the CLI and the guardian:

`Lifecycle`, `SystemProxyService`, `ProfileService`, `OverrideService`,
`ProxyService`, `InspectService`, `SettingsService`, `TopologyService`,
`PreferenceService`, plus `Guard`, `Doctor`, `Guides`/`Prompt`.

`Dependencies` names every adapter, so tests swap in an in-memory system
proxy, a fake fetcher or a short readiness timeout. `Owner` records which
surface (app, CLI) started the core.

### Events

A small in-process `Bus` carries `StatusChanged`, `CoreExited`,
`SystemProxyContended`, `TrafficSampled`, `LogReceived` and
`ProfilesChanged`. The desktop shell forwards them to the frontend as MyGo
events and to the tray and notifications.

## Infrastructure (`internal/infrastructure`)

`paths`, `storage` (JSON files with flock), `mihomo` (REST client, launcher,
discovery, installer), `process`, `ports` (free ports and holders),
`sysproxy` (macOS `networksetup`, Linux `gsettings`, Windows registry, and
file/memory fakes), `pac`, `fetch`, `jsoverride` (goja), `skills`,
`clilink`.

## Interfaces (`internal/interfaces`)

`cli` (the `kumo` command) and `desktop` (MyGo shell and bindings). See
[Interfaces](../interfaces/README.md).
