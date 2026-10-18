# Testing and Quality

## Commands

| Command | Checks |
| --- | --- |
| `make test` | `go vet` and `go test ./...`: domain rules (overrides, runtime config, system proxy ownership, topology, countries), the JS override runner |
| `KUMO_TEST_MIHOMO=<mihomo> go test ./internal/application` | lifecycle simulations against a real Mihomo, with a file-backed system proxy |
| `make lint` | TypeScript (`tsc --noEmit`), FSD layering (`scripts/check-fsd.mjs`), message catalogs (`scripts/check-i18n.mjs`) |
| `make e2e MIHOMO=<mihomo>` | the built Linux app driven through its real window (`scripts/e2e/lifecycle.sh`) |

CI (`build-release.yml`, job *Test*) runs generate, Go vet/test, typecheck,
FSD and message checks before any build.

## Lifecycle Simulations

`internal/application/lifecycle_sim_test.go` covers: start, select and stop;
the system proxy following the core; a takeover not being clobbered;
reclaim; a proxy turned off elsewhere; a port conflict and its resolution;
a crashed core; adopting a CLI-started core and shutting it down; a stale
pid that is not Mihomo surviving; a broken profile; switching profiles
without a restart and a rejected profile keeping the core running;
concurrent starts; and shutdown.

## End-to-End Lifecycle Test

`scripts/e2e/lifecycle.sh <mihomo> [app]` needs Xvfb, `dbus-run-session` and
WebKitGTK. It runs the app with `KUMO_HOME`, `KUMO_SYSTEM_PROXY=file:…` and
`KUMO_E2E_DRIVER=<dir>` (the app evaluates `cmd.js` in its page and writes
`out.json`), and checks:

1. Start from the UI → core runs, the system proxy points at Kumo.
2. Quit (SIGTERM) → core stopped, proxy cleared.
3. SIGKILL → the guardian stops the core and clears the proxy.
4. Another app takes the proxy over → the alert shows; quitting leaves that
   proxy in place.
5. Another program holds the mixed port → the dialog names it, nothing
   starts until *Use free ports*, then the core runs on the next free port.

## Test Seams

| Variable | Effect |
| --- | --- |
| `KUMO_HOME` | data directory |
| `KUMO_SYSTEM_PROXY=file:<path>` | system proxy kept in a JSON file shared by all processes |
| `KUMO_MIHOMO_PATH` | core to use |
| `KUMO_E2E_DRIVER` | directory through which tests drive the window |
| `MYGO_LOCALE` | force the UI language |

## Manual Checklist (macOS)

- First run: Setup steps, skip, finish with and without starting.
- Start/stop from the window and the tray; mode switch; node selection.
- System proxy in System Settings › Network follows the core; another app
  (e.g. Surge) taking it over shows the alert and is left alone on quit.
- TUN: grant, turn on, traffic flows; turn off.
- Quit with Cmd+Q and from the tray; force-quit in Activity Monitor → proxy
  cleared and core gone within a second.
- `kumo status` from Terminal while the app runs.
- Light and dark appearance; English and Simplified Chinese.
