import { t } from "@/shared/i18n";
import { delayGrade } from "@/shared/lib/format";
import { cn } from "@/shared/lib/utils";

const tone = {
  unknown: "text-muted-foreground/60",
  fast: "text-success-foreground",
  slow: "text-warning-foreground",
  timeout: "text-destructive-foreground",
};

/** A node's delay: green under 300 ms, amber above, red on timeout. */
export function DelayBadge({ delay, className }: { delay: number | undefined; className?: string }) {
  const grade = delayGrade(delay);
  const text = grade === "unknown" ? "—" : grade === "timeout" ? t.delayTimeout() : `${delay}`;
  return (
    <span className={cn("shrink-0 font-medium text-[0.6875rem] tabular-nums", tone[grade], className)}>
      {text}
      {grade !== "unknown" && grade !== "timeout" && <span className="ms-px opacity-60">ms</span>}
    </span>
  );
}
