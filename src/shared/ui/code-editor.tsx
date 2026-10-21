import { cn } from "@/shared/lib/utils";

/** A plain monospace editor for YAML, JavaScript and PAC scripts. */
export function CodeEditor({
  value,
  onChange,
  readOnly,
  className,
  label,
  minRows = 14,
}: {
  value: string;
  onChange?: (next: string) => void;
  readOnly?: boolean;
  className?: string;
  label: string;
  minRows?: number;
}) {
  return (
    <textarea
      aria-label={label}
      className={cn(
        "w-full resize-y rounded-lg border border-input bg-muted/30 px-3 py-2.5 font-mono text-[0.75rem] leading-relaxed text-foreground outline-none [tab-size:2] focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/24",
        readOnly && "bg-muted/50",
        className,
      )}
      onChange={(e) => onChange?.(e.target.value)}
      onKeyDown={(e) => {
        if (e.key !== "Tab" || readOnly) return;
        e.preventDefault();
        const el = e.currentTarget;
        const { selectionStart, selectionEnd } = el;
        const next = `${value.slice(0, selectionStart)}  ${value.slice(selectionEnd)}`;
        onChange?.(next);
        requestAnimationFrame(() => el.setSelectionRange(selectionStart + 2, selectionStart + 2));
      }}
      readOnly={readOnly}
      rows={minRows}
      spellCheck={false}
      value={value}
    />
  );
}
