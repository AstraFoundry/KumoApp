import type { Edge as FlowEdge, Node as FlowNode } from "@xyflow/react";
import type { Edge, Graph, TopologyNode } from "@/shared/api/mygo";

export const COLUMN_WIDTH = 236;
export const COLUMN_GAP = 84;
export const ROW_HEIGHT = 46;
export const NODE_HEIGHT = 36;
export const HEADER_OFFSET = 44;
/** A column shows this many nodes until expanded. */
export const COLLAPSED_ROWS = 10;

export interface TopologyNodeData extends Record<string, unknown> {
  node: TopologyNode;
  /** dimmed when another node is hovered or focused and this one is off its path. */
  dimmed: boolean;
  highlighted: boolean;
  focused: boolean;
  /** For the "N more" placeholder of a collapsed column. */
  more?: number;
}

export interface LayoutOptions {
  /** Show members brought in by include-all and filters. */
  dynamic: boolean;
  /** Columns showing every node. */
  expanded: Set<number>;
  /** The node the inspector shows; it stays visible in collapsed columns. */
  focus?: string;
  /** The node under the pointer. */
  hover?: string;
  /** "path" keeps only the focused node's upstream and downstream. */
  scope: "all" | "path";
}

/** Edges of the graph the canvas draws under the options. */
function visibleEdges(graph: Graph, options: LayoutOptions): Edge[] {
  return graph.edges.filter((e) => options.dynamic || e.kind !== "dynamic");
}

/**
 * The nodes upstream and downstream of id: everything traffic passes
 * through to reach it, and everything it can lead to.
 */
export function pathOf(id: string, edges: Edge[]): Set<string> {
  const out = new Map<string, string[]>();
  const into = new Map<string, string[]>();
  for (const e of edges) {
    out.set(e.source, [...(out.get(e.source) ?? []), e.target]);
    into.set(e.target, [...(into.get(e.target) ?? []), e.source]);
  }
  const walk = (map: Map<string, string[]>) => {
    const seen = new Set([id]);
    const stack = [id];
    while (stack.length) {
      const current = stack.pop() as string;
      for (const next of map.get(current) ?? []) {
        if (!seen.has(next)) {
          seen.add(next);
          stack.push(next);
        }
      }
    }
    return seen;
  };
  return new Set([...walk(out), ...walk(into)]);
}

/**
 * How deep each group sits: groups no other group lists are 0, and a group
 * listed by another sits one further right, so nested groups read left to
 * right instead of looping back within one column.
 */
function groupDepths(graph: Graph, edges: Edge[]): Map<string, number> {
  const groups = new Set(graph.nodes.filter((n) => n.column === 2).map((n) => n.id));
  const children = new Map<string, string[]>();
  for (const e of edges) {
    if (groups.has(e.source) && groups.has(e.target) && e.source !== e.target) {
      children.set(e.source, [...(children.get(e.source) ?? []), e.target]);
    }
  }
  const depth = new Map<string, number>();
  const visit = (id: string, d: number, trail: Set<string>) => {
    if (trail.has(id) || (depth.get(id) ?? -1) >= d || d > 6) return;
    depth.set(id, d);
    trail.add(id);
    for (const child of children.get(id) ?? []) visit(child, d + 1, trail);
    trail.delete(id);
  };
  const referenced = new Set([...children.values()].flat());
  for (const id of groups) if (!referenced.has(id)) visit(id, 0, new Set());
  for (const id of groups) if (!depth.has(id)) visit(id, 0, new Set());
  return depth;
}

/** Where each column starts: groups may span several sub-columns. */
export function columnX(column: number, depth: number, maxDepth: number): number {
  const step = COLUMN_WIDTH + COLUMN_GAP;
  if (column < 2) return column * step;
  if (column === 2) return (2 + depth) * step;
  return (3 + maxDepth) * step;
}

/**
 * Places the graph in four columns, left to right as traffic flows, keeping
 * the configuration's order in each, and collapses long columns.
 */
export function layout(
  graph: Graph,
  options: LayoutOptions,
): { nodes: FlowNode<TopologyNodeData>[]; edges: FlowEdge[]; hidden: number; headers: { x: number; column: number }[] } {
  const edges = visibleEdges(graph, options);
  const anchor = options.hover ?? options.focus;
  const path = anchor ? pathOf(anchor, edges) : undefined;
  const scoped = options.scope === "path" && options.focus ? pathOf(options.focus, edges) : undefined;

  const depths = groupDepths(graph, edges);
  const maxDepth = Math.max(0, ...depths.values());
  // Columns are keyed by their x: nested groups make sub-columns.
  const columns = new Map<number, TopologyNode[]>();
  for (const n of graph.nodes) {
    if (scoped && !scoped.has(n.id)) continue;
    const x = columnX(n.column, depths.get(n.id) ?? 0, maxDepth);
    columns.set(x, [...(columns.get(x) ?? []), n]);
  }

  const nodes: FlowNode<TopologyNodeData>[] = [];
  const shown = new Set<string>();
  let hidden = 0;
  for (const [x, list] of columns) {
    const column = x;
    const limit = options.expanded.has(column) || scoped ? list.length : COLLAPSED_ROWS;
    // Keep the focused node and the active path visible when collapsing.
    const keep = list.filter((n, i) => i < limit || n.id === options.focus || (path?.has(n.id) && list.length - limit < 40));
    const rest = list.length - keep.length;
    keep.forEach((n, row) => {
      shown.add(n.id);
      nodes.push({
        id: n.id,
        type: "topology",
        position: { x, y: HEADER_OFFSET + row * ROW_HEIGHT },
        data: {
          node: n,
          dimmed: path !== undefined && !path.has(n.id),
          highlighted: path !== undefined && path.has(n.id),
          focused: n.id === options.focus,
        },
        draggable: false,
        connectable: false,
        selectable: true,
      });
    });
    if (rest > 0) {
      hidden += rest;
      nodes.push({
        id: `more:${column}`,
        type: "more",
        position: { x, y: HEADER_OFFSET + keep.length * ROW_HEIGHT },
        data: { node: { id: `more:${column}`, kind: "entry", column: column as 0, label: "", order: 0 }, dimmed: false, highlighted: false, focused: false, more: rest },
        draggable: false,
        connectable: false,
      });
    }
  }

  const flowEdges: FlowEdge[] = edges
    .filter((e) => shown.has(e.source) && shown.has(e.target))
    .map((e) => {
      const onPath = path !== undefined && path.has(e.source) && path.has(e.target);
      const dimmed = path !== undefined && !onPath;
      return {
        id: e.id,
        source: e.source,
        target: e.target,
        type: "default",
        zIndex: e.active || onPath ? 1 : 0,
        className: [
          "kumo-edge",
          `kumo-edge-${e.kind}`,
          e.active ? "is-active" : "",
          e.broken ? "is-broken" : "",
          onPath ? "is-path" : "",
          dimmed ? "is-dimmed" : "",
        ].join(" "),
      };
    });
  const headers = [0, 1, 2, 3].map((column) => ({ column, x: columnX(column, 0, maxDepth) }));
  return { nodes, edges: flowEdges, hidden, headers };
}
