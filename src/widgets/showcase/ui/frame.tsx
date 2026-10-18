import type { ReactNode } from "react";
import { cn } from "@/shared/lib/utils";

/** A small window of another app, drawn as an illustration. */
export function Frame({ title, children, dark, className }: { title?: string; children: ReactNode; dark?: boolean; className?: string }) {
  return (
    <div
      className={cn(
        "w-[17.5rem] overflow-hidden rounded-[0.9rem] border shadow-[0_24px_48px_-16px_rgb(0_0_0/0.35),0_2px_6px_rgb(0_0_0/0.08)]",
        dark ? "border-white/10 bg-neutral-900/95 text-neutral-100" : "border-black/5 bg-white/95 text-neutral-800",
        className,
      )}
    >
      <div className={cn("flex h-7 items-center gap-1.5 border-b px-3", dark ? "border-white/5" : "border-black/5")}>
        {["#ff5f57", "#febc2e", "#28c840"].map((c) => (
          <span className="size-2 rounded-full" key={c} style={{ background: dark ? c : `${c}cc` }} />
        ))}
        {title && <span className={cn("ms-2 truncate text-[0.625rem]", dark ? "text-neutral-400" : "text-neutral-500")}>{title}</span>}
      </div>
      {children}
    </div>
  );
}

/** A floating pill under the window, like a menu bar status. */
export function Pill({ children }: { children: ReactNode }) {
  return (
    <div className="-mt-4 relative flex items-center gap-2 rounded-full border border-white/40 bg-white/80 px-3.5 py-1.5 text-[0.6875rem] text-neutral-800 shadow-[0_12px_32px_-12px_rgb(0_0_0/0.4)] backdrop-blur-md">
      {children}
    </div>
  );
}

export function Toggle({ on }: { on: boolean }) {
  return (
    <span className={cn("flex h-3.5 w-6 items-center rounded-full p-px transition-colors duration-300", on ? "bg-neutral-800" : "bg-neutral-200")}>
      <span className={cn("size-3 rounded-full bg-white shadow-sm transition-transform duration-300 ease-[var(--ease-out)]", on && "translate-x-2.5")} />
    </span>
  );
}
