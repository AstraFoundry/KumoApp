# Persistence and Logging

All Kumo processes share one data directory: `$KUMO_HOME`, else the user
configuration directory joined with `Kumo` (`~/Library/Application
Support/Kumo` on macOS, `~/.config/Kumo` on Linux, `%AppData%\Kumo` on
Windows). The layout matches the Swift app's, so its data carries over.

| Path | Holds |
| --- | --- |
| `state.json` (+ `.state.lock`) | the runtime record: state, pid, owner, settings, system proxy record, last exit. Writes take the lock; the Swift app's older format is migrated on read |
| `preferences.json` | app preferences (login, start core, quit on close, language, theme, setup done, updates) |
| `profiles/` | profile YAML, `profiles-metadata.json`, `current.txt` |
| `overrides/` | `overrides.json` and `files/` |
| `work/config.yaml` | the generated runtime config Mihomo runs |
| `work/core.pid`, `work/guardian.pid` | the running core and guardian |
| `cores/mihomo` | the core Kumo downloaded |
| `logs/core.log` | Mihomo's output, with a `--- Kumo started … ---` line per start |
| `agent-skills-state.json` | which agents have the Kumo skill |

JSON files are written to a temporary file and renamed. MyGo keeps window
state and the chosen language in its own app data.

## Logs

- `logs/core.log` collects stdout and stderr of every core Kumo starts.
  `kumo logs path` prints it; `kumo logs --follow` and the Logs page stream
  live logs from the controller, with the file's tail as history.
- A start that fails reports the last error line of the log.
- App diagnostics (shutdown, guardian, refresh errors) go to stderr.
