import { XIcon } from "lucide-react";
import type { Graph, TopologyNode } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { displayName, flag } from "@/shared/lib/format";
import { cn } from "@/shared/lib/utils";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { ScrollArea } from "@/shared/ui/scroll-area";
import { kindLabel, kindTone } from "./topology-node";

/**
 * Details of the focused node: what it is, what leads to it and where it
 * leads. A live select group's members can be picked right here.
 */
export function Inspector({
  graph,
  node,
  onClose,
  onFocus,
  onSelectMember,
  live,
}: {
  graph: Graph;
  node: TopologyNode;
  onClose: () => void;
  onFocus: (id: string) => void;
  onSelectMember?: (group: string, member: string) => void;
  live: boolean;
}) {
  const byID = new Map(graph.nodes.map((n) => [n.id, n]));
  const outgoing = graph.edges.filter((e) => e.source === node.id);
  const incoming = graph.edges.filter((e) => e.target === node.id);
  const selectable = live && node.kind === "group" && (node.type === "select" || node.type === "Selector");

  return (
    <aside className="flex w-72 shrink-0 flex-col border-s bg-background" aria-label={t.topoInspector()}>
      <div className="flex items-start gap-2 border-b px-4 py-3">
        <span aria-hidden="true" className={cn("mt-1.5 size-2 shrink-0 rounded-full", kindTone[node.kind])} />
        <div className="flex min-w-0 flex-1 flex-col gap-1">
          <span className="break-words font-medium text-sm" data-selectable>
            {node.kind === "proxy" && flag(node.country)} {node.label}
          </span>
          <div className="flex flex-wrap gap-1">
            <Badge size="sm" variant="secondary">
              {kindLabel(node.kind)}
            </Badge>
            {node.type && (
              <Badge size="sm" variant="outline">
                {node.type}
              </Badge>
            )}
          </div>
        </div>
        <Button aria-label={t.close()} className="-me-1.5 -mt-0.5" onClick={onClose} size="icon-xs" variant="ghost">
          <XIcon aria-hidden="true" />
        </Button>
      </div>
      <ScrollArea className="min-h-0 flex-1">
        <div className="flex flex-col gap-4 px-4 py-3">
          {node.payloads && node.payloads.length > 0 && (
            <Block title={t.topoMatches()}>
              <ul className="flex flex-col gap-1 font-mono text-[0.6875rem] text-muted-foreground">
                {node.payloads.map((p) => (
                  <li className="truncate" data-selectable key={p}>
                    {p}
                  </li>
                ))}
                {(node.count ?? 0) > node.payloads.length && <li>{t.topoAndMore(String((node.count ?? 0) - node.payloads.length))}</li>}
              </ul>
            </Block>
          )}
          {outgoing.length > 0 && (
            <Block title={node.kind === "group" || node.kind === "global" ? t.topoMembersTitle() : t.topoLeadsTo()}>
              <ul className="flex flex-col">
                {outgoing.map((e) => {
                  const target = byID.get(e.target);
                  if (!target) return null;
                  const active = e.active || node.selected === target.label;
                  return (
                    <li key={e.id}>
                      <button
                        className={cn(
                          "flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-xs outline-none hover:bg-accent focus-visible:bg-accent",
                          active && "font-medium text-brand-foreground",
                        )}
                        onClick={() => (selectable && onSelectMember ? onSelectMember(node.label, target.label) : onFocus(target.id))}
                        type="button"
                      >
                        <span aria-hidden="true" className={cn("size-1.5 shrink-0 rounded-full", kindTone[target.kind])} />
                        <span className="min-w-0 flex-1 truncate">{target.kind === "proxy" ? displayName(target.label) : target.label}</span>
                        {e.kind === "dynamic" && <span className="text-[0.625rem] text-muted-foreground">{t.topoDynamic()}</span>}
                        {active && <span className="text-[0.625rem]">{t.topoInUse()}</span>}
                      </button>
                    </li>
                  );
                })}
              </ul>
              {selectable && <p className="px-2 text-[0.6875rem] text-muted-foreground">{t.topoSelectHint()}</p>}
            </Block>
          )}
          {incoming.length > 0 && (
            <Block title={t.topoUsedBy()}>
              <ul className="flex flex-col">
                {incoming.map((e) => {
                  const source = byID.get(e.source);
                  if (!source) return null;
                  return (
                    <li key={e.id}>
                      <button
                        className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-xs outline-none hover:bg-accent focus-visible:bg-accent"
                        onClick={() => onFocus(source.id)}
                        type="button"
                      >
                        <span aria-hidden="true" className={cn("size-1.5 shrink-0 rounded-full", kindTone[source.kind])} />
                        <span className="min-w-0 flex-1 truncate">{source.kind === "rules" ? `${source.type} → ${source.label}` : source.label}</span>
                      </button>
                    </li>
                  );
                })}
              </ul>
            </Block>
          )}
        </div>
      </ScrollArea>
    </aside>
  );
}

function Block({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="flex flex-col gap-1.5">
      <h3 className="font-medium text-[0.6875rem] text-muted-foreground uppercase tracking-wide">{title}</h3>
      {children}
    </section>
  );
}
