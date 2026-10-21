import "@xyflow/react/dist/base.css";
import {
  type NodeMouseHandler,
  ReactFlow,
  ReactFlowProvider,
  useReactFlow,
  ViewportPortal,
} from "@xyflow/react";
import { AlertTriangleIcon, LocateFixedIcon, MinusIcon, PlusIcon } from "lucide-react";
import { type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useStatus } from "@/entities/status";
import { type Graph, Proxies, Topology } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { useResource } from "@/shared/lib/use-resource";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { Label } from "@/shared/ui/label";
import { Popover, PopoverPopup, PopoverTrigger } from "@/shared/ui/popover";
import { SearchField } from "@/shared/ui/search-field";
import { Switch } from "@/shared/ui/switch";
import { toastManager } from "@/shared/ui/toast";
import { ToggleGroup, ToggleGroupItem } from "@/shared/ui/toggle-group";
import { COLUMN_WIDTH, layout } from "../lib/layout";
import { Inspector } from "./inspector";
import { MoreNodeView, TopologyNodeView } from "./topology-node";

const nodeTypes = { topology: TopologyNodeView, more: MoreNodeView };

const columnTitles = [() => t.topoColumnEntry(), () => t.topoColumnRules(), () => t.topoColumnGroups(), () => t.topoColumnNodes()];

// Every control floating on the canvas shares this chrome and height, so
// panels line up wherever they sit.
const panel = "flex h-9 items-center rounded-lg border bg-background/90 p-[3px] shadow-xs/5 backdrop-blur";

/**
 * The configuration as a graph, from the outbound mode to the nodes.
 * `header` renders the page header with the canvas's actions (find,
 * issues), so they sit with the page's other header controls.
 */
export function TopologyCanvas({ header }: { header: (actions: ReactNode) => ReactNode }) {
  return (
    <ReactFlowProvider>
      <Canvas header={header} />
    </ReactFlowProvider>
  );
}

