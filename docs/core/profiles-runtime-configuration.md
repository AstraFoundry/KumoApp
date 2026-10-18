# Profiles and Runtime Configuration

## Profiles

A profile is a Clash/Mihomo YAML document with metadata in
`profiles/profiles-metadata.json` (the same format the Swift app wrote, so
existing profiles carry over). Kinds:

- `remote` — a subscription URL. Imports read `subscription-userinfo`
  (usage and expiry), `profile-update-interval` and `content-disposition`
  (name). A refresh can go through Kumo's own proxy (`useProxy`).
- `local` — a file or pasted YAML.
- `inline` — the built-in default (`DIRECT` only), editable in place.

Every import and edit is parsed and checked before it is saved; a broken
YAML never replaces a working one. `current.txt` names the active profile.

Switching, refreshing or editing the active profile while the core runs
reloads it (`PUT /configs?force=true`); a config the core rejects leaves it
running as it was. The app refreshes remote profiles whose interval passed
(`autoUpdate`) every 10 minutes; `kumo profile refresh` does it on demand.

Profiles that the Swift app generated through its bundled Sub-Store keep
their last content. Refreshing them needs Sub-Store, which is deferred (see
the [roadmap](../roadmap/service-mode-roadmap.md)).

## Overrides

Overrides apply to every profile, top to bottom (`overrides/overrides.json`,
contents in `overrides/files`). Each is local text or a URL, YAML or
JavaScript, and can be disabled.

- **YAML** deep-merges into the profile. Keys can carry operators:
  `+rules` / `rules+` prepend or append to a list, `prepend-rules` and
  `append-rules` do the same, `delete-key` removes a key, `key!` replaces
  instead of merging, `<key>` forces a literal key name.
- **JavaScript** defines `main(config)` and returns the changed config, as
  in Clash Verge. It runs in goja with a 5 s timeout and no I/O.

## Runtime Configuration

`runtimeconfig.Build` produces `work/config.yaml` from the profile, the
enabled overrides and Kumo's controlled keys, which always win:

- `external-controller`, `secret`, `mixed-port`, `mode`, `allow-lan`,
  `log-level`, `ipv6`, `find-process-mode`;
- geo data: `geodata-mode`, `geo-auto-update`, `geo-update-interval`,
  `geox-url`;
- `tun` when Kumo's TUN is on;
- `dns` (and `hosts`) when *Manage DNS* is on — otherwise the profile's own
  `dns` applies;
- `sniffer` when *Manage sniffer* is on — otherwise the profile's applies.

The profile's other listeners (`port`, `socks-port`, `redir-port`,
`tproxy-port`) stay as written. DNS policy values accept a string or a list
([ADR-002](../decisions/ADR-002-policy-value-types.md)); `hosts` is
top-level ([ADR-003](../decisions/ADR-003-hosts-top-level-vs-nested.md)).

`kumo config show` and Core → Running config show the generated file.
