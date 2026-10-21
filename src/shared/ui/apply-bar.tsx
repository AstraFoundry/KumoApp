import { t } from "@/shared/i18n";
import { cn } from "@/shared/lib/utils";
import { Button } from "@/shared/ui/button";

/**
 * The bar that slides up while a form has unapplied changes. It enters
 * from below with a strong ease-out and leaves the same way, so it reads
 * as one surface coming and going.
 */
export function ApplyBar({ dirty, applying, onApply, onDiscard, note }: { dirty: boolean; applying: boolean; onApply: () => void; onDiscard: () => void; note?: string }) {
  return (
    <div
      aria-hidden={!dirty}
      className={cn(
        "pointer-events-none sticky bottom-4 z-20 mx-auto flex w-full max-w-xl translate-y-[calc(100%+1rem)] items-center gap-3 rounded-xl border bg-popover/95 px-4 py-2.5 opacity-0 shadow-lg/5 backdrop-blur transition-[translate,opacity] duration-250 ease-[var(--ease-out)]",
        dirty && "pointer-events-auto translate-y-0 opacity-100",
      )}
      inert={!dirty || undefined}
    >
      <span className="min-w-0 flex-1 truncate text-sm">{note ?? t.applyUnsaved()}</span>
      <Button onClick={onDiscard} size="sm" variant="ghost">
        {t.applyDiscard()}
      </Button>
      <Button loading={applying} onClick={onApply} size="sm">
        {t.applyApply()}
      </Button>
    </div>
  );
}
