# Mihomo Runtime Controller

`application.Lifecycle` owns the core process. The runtime record
(`state.json`, guarded by `.state.lock`) is shared by every Kumo process, so
the app, the CLI and the guardian agree on what runs.

## Start

1. Refuse when a core already runs (`ErrAlreadyRunning`).
2. Resolve the core, first found wins: `--core`, the chosen path, the
   managed core (`cores/mihomo`), `KUMO_MIHOMO_PATH`, one bundled with the
   app, `PATH`, then common install directories.
3. Check ports. The mixed port and the controller must be free; a port held
   by another program is a `PortConflict` naming the holder (pid, process)
   and a suggested free port. Nothing is killed. The caller decides: the app
   shows a dialog, the CLI prints the conflict, and `resolvePorts` moves to
   the suggested ports.
4. Build the runtime config (profiles were validated when imported).
5. Launch detached (`-d work -f work/config.yaml -ext-ctl … -secret …`),
   with proxy variables cleared, record pid, owner and `processLaunched`.
6. Ready when the controller answers **and** `/proxies` lists groups;
   otherwise fail with the last error line of `logs/core.log`.
7. Apply the system proxy when it is enabled.

## Reload, Mode and Settings

Profile switches, override changes and DNS/sniffer/geo edits rebuild the
config and `PUT /configs?force=true` without restarting. A rejected config
leaves the running core as it was and reports the error. Turning TUN on, or
changing the controller, restarts. `mode` patches `/configs`.

## Stop and Shutdown

`Stop` releases the system proxy first, then terminates the core
(SIGINT → SIGTERM → SIGKILL), waits for its ports and records the stop.
`Shutdown` (every app quit) does the same for whatever core runs and
always stops the PAC server.

## Monitoring and Adoption

`Monitor` refreshes the state every 3 s: a core that died is recorded as
`failed` with its exit and a `CoreExited` event; a takeover of the system
proxy is detected (see [system integration](../operations/system-integration-permissions.md)).
A core started by the CLI is adopted by the app, and vice versa. A stale
pid file never kills a process unless it is Mihomo (`process.IsCore`).

## Guardian

The app spawns a copy of itself with `KUMO_GUARDIAN_PARENT=<pid>`. It waits
for the app to exit. If the app quit properly there is nothing left to do;
after a crash or SIGKILL it shuts the app's core down and releases the
system proxy Kumo set. On the next launch `Recover` adopts a surviving
core or releases a proxy left behind.

## Verification

- `internal/application/lifecycle_sim_test.go` runs real Mihomo
  (`KUMO_TEST_MIHOMO=<path> go test ./internal/application`) through start,
  selection, stop, proxy ownership, takeover and reclaim, port conflicts,
  crashes, adoption, stale pids, broken profiles, profile switches,
  concurrent starts and shutdown.
- `scripts/e2e/lifecycle.sh <mihomo>` drives the built app (see
  [testing](../quality/testing-quality.md)).
