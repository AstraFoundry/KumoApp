import { FileUpIcon, LinkIcon, TextIcon } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useStatus } from "@/entities/status";
import { Profiles } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { Button } from "@/shared/ui/button";
import { Checkbox } from "@/shared/ui/checkbox";
import { CodeEditor } from "@/shared/ui/code-editor";
import { Dialog, DialogDescription, DialogFooter, DialogHeader, DialogPanel, DialogPopup, DialogTitle } from "@/shared/ui/dialog";
import { Field, FieldLabel } from "@/shared/ui/field";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Switch } from "@/shared/ui/switch";
import { Tabs, TabsList, TabsPanel, TabsTab } from "@/shared/ui/tabs";
import { toastManager } from "@/shared/ui/toast";

type Source = "url" | "file" | "text";

/**
 * Adds a profile from a subscription URL, a YAML file or pasted text. The
 * profile becomes current unless the user unticks it.
 */
export function ImportProfileDialog({
  open,
  onOpenChange,
  initialURL,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  initialURL?: string;
}) {
  const status = useStatus();
  const [source, setSource] = useState<Source>("url");
  const [url, setURL] = useState("");
  const [name, setName] = useState("");
  const [yaml, setYAML] = useState("");
  const [useProxy, setUseProxy] = useState(false);
  const [activate, setActivate] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) return;
    setURL(initialURL ?? "");
    setSource("url");
    setError("");
  }, [open, initialURL]);

  async function submit(e?: FormEvent) {
    e?.preventDefault();
    setBusy(true);
    setError("");
    try {
      if (source === "file") {
        const summary = await Profiles.chooseFile(activate);
        if (!summary) return;
        done(summary.name);
        return;
      }
      const summary = await Profiles.import(
        source === "url" ? { url: url.trim(), name: name.trim(), useProxy, activate } : { yaml, name: name.trim(), activate },
      );
      done(summary.name);
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  function done(profileName: string) {
    toastManager.add({ title: t.importDone(profileName), type: "success" });
    setName("");
    setYAML("");
    onOpenChange(false);
  }

  const canSubmit = source === "file" || (source === "url" ? /^https?:\/\/\S+$/.test(url.trim()) : yaml.trim() !== "");

  return (
    <Dialog onOpenChange={onOpenChange} open={open}>
      <DialogPopup className="max-w-xl">
        <form onSubmit={submit}>
          <DialogHeader>
            <DialogTitle>{t.importTitle()}</DialogTitle>
            <DialogDescription>{t.importDescription()}</DialogDescription>
          </DialogHeader>
          <DialogPanel className="flex flex-col gap-4">
            <Tabs onValueChange={(v) => setSource(v as Source)} value={source}>
              <TabsList className="w-full">
                <TabsTab className="flex-1" value="url">
                  <LinkIcon aria-hidden="true" />
                  {t.importFromURL()}
                </TabsTab>
                <TabsTab className="flex-1" value="file">
                  <FileUpIcon aria-hidden="true" />
                  {t.importFromFile()}
                </TabsTab>
                <TabsTab className="flex-1" value="text">
                  <TextIcon aria-hidden="true" />
                  {t.importFromText()}
                </TabsTab>
              </TabsList>
              <TabsPanel className="flex flex-col gap-3 pt-4" value="url">
                <Field>
                  <FieldLabel>{t.importURL()}</FieldLabel>
                  <Input autoFocus className="font-mono" onChange={(e) => setURL(e.target.value)} placeholder="https://" value={url} />
                </Field>
                <Field>
                  <FieldLabel>{t.importName()}</FieldLabel>
                  <Input onChange={(e) => setName(e.target.value)} placeholder={t.importNamePlaceholder()} value={name} />
                </Field>
                <Label className="flex items-center justify-between gap-3 font-normal">
                  <span className="flex flex-col gap-0.5">
                    <span>{t.importUseProxy()}</span>
                    <span className="text-muted-foreground text-xs">
                      {status?.state === "running" ? t.importUseProxyHint() : t.importUseProxyStopped()}
                    </span>
                  </span>
                  <Switch checked={useProxy} disabled={status?.state !== "running"} onCheckedChange={setUseProxy} />
                </Label>
              </TabsPanel>
              <TabsPanel className="pt-4" value="file">
                <p className="text-muted-foreground text-sm">{t.importFileHint()}</p>
              </TabsPanel>
              <TabsPanel className="flex flex-col gap-3 pt-4" value="text">
                <Field>
                  <FieldLabel>{t.importName()}</FieldLabel>
                  <Input onChange={(e) => setName(e.target.value)} placeholder={t.importNamePlaceholder()} value={name} />
                </Field>
                <CodeEditor label={t.importFromText()} minRows={9} onChange={setYAML} value={yaml} />
              </TabsPanel>
            </Tabs>
            <Label className="flex items-center gap-2 font-normal">
              <Checkbox checked={activate} onCheckedChange={(v) => setActivate(v === true)} />
              {t.importActivate()}
            </Label>
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
            <Button disabled={!canSubmit} loading={busy} type="submit">
              {source === "file" ? t.importChooseFile() : t.importSubmit()}
            </Button>
          </DialogFooter>
        </form>
      </DialogPopup>
    </Dialog>
  );
}
