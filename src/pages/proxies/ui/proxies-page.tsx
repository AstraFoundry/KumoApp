import { ChevronRightIcon, SearchIcon } from "lucide-react";
import { useMemo, useState } from "react";
import { NodeTile } from "@/entities/proxy";
import { useStatus } from "@/entities/status";
import { useSelectNode } from "@/features/select-proxy";
import { TestDelayButton, useDelays } from "@/features/test-delay";
import { type Group, Proxies } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { displayName } from "@/shared/lib/format";
import { useResource } from "@/shared/lib/use-resource";
import { cn } from "@/shared/lib/utils";
import { Alert, AlertDescription } from "@/shared/ui/alert";
import { Badge } from "@/shared/ui/badge";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/shared/ui/input-group";
import { PageBody } from "@/shared/ui/page-body";
import { PageHeader } from "@/widgets/page-header";

const typeLabel: Record<string, () => string> = {
  Selector: () => t.groupTypeSelect(),
  URLTest: () => t.groupTypeAuto(),
  Fallback: () => t.groupTypeFallback(),
  LoadBalance: () => t.groupTypeBalance(),
  Relay: () => t.groupTypeRelay(),
};

function label(type: string) {
  const normalized = { select: "Selector", "url-test": "URLTest", fallback: "Fallback", "load-balance": "LoadBalance", relay: "Relay" }[type] ?? type;
  return typeLabel[normalized]?.() ?? type;
}

export function ProxiesPage() {
  const status = useStatus();
  const running = status?.state === "running";
  const groups = useResource(() => Proxies.groups(), [status?.state, status?.profile.id, status?.mode], { intervalMs: running ? 4000 : undefined });
  const { select, selected } = useSelectNode(() => void groups.reload());
  const { delays, test, isTesting } = useDelays();
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState<Record<string, boolean>>({});

  const visible = useMemo(() => {
    const q = query.trim().toLowerCase();
    return (groups.data ?? [])
      .filter((g) => !g.hidden && (g.name !== "GLOBAL" || status?.mode === "global"))
      .map((g) => (q ? { ...g, nodes: g.nodes.filter((n) => n.name.toLowerCase().includes(q)) } : g))
      .filter((g) => !q || g.nodes.length > 0);
  }, [groups.data, query, status?.mode]);

  const isOpen = (g: Group, index: number) => query.trim() !== "" || (open[g.name] ?? index === 0);

  return (
    <>
      <PageHeader
        actions={
          <InputGroup className="w-56">
            <InputGroupAddon>
              <SearchIcon aria-hidden="true" />
            </InputGroupAddon>
            <InputGroupInput aria-label={t.proxiesSearch()} onChange={(e) => setQuery(e.target.value)} placeholder={t.proxiesSearch()} size="sm" type="search" value={query} />
          </InputGroup>
        }
        id="proxies"
      />
      <PageBody>
        {status && !running && (
          <Alert>
            <AlertDescription>{t.proxiesPreview()}</AlertDescription>
          </Alert>
        )}
        {groups.data && visible.length === 0 && (
          <Empty>
            <EmptyHeader>
              <EmptyTitle>{query ? t.proxiesNoMatch() : t.proxiesEmpty()}</EmptyTitle>
              <EmptyDescription>{query ? t.proxiesNoMatchHint() : t.proxiesEmptyHint()}</EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
        <div className="flex flex-col gap-3">
          {visible.map((g, index) => {
            const expanded = isOpen(g, index);
            const now = selected(g.name, g.now);
            const selectable = g.live && (g.type === "Selector" || g.type === "select");
            return (
              <section className="overflow-hidden rounded-xl border bg-card shadow-xs/5" key={g.name}>
                <div className="flex items-center gap-2 pe-2">
                  <button
                    aria-expanded={expanded}
                    className="flex min-w-0 flex-1 items-center gap-2.5 py-3 ps-4 text-left outline-none focus-visible:bg-accent/50"
                    onClick={() => setOpen((o) => ({ ...o, [g.name]: !expanded }))}
                    type="button"
                  >
                    <ChevronRightIcon
                      aria-hidden="true"
                      className={cn("size-4 shrink-0 text-muted-foreground transition-transform duration-200 ease-[var(--ease-out)]", expanded && "rotate-90")}
                    />
                    <span className="truncate font-medium text-sm">{g.name}</span>
                    <Badge size="sm" variant="outline">
                      {label(g.type)}
                    </Badge>
                    <span className="truncate text-muted-foreground text-xs">
                      {now ? `→ ${displayName(now)}` : ""} · {t.proxiesCount(String(g.nodes.length))}
                    </span>
                  </button>
                  <TestDelayButton disabled={!running} onTest={() => test(g.name)} testing={isTesting(g.name)} />
                </div>
                {expanded && (
                  <div className="grid grid-cols-[repeat(auto-fill,minmax(12rem,1fr))] gap-1.5 border-t bg-muted/20 p-3">
                    {g.nodes.map((node) => (
                      <NodeTile
                        delay={delays[node.name]}
                        disabled={!selectable}
                        key={node.name}
                        node={node}
                        onSelect={() => select(g.name, node.name)}
                        selected={g.live && node.name === now}
                      />
                    ))}
                  </div>
                )}
              </section>
            );
          })}
        </div>
      </PageBody>
    </>
  );
}
