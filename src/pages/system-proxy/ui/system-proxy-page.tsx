import { useEffect, useMemo, useState } from "react";
import { setStatus, useStatus } from "@/entities/status";
import { ContentionAlert, SystemProxySwitch } from "@/features/toggle-system-proxy";
import { Preferences, type Settings as ProxySettings, SystemProxy } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { useResource } from "@/shared/lib/use-resource";
import { ApplyBar } from "@/shared/ui/apply-bar";
import { Badge } from "@/shared/ui/badge";
import { CodeEditor } from "@/shared/ui/code-editor";
import { ListEditor } from "@/shared/ui/list-editor";
import { OptionSelect } from "@/shared/ui/option-select";
import { PageBody, Section } from "@/shared/ui/page-body";
import { SettingRow, SettingsGroup } from "@/shared/ui/settings";
import { Switch } from "@/shared/ui/switch";
import { toastManager } from "@/shared/ui/toast";
import { ToggleGroup, ToggleGroupItem } from "@/shared/ui/toggle-group";
import { PageHeader } from "@/widgets/page-header";

export function SystemProxyPage() {
  const status = useStatus();
  const services = useResource(() => SystemProxy.services(), []);
  const prefs = useResource(() => Preferences.get(), []);
  const [draft, setDraft] = useState<ProxySettings>();
  const [applying, setApplying] = useState(false);
  const stored = status?.systemProxy.settings;

  useEffect(() => {
    if (stored && !draft) setDraft(stored);
  }, [stored, draft]);

  const dirty = useMemo(() => {
    if (!stored || !draft) return false;
    const { port: _a, ...a } = stored;
    const { port: _b, ...b } = draft;
    return JSON.stringify(a) !== JSON.stringify(b);
  }, [stored, draft]);

  async function apply() {
    if (!draft) return;
    setApplying(true);
    try {
      const next = await SystemProxy.updateSettings(draft);
      setStatus(next);
      setDraft(next.systemProxy.settings);
      toastManager.add({ title: t.settingsApplied(), type: "success" });
    } catch (err) {
      toastManager.add({ title: t.settingsApplyFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setApplying(false);
    }
  }

  async function setReclaim(reclaim: boolean) {
    if (!prefs.data) return;
    const next = { ...prefs.data, reclaimSystemProxy: reclaim };
    prefs.set(next);
    await Preferences.save(next);
  }

  const sp = status?.systemProxy;
  const state = !sp
    ? undefined
    : sp.active && sp.ownership === "takenOver"
      ? { label: t.spStateTaken(), variant: "warning" as const }
      : sp.active
        ? { label: t.spStateActive(sp.applied?.service ?? ""), variant: "success" as const }
        : sp.enabled
          ? { label: t.heroProxyWaiting(), variant: "info" as const }
          : { label: t.spStateOff(), variant: "secondary" as const };

  return (
    <>
      <PageHeader id="systemProxy" />
      <PageBody>
        <SettingsGroup>
          <SettingRow
            description={
              state && (
                <Badge className="mt-1" variant={state.variant}>
                  {state.label}
                </Badge>
              )
            }
            htmlFor="sp-switch"
            label={t.spUse()}
          >
            <SystemProxySwitch id="sp-switch" />
          </SettingRow>
          <SettingRow description={t.spReclaimHint()} label={t.spReclaim()}>
            <Switch checked={prefs.data?.reclaimSystemProxy ?? false} onCheckedChange={setReclaim} />
          </SettingRow>
        </SettingsGroup>
        <ContentionAlert />

        {draft && (
          <>
            <Section description={t.spHowHint()} title={t.spHow()}>
              <SettingsGroup>
                <SettingRow label={t.spMode()}>
                  <ToggleGroup onValueChange={(v) => v[0] && setDraft({ ...draft, mode: v[0] as ProxySettings["mode"] })} size="sm" value={[draft.mode]} variant="outline">
                    <ToggleGroupItem value="manual">{t.spManual()}</ToggleGroupItem>
                    <ToggleGroupItem value="pac">PAC</ToggleGroupItem>
                  </ToggleGroup>
                </SettingRow>
                <SettingRow description={t.spServiceHint()} label={t.spService()}>
                  <OptionSelect
                    className="w-52"
                    label={t.spService()}
                    onChange={(v) => setDraft({ ...draft, networkService: v })}
                    options={[{ value: "Automatic", label: t.spAutomatic() }, ...(services.data ?? []).map((s) => ({ value: s, label: s }))]}
                    value={draft.networkService || "Automatic"}
                  />
                </SettingRow>
                {draft.mode === "manual" ? (
                  <SettingRow description={t.spBypassHint()} label={t.spBypass()} stacked>
                    <ListEditor onChange={(bypassList) => setDraft({ ...draft, bypassList })} value={draft.bypassList} />
                  </SettingRow>
                ) : (
                  <SettingRow description={t.spPacHint()} label={t.spPac()} stacked>
                    <CodeEditor label={t.spPac()} minRows={8} onChange={(pacScript) => setDraft({ ...draft, pacScript })} value={draft.pacScript} />
                  </SettingRow>
                )}
              </SettingsGroup>
            </Section>
            <ApplyBar applying={applying} dirty={dirty} onApply={apply} onDiscard={() => setDraft(stored)} />
          </>
        )}
      </PageBody>
    </>
  );
}
