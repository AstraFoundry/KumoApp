import { GlobeIcon, LayersIcon } from "lucide-react";
import { flag } from "@/shared/lib/format";

/** The flag of the node's region, a globe when unknown, layers for groups. */
export function NodeFlag({ country, group }: { country?: string; group?: boolean }) {
  if (group) return <LayersIcon aria-hidden="true" className="size-3.5 shrink-0 text-muted-foreground" />;
  const emoji = flag(country);
  if (!emoji) return <GlobeIcon aria-hidden="true" className="size-3.5 shrink-0 text-muted-foreground/70" />;
  return (
    <span aria-hidden="true" className="w-3.5 shrink-0 text-center text-[0.8125rem] leading-none">
      {emoji}
    </span>
  );
}
