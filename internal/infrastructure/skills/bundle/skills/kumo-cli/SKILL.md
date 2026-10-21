---
name: kumo-cli
description: Drive Kumo from coding agents and automation through the `kumo` command-line interface. Use when starting, stopping, inspecting, configuring, or troubleshooting Kumo and its Mihomo core — profiles, proxies, rules, connections, overrides, DNS, sniffer, system proxy, TUN, the routing topology, or Kumo's agent skill installation.
---

# Kumo CLI

## Quick Start

Run `kumo doctor --json` first when diagnosing state. Add `--json` to every
command whose output you read. `kumo` is the Kumo app's own executable,
linked onto PATH from Kumo › Agents; it shares state with the running app.

## Output Contract

- `--json` prints `{ "ok": true|false, "data": ..., "error": "message" | null }`, always all three keys, no color or logs.
- Exit code `0` success, `1` failure (read `error`), `2` usage error.
- Prefer JSON fields over human-readable output.

## Discoverability

- `kumo help` shows common commands; `kumo -l` shows every command's usage.
- `kumo <command> -h` shows focused help.
- `kumo prompt <page> [intent]` prints a step-by-step recipe for a task, for
  pages such as `overview`, `profiles`, `proxies`, `topology`, `dns`, `tun`.
- `kumo completion zsh|bash|fish` emits shell completion.

## Command Map

| Need | Commands |
| --- | --- |
| State | `status`, `doctor` |
| Lifecycle | `start [--resolve-ports]`, `stop`, `restart`, `mode rule\|global\|direct` |
| Nodes | `proxies [--group]`, `select <group> <node>`, `delay <group>`, `delay --node <name>` |
| Profiles | `profile list\|current\|use\|import\|refresh\|show\|delete` |
| Overrides | `override list\|add\|show\|enable\|disable\|delete\|order` |
| Inspect | `connections`, `rules`, `providers`, `logs [--follow]`, `topology [--issues]` |
| System | `sysproxy status\|on\|off\|services\|set`, `tun status\|on\|off\|grant` |
| Config | `dns show\|on\|off`, `sniffer show\|on\|off`, `config show\|path\|set`, `core status\|list\|use\|auto\|install` |
| Skills | `skills status\|install\|uninstall [--agent <name>\|all] [--scope global\|project]` |

## Workflow

1. Read first: `kumo status --json` or `kumo doctor --json`.
2. Use `--dry-run` before commands that change system settings or installed skills.
3. After a write, re-run the narrow read command to verify.
4. When `start` reports a port conflict, tell the user who holds the port; only
   use `--resolve-ports` when they agree to move Kumo to a free port.

## Safety Rules

- Never kill another program to free a port, and never turn off a system proxy
  that another app set — Kumo releases only the proxy it set.
- Don't edit Kumo's state files or the generated config; use `kumo` commands.
- `tun grant` asks for an administrator password: run it only when the user asks.
- Don't pass `--force` to skill installation unless the user wants an existing
  skill directory replaced.
