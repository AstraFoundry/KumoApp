import "@xyflow/react/dist/base.css";
import {
  type NodeMouseHandler,
  ReactFlow,
  ReactFlowProvider,
  useReactFlow,
  ViewportPortal,
} from "@xyflow/react";
import { AlertTriangleIcon, LocateFixedIcon, MinusIcon, PlusIcon, SearchIcon } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useStatus } from "@/entities/status";
import { type Graph, Proxies, Topology } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { useResource } from "@/shared/lib/use-resource";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/shared/ui/input-group";
import { Label } from "@/shared/ui/label";
import { Popover, PopoverPopup, PopoverTrigger } from "@/shared/ui/popover";
import { Switch } from "@/shared/ui/switch";
import { toastManager } from "@/shared/ui/toast";
import { ToggleGroup, ToggleGroupItem } from "@/shared/ui/toggle-group";
import { COLUMN_WIDTH, layout } from "../lib/layout";
import { Inspector } from "./inspector";
import { MoreNodeView, TopologyNodeView } from "./topology-node";

const nodeTypes = { topology: TopologyNodeView, more: MoreNodeView };

const columnTitles = [() => t.topoColumnEntry(), () => t.topoColumnRules(), () => t.topoColumnGroups(), () => t.topoColumnNodes()];

/** The configuration as a graph, from the outbound mode to the nodes. */
export function TopologyCanvas() {
  return (
    <ReactFlowProvider>
      <Canvas />
    </ReactFlowProvider>
  );
}

function Canvas() {
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
      if (!box || laid.nodes.length === 0) return;
      const bounds = flow.getNodesBounds(laid.nodes);
      const pad = { x: 32, top: 64, bottom: 56 };
      const zoom = Math.min(1, Math.max(0.7, Math.min((box.width - pad.x * 2) / bounds.width, (box.height - pad.top - pad.bottom) / bounds.height)));
      const width = bounds.width * zoom;
      const height = bounds.height * zoom;
      const x = width < box.width - pad.x * 2 ? (box.width - width) / 2 : pad.x;
      const y = height < box.height - pad.top - pad.bottom ? pad.top + (box.height - pad.top - pad.bottom - height) / 2 : pad.top;
      void flow.setViewport({ x: x - bounds.x * zoom, y: y - bounds.y * zoom, zoom }, { duration });
    },
    [flow, laid.nodes],
  );

  // Fit once the first graph arrives and whenever the scope changes.
  // biome-ignore lint/correctness/useExhaustiveDependencies: fit on these changes only
  useEffect(() => {
    if (!data) return;
    const frame = requestAnimationFrame(() => fit());
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

  if (graph.error && !data) {
    return (
      <Empty className="flex-1">
        <EmptyHeader>
          <EmptyTitle>{t.topoError()}</EmptyTitle>
          <EmptyDescription data-selectable>{errorMessage(graph.error)}</EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
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

        <div className="pointer-events-none absolute inset-x-3 top-3 flex items-start gap-2">
          <div className="pointer-events-auto flex items-center gap-2 rounded-xl border bg-background/90 p-1 shadow-xs/5 backdrop-blur">
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
          <form className="pointer-events-auto" onSubmit={find}>
            <InputGroup className="w-52 bg-background/90 backdrop-blur">
              <InputGroupAddon>
                <SearchIcon aria-hidden="true" />
              </InputGroupAddon>
              <InputGroupInput aria-label={t.topoFind()} onChange={(e) => setSearch(e.target.value)} placeholder={t.topoFind()} size="sm" value={search} />
            </InputGroup>
          </form>
          <div className="ms-auto flex items-center gap-1">
            {data && data.issues.length > 0 && (
              <Popover>
                <PopoverTrigger render={<Button className="pointer-events-auto bg-background/90" size="sm" variant="outline" />}>
                  <AlertTriangleIcon aria-hidden="true" className="text-warning" />
                  {t.topoIssues(String(data.issues.length))}
                </PopoverTrigger>
                <PopoverPopup align="end" className="w-96">
                  <ul className="flex flex-col gap-2 text-xs">
                    {data.issues.map((issue) => (
                      <li key={issue.message}>
                        <button
                          className="text-left hover:underline"
                          onClick={() => setFocus(`missing:${issue.missing}`)}
                          type="button"
                        >
                          {issue.message}
                        </button>
                      </li>
                    ))}
                  </ul>
                </PopoverPopup>
              </Popover>
            )}
            <div className="pointer-events-auto flex items-center rounded-lg border bg-background/90 shadow-xs/5 backdrop-blur">
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

        <Legend hidden={laid.hidden} />
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
    <div className="pointer-events-none absolute bottom-3 left-3 flex items-center gap-4 rounded-lg border bg-background/90 px-3 py-1.5 text-[0.6875rem] text-muted-foreground shadow-xs/5 backdrop-blur">
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
