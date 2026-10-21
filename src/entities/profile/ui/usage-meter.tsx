import type { Usage } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { formatBytes, formatDate } from "@/shared/lib/format";
import { cn } from "@/shared/lib/utils";

/** A subscription's traffic: used of total, and when it expires. */
export function UsageMeter({ usage, className }: { usage: Usage; className?: string }) {
  const used = usage.upload + usage.download;
  const ratio = usage.total > 0 ? Math.min(1, used / usage.total) : 0;
  return (
    <div className={cn("flex flex-col gap-1.5", className)}>
      <div className="h-1.5 overflow-hidden rounded-full bg-muted" role="meter" aria-valuemin={0} aria-valuemax={usage.total} aria-valuenow={used} aria-label={t.profileUsage()}>
        <div
          className={cn("h-full rounded-full bg-brand transition-[width] duration-500 ease-[var(--ease-out)]", ratio > 0.9 && "bg-warning")}
          style={{ width: `${ratio * 100}%` }}
        />
      </div>
      <div className="flex justify-between text-muted-foreground text-xs tabular-nums">
        <span>
          {formatBytes(used)} / {usage.total > 0 ? formatBytes(usage.total) : "∞"}
        </span>
        {usage.expire ? <span>{t.profileExpires(formatDate(usage.expire))}</span> : null}
      </div>
    </div>
  );
}
