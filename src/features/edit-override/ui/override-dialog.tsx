import { useEffect, useState } from "react";
import { type Format, Overrides } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { Button } from "@/shared/ui/button";
import { CodeEditor } from "@/shared/ui/code-editor";
import { Dialog, DialogDescription, DialogFooter, DialogHeader, DialogPanel, DialogPopup, DialogTitle } from "@/shared/ui/dialog";
import { Field, FieldLabel } from "@/shared/ui/field";
import { Input } from "@/shared/ui/input";
import { toastManager } from "@/shared/ui/toast";
import { ToggleGroup, ToggleGroupItem } from "@/shared/ui/toggle-group";

const yamlExample = `# Add rules before the profile's own
+rules:
  - DOMAIN-SUFFIX,example.com,DIRECT
`;

const jsExample = `// Change the configuration and return it.
function main(config) {
  config.rules = ["DOMAIN-SUFFIX,example.com,DIRECT", ...(config.rules ?? [])];
  return config;
}
`;

/** Creates or edits an override; "new" opens an empty one. */
export function OverrideDialog({ id, onOpenChange }: { id: string | undefined; onOpenChange: (open: boolean) => void }) {
  const [name, setName] = useState("");
  const [format, setFormat] = useState<Format>("yaml");
  const [content, setContent] = useState("");
  const [url, setURL] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const creating = id === "new";

  useEffect(() => {
    setError("");
    if (!id) return;
    if (creating) {
      setName("");
      setFormat("yaml");
      setContent(yamlExample);
      setURL("");
      return;
    }
    void Overrides.get(id).then((o) => {
      setName(o.name);
      setFormat(o.format);
      setContent(o.content);
      setURL(o.remoteURL ?? "");
    });
  }, [id, creating]);

  async function save() {
    setBusy(true);
    setError("");
    try {
      await Overrides.save({ id: creating ? undefined : id, name, format, content, url: url.trim() || undefined });
      toastManager.add({ title: t.overrideSaved(name || t.overrideUntitled()), type: "success" });
      onOpenChange(false);
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <Dialog onOpenChange={onOpenChange} open={id !== undefined}>
      <DialogPopup className="max-w-3xl">
        <DialogHeader>
          <DialogTitle>{creating ? t.overrideNew() : t.overrideEdit()}</DialogTitle>
          <DialogDescription>{t.overrideDialogHint()}</DialogDescription>
        </DialogHeader>
        <DialogPanel className="flex flex-col gap-3">
          <div className="grid grid-cols-[1fr_auto] items-end gap-3">
            <Field>
              <FieldLabel>{t.importName()}</FieldLabel>
              <Input onChange={(e) => setName(e.target.value)} placeholder={t.overrideUntitled()} value={name} />
            </Field>
            <ToggleGroup
              onValueChange={(v) => {
                const next = v[0] as Format | undefined;
                if (!next) return;
                if (creating && (content === yamlExample || content === jsExample)) setContent(next === "yaml" ? yamlExample : jsExample);
                setFormat(next);
              }}
              size="sm"
              value={[format]}
              variant="outline"
            >
              <ToggleGroupItem value="yaml">YAML</ToggleGroupItem>
              <ToggleGroupItem value="javascript">JavaScript</ToggleGroupItem>
            </ToggleGroup>
          </div>
          <Field>
            <FieldLabel>{t.overrideURL()}</FieldLabel>
            <Input className="font-mono" onChange={(e) => setURL(e.target.value)} placeholder={t.overrideURLPlaceholder()} value={url} />
          </Field>
          <CodeEditor className="h-72" label={t.overrideContent()} onChange={setContent} value={content} />
          {error && (
            <p className="text-destructive-foreground text-sm" data-selectable role="alert">
              {error}
            </p>
          )}
        </DialogPanel>
        <DialogFooter>
          <Button onClick={() => onOpenChange(false)} variant="ghost">
            {t.cancel()}
          </Button>
          <Button loading={busy} onClick={save}>
            {t.save()}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
