import { CopyIcon, DownloadIcon, FileCodeIcon, FolderOpenIcon } from "lucide-react";
import { useState } from "react";
import { useStatus } from "@/entities/status";
import { useCoreSettings } from "@/features/edit-core-settings";
import { App, Settings } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { ApplyBar } from "@/shared/ui/apply-bar";
import { Button } from "@/shared/ui/button";
import { CodeEditor } from "@/shared/ui/code-editor";
import { Dialog, DialogFooter, DialogHeader, DialogPanel, DialogPopup, DialogTitle } from "@/shared/ui/dialog";
import { Input } from "@/shared/ui/input";
import { NumberInput } from "@/shared/ui/number-input";
import { OptionSelect } from "@/shared/ui/option-select";
import { PageBody, Section } from "@/shared/ui/page-body";
import { SettingRow, SettingsGroup } from "@/shared/ui/settings";
import { Switch } from "@/shared/ui/switch";
import { toastManager } from "@/shared/ui/toast";
import { PageHeader } from "@/widgets/page-header";

export function CorePage() {
  const status = useStatus();
  const form = useCoreSettings();
  const [installing, setInstalling] = useState(false);
  const [config, setConfig] = useState<string>();
  const [controller, setController] = useState<{ host: string; port: number; secret: string }>();
  const { stored, draft, edit } = form;

  async function install() {
    setInstalling(true);
    try {
      const installed = await Settings.installCore();
      toastManager.add({ title: t.coreInstalled(installed.version), type: "success" });
      await form.reload();
    } catch (err) {
      toastManager.add({ title: t.coreInstallFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setInstalling(false);
    }
  }

  async function choose(path: string) {
    try {
      if (path === "choose") await Settings.chooseCore();
      else await Settings.setCorePath(path);
      await form.reload();
    } catch (err) {
      toastManager.add({ title: errorMessage(err), type: "error" });
    }
  }

  async function saveController() {
    if (!controller) return;
    try {
      await Settings.setController(controller.host, controller.port, controller.secret === "" ? null : controller.secret);
      setController(undefined);
      await form.reload();
    } catch (err) {
      toastManager.add({ title: t.settingsApplyFailed(), description: errorMessage(err), type: "error" });
    }
  }

  const candidates = stored?.candidates ?? [];
  const coreOptions = [
    ...candidates.map((c) => ({ value: c.path, label: `${c.path} · ${c.source}` })),
    { value: "choose", label: t.coreChoose() },
  ];

  return (
    <>
      <PageHeader id="core" />
      <PageBody>
        <Section
          action={
            <Button loading={installing} onClick={install} size="sm" variant="outline">
              <DownloadIcon aria-hidden="true" />
              {candidates.some((c) => c.source === "managed") ? t.coreUpdate() : t.coreInstall()}
            </Button>
          }
          description={t.coreMihomoHint()}
          title="Mihomo"
        >
          <SettingsGroup>
            <SettingRow description={candidates.length === 0 ? t.coreMissing() : undefined} label={t.coreExecutable()}>
              {candidates.length > 0 ? (
                <OptionSelect className="w-96 font-mono text-xs" label={t.coreExecutable()} onChange={choose} options={coreOptions} value={stored?.corePath ?? ""} />
              ) : (
                <Button onClick={() => choose("choose")} size="sm" variant="outline">
                  {t.coreChoose()}
                </Button>
              )}
            </SettingRow>
            <SettingRow label={t.coreGenerated()}>
              <Button
                onClick={async () => {
                  try {
                    setConfig(await Settings.runtimeConfig());
                  } catch (err) {
                    toastManager.add({ title: errorMessage(err), type: "error" });
                  }
                }}
                size="sm"
                variant="outline"
              >
                <FileCodeIcon aria-hidden="true" />
                {t.coreView()}
              </Button>
            </SettingRow>
          </SettingsGroup>
        </Section>

        {draft && stored && (
          <>
            <Section title={t.corePorts()}>
              <SettingsGroup>
                <SettingRow description={t.coreMixedPortHint()} label={t.heroMixedPort()}>
                  <NumberInput label={t.heroMixedPort()} max={65535} min={1} onChange={(v) => edit((d) => ({ ...d, mixedPort: v }))} value={draft.mixedPort} />
                </SettingRow>
                <SettingRow description={t.coreAllowLANHint()} label={t.coreAllowLAN()}>
                  <Switch checked={draft.allowLAN} onCheckedChange={(v) => edit((d) => ({ ...d, allowLAN: v }))} />
                </SettingRow>
                <SettingRow label="IPv6">
                  <Switch checked={draft.ipv6} onCheckedChange={(v) => edit((d) => ({ ...d, ipv6: v }))} />
                </SettingRow>
                <SettingRow description={t.coreControllerHint()} label={t.heroController()}>
                  <span className="font-mono text-muted-foreground text-xs" data-selectable>
                    {stored.controllerHost}:{stored.controllerPort}
                    {stored.hasSecret ? " · 🔒" : ""}
                  </span>
                  <Button onClick={() => setController({ host: stored.controllerHost, port: stored.controllerPort, secret: "" })} size="sm" variant="outline">
                    {t.edit()}
                  </Button>
                </SettingRow>
              </SettingsGroup>
            </Section>

            <Section title={t.coreBehavior()}>
              <SettingsGroup>
                <SettingRow label={t.coreLogLevel()}>
                  <OptionSelect
                    label={t.coreLogLevel()}
                    onChange={(v) => edit((d) => ({ ...d, logLevel: v }))}
                    options={["silent", "error", "warning", "info", "debug"].map((v) => ({ value: v, label: v }))}
                    value={draft.logLevel}
                  />
                </SettingRow>
                <SettingRow description={t.coreFindProcessHint()} label={t.coreFindProcess()}>
                  <OptionSelect
                    label={t.coreFindProcess()}
                    onChange={(v) => edit((d) => ({ ...d, findProcessMode: v }))}
                    options={[
                      { value: "always", label: t.coreFindAlways() },
                      { value: "strict", label: t.coreFindStrict() },
                      { value: "off", label: t.coreFindOff() },
                    ]}
                    value={draft.findProcessMode}
                  />
                </SettingRow>
              </SettingsGroup>
            </Section>

            <Section description={t.coreGeoHint()} title={t.coreGeo()}>
              <SettingsGroup>
                <SettingRow label={t.coreGeoAuto()}>
                  {draft.geoData.autoUpdate && (
                    <span className="flex items-center gap-1.5 text-muted-foreground text-xs">
                      <NumberInput
                        className="w-16"
                        label={t.coreGeoInterval()}
                        min={1}
                        onChange={(v) => edit((d) => ({ ...d, geoData: { ...d.geoData, updateIntervalHours: v } }))}
                        value={draft.geoData.updateIntervalHours}
                      />
                      {t.coreHours()}
                    </span>
                  )}
                  <Switch checked={draft.geoData.autoUpdate} onCheckedChange={(v) => edit((d) => ({ ...d, geoData: { ...d.geoData, autoUpdate: v } }))} />
                </SettingRow>
                <SettingRow description={t.coreGeoDatHint()} label={t.coreGeoDat()}>
                  <Switch checked={draft.geoData.usesDatMode} onCheckedChange={(v) => edit((d) => ({ ...d, geoData: { ...d.geoData, usesDatMode: v } }))} />
                </SettingRow>
                {(
                  [
                    ["geoIPURL", "GeoIP"],
                    ["geoSiteURL", "GeoSite"],
                    ["mmdbURL", "MMDB"],
                    ["asnURL", "ASN"],
                  ] as const
                ).map(([key, label]) => (
                  <SettingRow key={key} label={label} stacked>
                    <Input
                      className="font-mono text-xs"
                      onChange={(e) => edit((d) => ({ ...d, geoData: { ...d.geoData, [key]: e.target.value } }))}
                      size="sm"
                      value={draft.geoData[key]}
                    />
                  </SettingRow>
                ))}
              </SettingsGroup>
            </Section>
          </>
        )}
        <ApplyBar applying={form.applying} dirty={form.dirty} note={status?.state === "running" ? t.applyReloadNote() : undefined} onApply={form.apply} onDiscard={form.discard} />
      </PageBody>

      <Dialog onOpenChange={(open) => !open && setConfig(undefined)} open={config !== undefined}>
        <DialogPopup className="max-w-3xl">
          <DialogHeader>
            <DialogTitle>{t.coreGenerated()}</DialogTitle>
          </DialogHeader>
          <DialogPanel>
            <CodeEditor className="h-[60vh]" label={t.coreGenerated()} readOnly value={config ?? ""} />
          </DialogPanel>
          <DialogFooter>
            <Button onClick={() => App.reveal("config")} variant="ghost">
              <FolderOpenIcon aria-hidden="true" />
              {t.revealInFolder()}
            </Button>
            <Button onClick={() => config && App.copyText(config)} variant="outline">
              <CopyIcon aria-hidden="true" />
              {t.copy()}
            </Button>
          </DialogFooter>
        </DialogPopup>
      </Dialog>

      <Dialog onOpenChange={(open) => !open && setController(undefined)} open={controller !== undefined}>
        <DialogPopup className="max-w-md">
          <DialogHeader>
            <DialogTitle>{t.heroController()}</DialogTitle>
          </DialogHeader>
          {controller && (
            <DialogPanel className="flex flex-col gap-3">
              <div className="flex gap-2">
                <Input className="flex-1 font-mono" onChange={(e) => setController({ ...controller, host: e.target.value })} value={controller.host} />
                <NumberInput label={t.heroController()} max={65535} min={1} onChange={(port) => setController({ ...controller, port })} value={controller.port} />
              </div>
              <Input
                autoComplete="off"
                className="font-mono"
                onChange={(e) => setController({ ...controller, secret: e.target.value })}
                placeholder={t.coreSecretPlaceholder()}
                type="password"
                value={controller.secret}
              />
              <p className="text-muted-foreground text-xs">{t.coreControllerRestart()}</p>
            </DialogPanel>
          )}
          <DialogFooter>
            <Button onClick={() => setController(undefined)} variant="ghost">
              {t.cancel()}
            </Button>
            <Button onClick={saveController}>{t.save()}</Button>
          </DialogFooter>
        </DialogPopup>
      </Dialog>
    </>
  );
}
