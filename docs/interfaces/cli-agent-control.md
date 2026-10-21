# CLI and Agent Control

The `kumo` command is the same executable as the app. It runs as the CLI
when started with arguments, or on macOS when invoked through a link named
`kumo`. It shares the data directory and the application services with the
app, so a core started by one is seen, adopted and stopped by the other.

## Output Contract

- Every command takes `--json` and prints `{"ok": bool, "data": …, "error": string | null}`.
- Exit codes: `0` success, `1` the operation failed, `2` usage error.
- Without `--json`, output is a short human summary; `--color` controls color.
- `--dry-run` previews commands that change state where it applies.

## Commands

| Area | Commands |
| --- | --- |
| Lifecycle | `status`, `doctor`, `start [--core <path>] [--resolve-ports]`, `stop`, `restart`, `mode [rule\|global\|direct]` |
| Proxies | `proxies [--group]`, `select <group> <node>`, `delay <group>` / `delay --node <name>` |
| Profiles | `profile list\|current\|use\|import\|refresh\|show\|delete` |
| Overrides | `override list\|add\|show\|enable\|disable\|delete\|order` |
| Inspect | `connections` (+ close), `rules`, `providers` (+ `update`, `geo-update`), `logs [--follow] [path]`, `topology [--issues]` |
| System | `sysproxy status\|on\|off\|services\|set`, `tun status\|on\|off\|grant` |
| Configuration | `dns show\|on\|off`, `sniffer show\|on\|off`, `config show\|path\|set`, `core status\|list\|use\|auto\|install` |
| Agents | `prompt <page> [intent]`, `skills`, `cli-link` |
| Misc | `completion`, `version`, `help`, `-l` (usage of every command) |

`kumo start` refuses to take a port another program holds and reports the
holder with a free port to use; `--resolve-ports` switches to the suggested
ports. `kumo stop` releases only a system proxy Kumo set.

## Prompts

`application.Guides` lists, for each page, intents with a goal and the
`kumo` steps that reach it. `kumo prompt <page> [intent]` and the page
header's agent button print the same prompt:

```text
Goal: <goal>.

Use Kumo's command-line tool, `kumo`, …
- Add --json to every command and read ok, data and error …
Steps:
1. `kumo status --json` — …
Report what you changed and the final `kumo status --json`.
```

## Skills

`kumo skills` (and the Agents page) installs the embedded Kumo skill
(`internal/infrastructure/skills/bundle`) for Claude Code, Codex, Cursor,
Gemini CLI or a project's `.agents`, globally or per project where the
agent supports it, and reports installed, outdated and detected agents.

## CLI Link

`kumo cli-link` (and the Agents page) links the executable as
`/usr/local/bin/kumo` on macOS (asking for an administrator password when
needed) or `~/.local/bin/kumo` on Linux. A `kumo` that is not Kumo's is
reported and left alone.
