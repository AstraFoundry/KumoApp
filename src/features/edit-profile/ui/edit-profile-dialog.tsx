import { useEffect, useState } from "react";
import { Profiles, type Summary } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { Button } from "@/shared/ui/button";
import { CodeEditor } from "@/shared/ui/code-editor";
import { Dialog, DialogFooter, DialogHeader, DialogPanel, DialogPopup, DialogTitle } from "@/shared/ui/dialog";
import { Field, FieldLabel } from "@/shared/ui/field";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Switch } from "@/shared/ui/switch";
import { toastManager } from "@/shared/ui/toast";

/** Edits a profile's name, subscription, update behavior and YAML. */
export function EditProfileDialog({ profile, onOpenChange }: { profile: Summary | undefined; onOpenChange: (open: boolean) => void }) {
  const [name, setName] = useState("");
  const [url, setURL] = useState("");
  const [autoUpdate, setAutoUpdate] = useState(true);
  const [useProxy, setUseProxy] = useState(false);
  const [yaml, setYAML] = useState("");
  const [original, setOriginal] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!profile) return;
    setName(profile.name);
    setURL(profile.remoteURL ?? "");
    setAutoUpdate(profile.autoUpdate);
    setUseProxy(profile.useProxy);
    setError("");
    void Profiles.content(profile.id).then((text) => {
      setYAML(text);
      setOriginal(text);
    });
  }, [profile]);

  async function save() {
    if (!profile) return;
    setBusy(true);
    setError("");
    try {
      await Profiles.update({ id: profile.id, name, remoteURL: url, autoUpdate, useProxy, yaml: yaml === original ? undefined : yaml });
      toastManager.add({ title: t.editSaved(name), type: "success" });
      onOpenChange(false);
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <Dialog onOpenChange={onOpenChange} open={profile !== undefined}>
      <DialogPopup className="max-w-3xl">
        <DialogHeader>
          <DialogTitle>{t.editTitle()}</DialogTitle>
        </DialogHeader>
        <DialogPanel className="flex flex-col gap-4">
          <div className="grid grid-cols-2 gap-3">
            <Field>
              <FieldLabel>{t.importName()}</FieldLabel>
              <Input onChange={(e) => setName(e.target.value)} value={name} />
            </Field>
            <Field>
              <FieldLabel>{t.importURL()}</FieldLabel>
              <Input className="font-mono" onChange={(e) => setURL(e.target.value)} placeholder={t.editURLPlaceholder()} value={url} />
            </Field>
          </div>
          {url.trim() !== "" && (
            <div className="flex gap-6">
              <Label className="flex items-center gap-2 font-normal">
                <Switch checked={autoUpdate} onCheckedChange={setAutoUpdate} />
                {t.editAutoUpdate()}
              </Label>
              <Label className="flex items-center gap-2 font-normal">
                <Switch checked={useProxy} onCheckedChange={setUseProxy} />
                {t.importUseProxy()}
              </Label>
            </div>
          )}
          <CodeEditor className="h-80" label="YAML" onChange={setYAML} value={yaml} />
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
