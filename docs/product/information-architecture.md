# Information Architecture

Kumo serves one daily job: connect through a chosen node, see that it works,
and get out of the way. Everything else is reachable but not in the way.

## Sidebar

| Group | Pages | Purpose |
| --- | --- | --- |
| Daily | Overview, Profiles, Proxies, Topology | Connect, pick a profile, pick nodes, understand routing |
| Inspect | Connections, Logs, Rules | See what is happening now |
| Configure | Core, System Proxy, DNS, TUN, Sniffer, Resources, Overrides | Change how the core runs |
| Kumo | Agents, Settings, About | The app itself |

The sidebar footer shows whether the core runs and the current throughput.
Pages that read from the running core (Connections, Rules, Resources) show
one empty state with a Start button while it is stopped.

## Overview

Status hero (state, profile, mode, uptime), the start/stop button, the
outbound mode switch, system proxy and TUN switches, a traffic chart and the
selected node of each main group. A takeover of the system proxy by another
app shows here as an alert with a Reclaim action.

## Topology

A React Flow canvas of how traffic is routed, in four columns: entry and
mode → rules (consecutive rules with the same target collapse into one
node; RULE-SET rules stand alone) → proxy groups (nested groups take
sub-columns by depth) → nodes, providers and built-ins. Selected paths are
drawn solid, dynamic membership (`include-all`, `filter`) dashed, broken
references red. Clicking a node focuses its path; the inspector lists its
relations as text. Issues (references to undefined proxies, groups, providers or rule sets) open from the
toolbar.

## First Run

`/setup` takes the whole window: a photo with an animated mock of the
current step on the left, the step on the right. Steps: Welcome, Core,
Profile, System proxy, CLI, Agents, Done. Every step can be skipped;
Settings → Run Setup walks through it again.

## About

App identity over the same backdrop and mock, update status, privacy
promises (local only, talks only to what you add, leaves the system as it
found it), team, license, source and credits.

## Agent Guidance

Each page header has a small, quiet agent button. It copies a prompt that
tells a coding agent how to do one of the page's intents with the `kumo`
command. The Agents page lists every prompt, installs the Kumo skill for
Claude Code, Codex, Cursor, Gemini CLI or `.agents`, and links the `kumo`
command onto the PATH.
