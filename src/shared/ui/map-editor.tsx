import { PlusIcon, XIcon } from "lucide-react";
import { useState } from "react";
import { t } from "@/shared/i18n";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";

/**
 * Edits a map whose values are one or more strings, such as DNS policies
 * and hosts. Several values are written comma-separated.
 */
export function MapEditor({
  value,
  onChange,
  keyPlaceholder,
  valuePlaceholder,
}: {
  value: Record<string, string[]>;
  onChange: (next: Record<string, string[]>) => void;
  keyPlaceholder: string;
  valuePlaceholder: string;
}) {
  const [key, setKey] = useState("");
  const [entry, setEntry] = useState("");
  const keys = Object.keys(value).sort();

  function add() {
    const values = entry
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    if (!key.trim() || values.length === 0) return;
    onChange({ ...value, [key.trim()]: values });
    setKey("");
    setEntry("");
  }

  return (
    <div className="flex w-full flex-col gap-1.5">
      {keys.length > 0 && (
        <div className="flex flex-col divide-y rounded-lg border">
          {keys.map((k) => (
            <div className="flex items-center gap-3 px-2.5 py-1.5 font-mono text-xs" key={k}>
              <span className="w-2/5 truncate" data-selectable>
                {k}
              </span>
              <span className="min-w-0 flex-1 truncate text-muted-foreground" data-selectable>
                {(value[k] ?? []).join(", ")}
              </span>
              <Button
                aria-label={t.listRemove(k)}
                className="text-muted-foreground"
                onClick={() => {
                  const next = { ...value };
                  delete next[k];
                  onChange(next);
                }}
                size="icon-xs"
                variant="ghost"
              >
                <XIcon aria-hidden="true" />
              </Button>
            </div>
          ))}
        </div>
      )}
      <form
        className="flex gap-1.5"
        onSubmit={(e) => {
          e.preventDefault();
          add();
        }}
      >
        <Input className="w-2/5 font-mono" onChange={(e) => setKey(e.target.value)} placeholder={keyPlaceholder} size="sm" value={key} />
        <Input className="flex-1 font-mono" onChange={(e) => setEntry(e.target.value)} placeholder={valuePlaceholder} size="sm" value={entry} />
        <Button aria-label={t.listAdd()} disabled={!key.trim() || !entry.trim()} size="icon-sm" type="submit" variant="outline">
          <PlusIcon aria-hidden="true" />
        </Button>
      </form>
    </div>
  );
}
