import { PlusIcon, XIcon } from "lucide-react";
import { useState } from "react";
import { t } from "@/shared/i18n";
import { cn } from "@/shared/lib/utils";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";

/**
 * Edits a list of strings: each entry is a removable row, and a field at
 * the end adds more. Pasting several lines adds each as its own entry.
 */
export function ListEditor({
  value,
  onChange,
  placeholder,
  mono = true,
  validate,
}: {
  value: string[];
  onChange: (next: string[]) => void;
  placeholder?: string;
  mono?: boolean;
  validate?: (entry: string) => boolean;
}) {
  const [draft, setDraft] = useState("");
  const invalid = draft.trim() !== "" && validate !== undefined && !validate(draft.trim());

  function add(text: string) {
    const entries = text
      .split(/[\n,]/)
      .map((s) => s.trim())
      .filter((s) => s !== "" && !value.includes(s));
    if (entries.length === 0) return;
    onChange([...value, ...entries]);
    setDraft("");
  }

  return (
    <div className="flex w-full flex-col gap-1.5">
      {value.length > 0 && (
        <ul className="flex flex-wrap gap-1.5">
          {value.map((entry, index) => (
            <li
              className={cn(
                "flex h-7 items-center gap-1 rounded-md border bg-muted/40 ps-2 pe-0.5 text-xs",
                mono && "font-mono",
              )}
              key={`${entry}-${index}`}
            >
              <span className="max-w-72 truncate" data-selectable>
                {entry}
              </span>
              <Button
                aria-label={t.listRemove(entry)}
                className="text-muted-foreground"
                onClick={() => onChange(value.filter((_, i) => i !== index))}
                size="icon-xs"
                variant="ghost"
              >
                <XIcon aria-hidden="true" />
              </Button>
            </li>
          ))}
        </ul>
      )}
      <form
        className="flex gap-1.5"
        onSubmit={(e) => {
          e.preventDefault();
          if (!invalid) add(draft);
        }}
      >
        <Input
          aria-invalid={invalid || undefined}
          className={cn(mono && "font-mono")}
          onChange={(e) => setDraft(e.target.value)}
          onPaste={(e) => {
            const text = e.clipboardData.getData("text");
            if (text.includes("\n")) {
              e.preventDefault();
              add(text);
            }
          }}
          placeholder={placeholder ?? t.listAddPlaceholder()}
          size="sm"
          value={draft}
        />
        <Button aria-label={t.listAdd()} disabled={draft.trim() === "" || invalid} size="icon-sm" type="submit" variant="outline">
          <PlusIcon aria-hidden="true" />
        </Button>
      </form>
    </div>
  );
}

/** Edits a list of ports. */
export function PortListEditor({ value, onChange }: { value: number[]; onChange: (next: number[]) => void }) {
  return (
    <ListEditor
      onChange={(next) => onChange(next.map(Number).filter((n) => n > 0 && n < 65536))}
      placeholder="443"
      validate={(s) => /^\d+$/.test(s) && Number(s) > 0 && Number(s) < 65536}
      value={value.map(String)}
    />
  );
}
