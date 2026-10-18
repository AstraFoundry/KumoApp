import type { Node } from "@/shared/api/mygo";
import { displayName } from "@/shared/lib/format";
import { cn } from "@/shared/lib/utils";
import { DelayBadge } from "./delay-badge";
import { NodeFlag } from "./node-flag";

/**
 * One node of a group as a pressable tile. The selected one carries the
 * brand color; delay sits at the end.
 */
export function NodeTile({
  node,
  selected,
  disabled,
  onSelect,
  delay,
}: {
  node: Node;
  selected: boolean;
  disabled?: boolean;
  onSelect?: () => void;
  delay?: number;
}) {
  return (
    <button
      aria-pressed={selected}
      className={cn(
        "group/node relative flex h-9 min-w-0 items-center gap-2 rounded-lg border bg-card px-2.5 text-left text-[0.8125rem] outline-none",
        "transition-[background-color,border-color,transform] duration-150 ease-[var(--ease-out)] active:scale-[0.98]",
        "hover:border-foreground/12 hover:bg-accent/60 focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none",
        selected && "border-brand/40 bg-brand-soft hover:border-brand/50 hover:bg-brand-soft",
        disabled && "opacity-60",
      )}
      disabled={disabled}
      onClick={onSelect}
      title={node.name}
      type="button"
    >
      {selected && <span aria-hidden="true" className="absolute inset-y-2 start-0 w-0.5 rounded-full bg-brand" />}
      <NodeFlag country={node.country} group={node.group} />
      <span className={cn("min-w-0 flex-1 truncate", selected && "font-medium")}>{displayName(node.name)}</span>
      <DelayBadge delay={delay ?? node.delay} />
    </button>
  );
}
