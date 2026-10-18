import type { ReactNode } from "react";
import { cn } from "@/shared/lib/utils";

/**
 * The scrolling area under a page header, at a readable width. It shares
 * the header's 20px gutter and stays left-aligned under the title, so the
 * two never drift apart on wide windows.
 */
export function PageBody({ children, className, wide }: { children: ReactNode; className?: string; wide?: boolean }) {
  return (
    <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain">
      <div className={cn("flex w-full flex-col gap-6 px-5 py-5", wide ? "max-w-none" : "max-w-[60rem]", className)}>{children}</div>
    </div>
  );
}

/** A titled group of settings or content on a page. */
export function Section({ title, description, action, children }: { title: string; description?: string; action?: ReactNode; children: ReactNode }) {
  return (
    <section className="flex flex-col gap-3">
      <div className="flex items-end gap-3">
        <div className="flex min-w-0 flex-col gap-0.5">
          <h2 className="font-semibold text-sm">{title}</h2>
          {description && <p className="text-muted-foreground text-sm">{description}</p>}
        </div>
        {action && <div className="ms-auto flex shrink-0 items-center gap-2">{action}</div>}
      </div>
      {children}
    </section>
  );
}
