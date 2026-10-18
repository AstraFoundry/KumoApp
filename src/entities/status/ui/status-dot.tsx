import type { RunState } from "@/shared/api/mygo";
import { cn } from "@/shared/lib/utils";

const tone: Record<RunState, string> = {
  running: "bg-success",
  starting: "bg-warning animate-pulse",
  stopping: "bg-warning animate-pulse",
  failed: "bg-destructive",
  stopped: "bg-muted-foreground/40",
};

/** A dot that says the core's state by color; pair it with a label. */
export function StatusDot({ state, className }: { state: RunState | undefined; className?: string }) {
  return (
    <span
      aria-hidden="true"
      className={cn("relative inline-flex size-2 shrink-0 rounded-full", tone[state ?? "stopped"], className)}
    >
      {state === "running" && <span className="absolute inset-0 animate-ping rounded-full bg-success/40 [animation-duration:2.4s]" />}
    </span>
  );
}
