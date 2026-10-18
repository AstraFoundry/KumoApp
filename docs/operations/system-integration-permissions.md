# System Integration and Permissions

## System Proxy

Kumo sets the system proxy only while its core runs and only on the network
service the user chose (*Automatic* follows the active one).

| Platform | Adapter |
| --- | --- |
| macOS | `networksetup` (web, secure web, SOCKS, bypass, auto proxy URL) |
| Linux | `gsettings` (GNOME `org.gnome.system.proxy`) |
| Windows | `HKCU\…\Internet Settings` |
| Tests | `KUMO_SYSTEM_PROXY=file:<path>` — a JSON file shared by every process |

**Manual** points HTTP, HTTPS and SOCKS at `127.0.0.1:<mixed-port>` with the
bypass list. **PAC** serves the script on a loopback port from the app (the
CLI refuses PAC, since it would stop serving when the command exits).

### Ownership

When Kumo applies the proxy it records what it set (`Applied`) and what was
there before. Each check (every 3 s while the app runs) compares the system
with that record:

- **owned** — the system still shows Kumo's proxy.
- **taken over** — another app set its own proxy. Kumo leaves it, shows an
  alert with *Reclaim* and notifies, naming the program that listens on the
  new port when it can. With *Reclaim when taken over* on, Kumo sets its
  proxy again.
- **cleared** — someone turned the proxy off. Kumo turns its switch off and
  does not fight it.

Release (stop, quit, guardian) disables the proxy **only when it is still
owned**; a proxy another app took over is never cleared.

## TUN Permission

TUN needs Mihomo to create a network interface. Instead of a privileged
helper service, Kumo grants that once, to the core binary:

- macOS: `chown root:admin` and `chmod u+s` through an administrator prompt.
- Linux: `pkexec setcap cap_net_admin,cap_net_bind_service=+eip`.

The TUN page shows whether the core has permission (root, setuid or
capability) and offers *Grant…*. A replaced core binary needs a new grant.

## Other Integration

- **CLI link**: `/usr/local/bin/kumo` (macOS, administrator prompt when
  needed) or `~/.local/bin/kumo` (Linux) pointing at the app's executable.
- **Open at login**: MyGo's login item; a launch at login stays in the menu
  bar without opening the window once setup is done.
- **URL schemes**: `kumo://` and `clash://install-config?url=<subscription>`.
- **Notifications**: core exits, proxy takeovers and reclaims, available
  updates.
- **Single instance**: a second launch focuses the running app.
