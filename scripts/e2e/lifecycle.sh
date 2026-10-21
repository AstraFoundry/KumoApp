#!/usr/bin/env bash
# End-to-end lifecycle test of the built Linux app on a virtual display.
# It drives the real window through KUMO_E2E_DRIVER, plays the system proxy
# through a file (KUMO_SYSTEM_PROXY=file:…) and checks that Kumo:
#   1. starts the core from the UI and sets the system proxy;
#   2. on quit, stops the core and clears the proxy it set;
#   3. after SIGKILL, has its guardian stop the core and clear the proxy;
#   4. when another app takes the proxy over, shows it and leaves it alone,
#      also on quit;
#   5. when another program holds its port, says who and starts on a free
#      one only when asked.
#
# Usage: scripts/e2e/lifecycle.sh <mihomo> [app binary]
# Needs Xvfb, dbus-run-session and a WebKitGTK runtime.
set -euo pipefail

MIHOMO=${1:?usage: lifecycle.sh <mihomo> [app]}
APP=${2:-build/linux-amd64/kumo}
WORK=$(mktemp -d)
DRIVER=$WORK/driver
PROXY=$WORK/sysproxy.json
DISPLAY_NUM=:${E2E_DISPLAY:-97}
export KUMO_HOME=$WORK/home KUMO_SYSTEM_PROXY=file:$PROXY KUMO_MIHOMO_PATH=$MIHOMO
export XDG_CONFIG_HOME=$WORK/xdg MYGO_LOCALE=en WEBKIT_DISABLE_COMPOSITING_MODE=1
MIXED=27990 CONTROLLER=29990
APP_PID=

pass() { printf '  \033[32mok\033[0m  %s\n' "$*"; }
fail() { printf '  \033[31mFAIL\033[0m %s\n' "$*"; [ -d "$DRIVER" ] && echo "  app at: $(eval_js 'return location.pathname + " | " + document.body.innerText.replace(/\s+/g, " ").slice(0, 300)')" && tail -5 "$WORK/app.log"; exit 1; }

cleanup() {
  [ -n "$APP_PID" ] && kill -9 "$APP_PID" 2>/dev/null || true
  pkill -xf "$(realpath "$APP")" 2>/dev/null || true # the guardian
  [ -f "$KUMO_HOME/work/core.pid" ] && kill -9 "$(cat "$KUMO_HOME/work/core.pid")" 2>/dev/null || true
  [ -n "${XVFB_PID:-}" ] && kill "$XVFB_PID" 2>/dev/null || true
  rm -rf "$WORK"
}
trap cleanup EXIT

kumo() { "$APP" "$@" --json; }
field() { python3 -c "import json,sys; d=json.load(sys.stdin); print(eval('d'+sys.argv[1]))" "$1"; }

# eval <js>: runs js in the app's page and prints the result's value.
eval_js() {
  rm -f "$DRIVER/out.json"
  printf '%s' "$1" > "$DRIVER/cmd.js"
  for _ in $(seq 1 100); do
    [ -f "$DRIVER/out.json" ] && { field "['value'] if 'value' in d else d['error']" < "$DRIVER/out.json"; return; }
    sleep 0.1
  done
  echo timeout
}

# js_true <js>: whether js returns 1 in the app's page.
js_true() { [ "$(eval_js "$1")" = 1 ]; }

# wait_for <description> <command…>: retries the command for 20 seconds.
wait_for() {
  local what=$1; shift
  for _ in $(seq 1 100); do "$@" >/dev/null 2>&1 && { pass "$what"; return; }; sleep 0.2; done
  fail "$what"
}


core_running() { [ "$(kumo status | field "['data']['state']")" = running ]; }
core_gone() { ! curl -sf "http://127.0.0.1:$CONTROLLER/version" >/dev/null; }
proxy_ours() { grep -q "\"port\": $MIXED" "$PROXY"; }
proxy_cleared() { ! grep -q '"enabled": true' "$PROXY"; }
app_gone() { ! kill -0 "$APP_PID" 2>/dev/null; }