function Canvas({ header }: { header: (actions: ReactNode) => ReactNode }) {
  const status = useStatus();
  const running = status?.state === "running";
  const graph = useResource(() => Topology.graph(), [status?.state, status?.mode, status?.profile.id], {
    intervalMs: running ? 5000 : undefined,
  });
  const [focus, setFocus] = useState<string>();
  const [hover, setHover] = useState<string>();
  const [dynamic, setDynamic] = useState(true);
  const [scope, setScope] = useState<"all" | "path">("all");
  const [expanded, setExpanded] = useState<Set<number>>(new Set());
  const [search, setSearch] = useState("");
  const flow = useReactFlow();

  const data: Graph | undefined = graph.data;
  const laid = useMemo(
    () => (data ? layout(data, { dynamic, expanded, focus, hover, scope }) : { nodes: [], edges: [], hidden: 0, headers: [] }),
    [data, dynamic, expanded, focus, hover, scope],
  );
  const focused = data?.nodes.find((n) => n.id === focus);

  const container = useRef<HTMLDivElement>(null);

  // Fit the graph, but never below a readable zoom: a large configuration
  // starts at its top-left corner and pans, rather than shrinking to dots.
  const fit = useCallback(
    (duration = 0) => {
      const box = container.current?.getBoundingClientRect();
      // Bounds by id read React Flow's internal nodes, which carry the
      // measured sizes; the controlled nodes we pass in never do.
      const ids = flow.getNodes().filter((n) => !n.hidden).map((n) => n.id);
      if (!box || ids.length === 0) return;
      const bounds = flow.getNodesBounds(ids);
      // x is the page gutter: the first column lines up with the title and panels.
      const pad = { x: 20, top: 64, bottom: 60 };
      const zoom = Math.min(1, Math.max(0.7, Math.min((box.width - pad.x * 2) / bounds.width, (box.height - pad.top - pad.bottom) / bounds.height)));
      const width = bounds.width * zoom;
      const height = bounds.height * zoom;
      const x = width < box.width - pad.x * 2 ? (box.width - width) / 2 : pad.x;
      const y = height < box.height - pad.top - pad.bottom ? pad.top + (box.height - pad.top - pad.bottom - height) / 2 : pad.top;
      void flow.setViewport({ x: x - bounds.x * zoom, y: y - bounds.y * zoom, zoom }, { duration });
    },
    [flow],
  );

  // Fit when the first graph arrives and whenever the scope changes, once
  // React Flow has measured the nodes (their bounds are empty before).
  // biome-ignore lint/correctness/useExhaustiveDependencies: fit on these changes only
  useEffect(() => {
    if (!data) return;
    let frame = 0;
    let tries = 0;
    const attempt = () => {
      const nodes = flow.getNodes().filter((n) => !n.hidden);
      if (nodes.length > 0 && nodes.every((n) => flow.getInternalNode(n.id)?.measured.width)) fit();
      else if (tries++ < 60) frame = requestAnimationFrame(attempt);
    };
    frame = requestAnimationFrame(attempt);
    return () => cancelAnimationFrame(frame);
  }, [data === undefined, scope, focus && scope === "path"]);

  const onNodeClick: NodeMouseHandler = (_, node) => {
    if (node.id.startsWith("more:")) {
      setExpanded((s) => new Set(s).add(Number(node.id.slice(5))));
      return;
    }
    setFocus((f) => (f === node.id ? undefined : node.id));
  };

  function find(e: React.FormEvent) {
    e.preventDefault();
    const q = search.trim().toLowerCase();
    if (!q || !data) return;
    const match = data.nodes.find((n) => n.label.toLowerCase().includes(q));
    if (!match) {
      toastManager.add({ title: t.topoNotFound(search), type: "info" });
      return;
    }
    setFocus(match.id);
    requestAnimationFrame(() => {
      const placed = flow.getNode(match.id);
      if (placed) void flow.setCenter(placed.position.x + COLUMN_WIDTH / 2, placed.position.y + 18, { zoom: 1, duration: 300 });
    });
  }

  async function selectMember(group: string, member: string) {
    try {
      await Proxies.select(group, member);
      await graph.reload();
    } catch (err) {
      toastManager.add({ title: t.selectFailed(), description: errorMessage(err), type: "error" });
    }
  }

  const actions = (
    <>
      {data && data.issues.length > 0 && (
        <Popover>
          <PopoverTrigger render={<Button size="sm" variant="outline" />}>
            <AlertTriangleIcon aria-hidden="true" className="text-warning" />
            {t.topoIssues(String(data.issues.length))}
          </PopoverTrigger>
          <PopoverPopup align="end" className="w-96">
            <ul className="flex flex-col gap-2 text-xs">
              {data.issues.map((issue) => (
                <li key={issue.message}>
                  <button className="text-left hover:underline" onClick={() => setFocus(`missing:${issue.missing}`)} type="button">
                    {issue.message}
                  </button>
                </li>
              ))}
            </ul>
          </PopoverPopup>
        </Popover>
      )}
      <form onSubmit={find}>
        <SearchField label={t.topoFind()} onChange={setSearch} value={search} />
      </form>
    </>
  );

  if (graph.error && !data) {
    return (
      <>
      {header(null)}
      <Empty className="flex-1">
        <EmptyHeader>
          <EmptyTitle>{t.topoError()}</EmptyTitle>
          <EmptyDescription data-selectable>{errorMessage(graph.error)}</EmptyDescription>
        </EmptyHeader>
      </Empty>
      </>
    );
  }

  return (
    <>
    {header(actions)}
    <div className="flex min-h-0 flex-1">
      <div className="relative min-w-0 flex-1 bg-muted/20" ref={container}>
        <ReactFlow
          className="kumo-flow"
          edges={laid.edges}
          maxZoom={1.6}
          minZoom={0.25}
          nodeTypes={nodeTypes}
          nodes={laid.nodes}
          nodesConnectable={false}
          nodesDraggable={false}
          onNodeClick={onNodeClick}
          onNodeMouseEnter={(_, node) => !node.id.startsWith("more:") && setHover(node.id)}
          onNodeMouseLeave={() => setHover(undefined)}
          onPaneClick={() => setFocus(undefined)}
          panOnScroll
          proOptions={{ hideAttribution: true }}
          zoomOnDoubleClick={false}
        >
          <ViewportPortal>
            {laid.headers.map(({ column, x }) => (
              <div
                className="absolute font-medium text-[0.6875rem] text-muted-foreground uppercase tracking-wide"
                key={column}
                style={{ transform: `translate(${x + 4}px, 6px)` }}
              >
                {columnTitles[column]?.()}
              </div>
            ))}
          </ViewportPortal>
        </ReactFlow>

        <div className="pointer-events-none absolute top-3 left-5">
          <div className={`${panel} pointer-events-auto gap-2`}>
            <ToggleGroup onValueChange={(v) => v[0] && setScope(v[0] as "all" | "path")} size="sm" value={[scope]} variant="default">
              <ToggleGroupItem className="px-2.5" value="all">
                {t.topoScopeAll()}
              </ToggleGroupItem>
              <ToggleGroupItem className="px-2.5" disabled={!focus} value="path">
                {t.topoScopePath()}
              </ToggleGroupItem>
            </ToggleGroup>
            <Label className="flex items-center gap-2 pe-2 font-normal text-xs">
              <Switch checked={dynamic} onCheckedChange={setDynamic} />
              {t.topoDynamicMembers()}
            </Label>
          </div>
        </div>

        <div className="pointer-events-none absolute inset-x-5 bottom-3 flex items-end justify-between gap-2">
          <Legend hidden={laid.hidden} />
          <div className={`${panel} pointer-events-auto`}>
            <Button aria-label={t.topoZoomOut()} onClick={() => flow.zoomOut({ duration: 150 })} size="icon-sm" variant="ghost">
              <MinusIcon aria-hidden="true" />
            </Button>
            <Button aria-label={t.topoZoomIn()} onClick={() => flow.zoomIn({ duration: 150 })} size="icon-sm" variant="ghost">
              <PlusIcon aria-hidden="true" />
            </Button>
            <Button aria-label={t.topoFit()} onClick={() => fit(250)} size="icon-sm" variant="ghost">
              <LocateFixedIcon aria-hidden="true" />
            </Button>
          </div>
        </div>
      </div>
      {focused && data && (
        <Inspector
          graph={data}
          live={running}
          node={focused}
          onClose={() => setFocus(undefined)}
          onFocus={(id) => {
            setFocus(id);
          }}
          onSelectMember={selectMember}
        />
      )}
    </div>
    </>
  );
}

function Legend({ hidden }: { hidden: number }) {
  const items = [
    { className: "kumo-legend-active", label: t.topoLegendActive() },
    { className: "kumo-legend-member", label: t.topoLegendMember() },
    { className: "kumo-legend-dynamic", label: t.topoLegendDynamic() },
    { className: "kumo-legend-broken", label: t.topoLegendBroken() },
  ];
  return (
    <div className={`${panel} pointer-events-none gap-4 px-3 text-[0.6875rem] text-muted-foreground`}>
      {items.map((item) => (
        <span className="flex items-center gap-1.5" key={item.label}>
          <svg aria-hidden="true" className={item.className} height="6" width="22">
            <line x1="0" x2="22" y1="3" y2="3" />
          </svg>
          {item.label}
        </span>
      ))}
      {hidden > 0 && <Badge variant="secondary">{t.topoHidden(String(hidden))}</Badge>}
    </div>
  );
}
