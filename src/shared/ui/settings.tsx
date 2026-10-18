import type { ReactNode } from "react";
import { cn } from "@/shared/lib/utils";

/**
 * A grouped list of settings, as macOS forms show them: one card, rows
 * divided by hairlines, the label on the left and the control on the right.
 */
export function SettingsGroup({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div className={cn("flex flex-col divide-y overflow-hidden rounded-xl border bg-card shadow-xs/5", className)} data-slot="settings-group">
      {children}
    </div>
  );
}

/** One setting: label, an optional description, and its control. */
export function SettingRow({
  label,
  description,
  children,
  htmlFor,
  stacked,
}: {
  label: ReactNode;
  description?: ReactNode;
  children?: ReactNode;
  htmlFor?: string;
  /** Puts the control under the label, for wide controls such as lists. */
  stacked?: boolean;
}) {
  return (
    <div className={cn("flex gap-4 px-4 py-3", stacked ? "flex-col gap-2.5" : "items-center")}>
      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <label className="font-medium text-sm" htmlFor={htmlFor}>
          {label}
        </label>
        {description && <p className="text-muted-foreground text-xs leading-relaxed">{description}</p>}
      </div>
      {children && <div className={cn("flex shrink-0 items-center gap-2", stacked && "w-full")}>{children}</div>}
    </div>
  );
}