launch() {
  mkdir -p "$DRIVER"
  KUMO_E2E_DRIVER=$DRIVER DISPLAY=$DISPLAY_NUM dbus-run-session -- "$APP" > "$WORK/app.log" 2>&1 &
  local session=$!
  for _ in $(seq 1 50); do APP_PID=$(pgrep -P "$session" -x "$(basename "$APP")" || true); [ -n "$APP_PID" ] && break; sleep 0.1; done
  [ -n "$APP_PID" ] || fail "app launches"
  wait_for "app window answers" js_true 'return 1'
}

click_core_toggle() {
  wait_for "core toggle is ready" js_true 'const b=document.querySelector("[data-e2e=core-toggle]"); return b && !b.disabled ? 1 : 0'
  eval_js 'document.querySelector("[data-e2e=core-toggle]").click(); return 1' >/dev/null
}

Xvfb "$DISPLAY_NUM" -screen 0 1440x900x24 >/dev/null 2>&1 &
XVFB_PID=$!
sleep 1

echo "prepare"
mkdir -p "$KUMO_HOME"
cat > "$WORK/profile.yaml" <<YAML
proxies:
  - { name: node, type: socks5, server: 127.0.0.1, port: 1 }
proxy-groups:
  - { name: Proxy, type: select, proxies: [node, DIRECT] }
rules:
  - MATCH,Proxy
YAML
kumo profile import "$WORK/profile.yaml" --name e2e --activate >/dev/null
kumo config set mixed-port $MIXED >/dev/null
kumo config set controller 127.0.0.1:$CONTROLLER >/dev/null
kumo sysproxy on >/dev/null
echo '{"hasCompletedOnboarding": true, "theme": "system"}' > "$KUMO_HOME/preferences.json"
pass "profile, ports and system proxy preference set"

echo "1. start from the UI"
launch
click_core_toggle
wait_for "core runs" core_running
wait_for "system proxy points at Kumo" proxy_ours

echo "2. quit"
kill -TERM "$APP_PID"
wait_for "app exits" app_gone
wait_for "core stopped" core_gone
wait_for "system proxy cleared" proxy_cleared

echo "3. SIGKILL"
launch
click_core_toggle
wait_for "core runs" core_running
wait_for "system proxy points at Kumo" proxy_ours
kill -KILL "$APP_PID"
wait_for "guardian stops the core" core_gone
wait_for "guardian clears the system proxy" proxy_cleared

echo "4. contention"
launch
click_core_toggle
wait_for "core runs" core_running
wait_for "system proxy points at Kumo" proxy_ours
python3 - "$PROXY" <<'PY'
import json, sys
path = sys.argv[1]
state = json.load(open(path))
p = {"enabled": True, "host": "127.0.0.1", "port": 6152}
state["Wi-Fi"].update(web=p, secure=p, socks=p)
json.dump(state, open(path, "w"), indent=2)
PY
eval_js 'history.pushState(null,"","/"); dispatchEvent(new PopStateEvent("popstate")); return 1' >/dev/null
wait_for "UI shows the takeover" js_true 'return document.querySelector("[data-e2e=contention]") ? 1 : 0'
kill -TERM "$APP_PID"
wait_for "app exits" app_gone
wait_for "core stopped" core_gone
grep -q '"port": 6152' "$PROXY" && pass "other app's proxy left in place" || fail "other app's proxy left in place"

echo "5. port held by another program"
python3 -c "import socket,time; s=socket.socket(); s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1); s.bind(('127.0.0.1', $MIXED)); s.listen(); time.sleep(120)" &
SQUATTER=$!
sleep 0.5
launch
click_core_toggle
wait_for "UI names the conflict" js_true 'return document.querySelector("[data-e2e=port-conflict]")?.textContent.includes("python") ? 1 : 0'
core_gone && pass "core not started on a taken port" || fail "core not started on a taken port"
eval_js 'document.querySelector("[data-e2e=use-free-ports]").click(); return 1' >/dev/null
wait_for "core runs on a free port" core_running
moved=$(kumo status | field "['data']['mixedPort']")
[ "$moved" != "$MIXED" ] && pass "mixed port moved to $moved" || fail "mixed port moved"
kill "$SQUATTER"
kill -TERM "$APP_PID"
wait_for "app exits" app_gone
wait_for "system proxy cleared" proxy_cleared

echo "all lifecycle checks passed"
