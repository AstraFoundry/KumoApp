# Topology

`domain/topology.Build(config, mode, selections)` turns the running (or
generated) config into a graph for the Topology page and `kumo topology`.

## Nodes

| Column | Kinds |
| --- | --- |
| 0 | `entry` (the inbound and mode), `global` in global mode |
| 1 | `rules` (consecutive rules with one target), `ruleSet` (each RULE-SET), `final` (MATCH) |
| 2 | `group` (proxy groups; nested groups get sub-columns by depth in the canvas) |
| 3 | `proxy`, `provider`, `builtin` (DIRECT, REJECT…), `missing` |

## Edges

- `route`, `member` and `provider` edges follow the configuration.
- `dynamic` edges come from `include-all`, `include-all-proxies`,
  `include-all-providers` and `filter`; the canvas hides them unless
  *Dynamic members* is on and draws them dashed.
- `active` edges are the path the mode and current selections take.
- `broken` edges point at a `missing` node.

## Issues

References to an undefined proxy, group, provider or rule set become
issues, listed in the toolbar popover and by `kumo topology --issues`.

## Canvas (`widgets/topology-canvas`)

React Flow with a custom layout (`lib/layout.ts`): fixed columns, rows
ordered by first reference, columns over 10 rows collapse into a "more"
node. Clicking a node focuses its path (scope *Path*); the inspector shows
the node's relations in text. Fit keeps a minimum zoom of 0.7 so labels
stay readable; Find jumps to a node by name.
