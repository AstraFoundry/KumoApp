import { Handle, type NodeProps, Position } from "@xyflow/react";
import { ChevronDownIcon } from "lucide-react";
import type { TopologyKind } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { displayName, flag } from "@/shared/lib/format";
import { cn } from "@/shared/lib/utils";
import { COLUMN_WIDTH, NODE_HEIGHT, type TopologyNodeData } from "../lib/layout";

/** Each kind of node has one accent, used for its dot only. */
export const kindTone: Record<TopologyKind, string> = {
  entry: "bg-brand",
  rules: "bg-violet-500",
  ruleSet: "bg-violet-500",
  ruleProvider: "bg-violet-500",
  final: "bg-violet-500",
  group: "bg-sky-500",
  global: "bg-sky-500",
  proxy: "bg-emerald-500",
  provider: "bg-amber-500",
  builtin: "bg-muted-foreground/50",
  missing: "bg-destructive",
};

export function kindLabel(kind: TopologyKind): string {
  switch (kind) {
    case "entry":
      return t.topoKindEntry();
    case "rules":
      return t.topoKindRules();
    case "ruleSet":
    case "ruleProvider":
      return t.topoKindRuleSet();
    case "final":
      return t.topoKindFinal();
    case "group":
    case "global":
      return t.topoKindGroup();
    case "proxy":
      return t.topoKindProxy();
    case "provider":
      return t.topoKindProvider();
    case "builtin":
      return t.topoKindBuiltin();
    case "missing":
      return t.topoKindMissing();
  }
}

function modeTitle(mode: string) {
  if (mode === "global") return t.modeGlobal();
  if (mode === "direct") return t.modeDirect();
  return t.modeRule();
}

/** What a node says under its label, as briefly as possible. */
function detail(data: TopologyNodeData): string {
  const n = data.node;
  switch (n.kind) {
    case "entry":
      return t.topoEntryDetail();
    case "rules":
      return t.topoRulesDetail(n.count ?? 0);
    case "ruleSet":
      return n.type ? `RULE-SET · ${n.type}` : "RULE-SET";
    case "final":
      return t.topoFinalDetail();
    case "group":
    case "global":
      return [n.type, n.count ? t.topoMembers(String(n.count)) : ""].filter(Boolean).join(" · ");
    case "missing":
      return t.topoMissingDetail();
    default:
      return n.type ?? kindLabel(n.kind);
  }
}

export function TopologyNodeView({ data, selected }: NodeProps) {
  const d = data as TopologyNodeData;
  const n = d.node;
  const label = n.kind === "entry" ? modeTitle(n.label) : n.kind === "rules" || n.kind === "final" ? `→ ${n.label}` : n.kind === "proxy" ? displayName(n.label) : n.label;
  const emoji = n.kind === "proxy" ? flag(n.country) : "";
  return (
    <div
      className={cn(
        "flex items-center gap-2.5 rounded-xl border bg-card px-3 text-left shadow-xs/5",
        "transition-[opacity,border-color,box-shadow,background-color] duration-150 ease-[var(--ease-out)]",
        d.dimmed && "opacity-30",
        d.highlighted && "border-foreground/20",
        (d.focused || selected) && "border-brand/60 shadow-[0_0_0_3px_var(--brand-soft)]",
        n.kind === "missing" && "border-destructive/40 border-dashed bg-destructive/4",
        n.kind === "entry" && "border-brand/30 bg-brand-soft",
      )}
      style={{ width: COLUMN_WIDTH, height: NODE_HEIGHT }}
    >
      <Handle className="!invisible" isConnectable={false} position={Position.Left} type="target" />
      <span aria-hidden="true" className={cn("size-1.5 shrink-0 rounded-full", kindTone[n.kind])} />
      {emoji && (
        <span aria-hidden="true" className="-me-1 text-xs leading-none">
          {emoji}
        </span>
      )}
      <span className="flex min-w-0 flex-1 items-baseline gap-2">
        <span className={cn("min-w-0 shrink truncate text-[0.8125rem]", (n.kind === "group" || n.kind === "entry") && "font-medium")}>{label}</span>
        <span className="ms-auto min-w-0 shrink-[2] truncate text-[0.6875rem] text-muted-foreground">{detail(d)}</span>
      </span>
      <Handle className="!invisible" isConnectable={false} position={Position.Right} type="source" />
    </div>
  );
}

/** Stands for the nodes a collapsed column hides; pressing it expands. */
export function MoreNodeView({ data }: NodeProps) {
  const d = data as TopologyNodeData;
  return (
    <div
      className="flex items-center justify-center gap-1.5 rounded-xl border border-dashed text-muted-foreground text-xs transition-colors hover:border-foreground/20 hover:text-foreground"
      style={{ width: COLUMN_WIDTH, height: NODE_HEIGHT }}
    >
      {t.topoMore(String(d.more ?? 0))}
      <ChevronDownIcon aria-hidden="true" className="size-3.5" />
    </div>
  );
}
