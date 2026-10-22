import { CheckCircle2Icon, ShieldIcon } from "lucide-react";
import { useState } from "react";
import { useStatus } from "@/entities/status";
import { useCoreSettings } from "@/features/edit-core-settings";
import { TunSwitch } from "@/features/toggle-tun";
import { Settings, type TUNSettings } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { Alert, AlertAction, AlertDescription, AlertTitle } from "@/shared/ui/alert";
import { ApplyBar } from "@/shared/ui/apply-bar";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { ListEditor } from "@/shared/ui/list-editor";
import { NumberInput } from "@/shared/ui/number-input";
import { OptionSelect } from "@/shared/ui/option-select";
import { PageBody, Section } from "@/shared/ui/page-body";
import { SettingRow, SettingsGroup } from "@/shared/ui/settings";
import { Switch } from "@/shared/ui/switch";
import { toastManager } from "@/shared/ui/toast";
import { PageHeader } from "@/widgets/page-header";

export function TunPage() {
  const status = useStatus();
  const form = useCoreSettings();
  const [granting, setGranting] = useState(false);
  const tun = form.draft?.tun;
  const permission = form.stored?.tunPermission;
  const set = (patch: Partial<TUNSettings>) => form.edit((d) => ({ ...d, tun: { ...d.tun, ...patch } }));

  async function grant() {
    setGranting(true);
    try {
      await Settings.grantTUN();
      await form.reload();
      toastManager.add({ title: t.tunGranted(), type: "success" });
    } catch (err) {
      toastManager.add({ title: t.tunGrantFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setGranting(false);
    }
  }

  return (
    <>
      <PageHeader id="tun" />
      <PageBody>
        {permission &&
          (permission.granted ? (
            <Alert variant="success">
              <CheckCircle2Icon aria-hidden="true" />
              <AlertTitle>{t.tunPermissionOK()}</AlertTitle>
              <AlertDescription>{t.tunPermissionMethod(permission.method ?? "")}</AlertDescription>
            </Alert>
          ) : (
            <Alert variant="info">
              <ShieldIcon aria-hidden="true" />
              <AlertTitle>{t.tunPermissionNeeded()}</AlertTitle>
              <AlertDescription>{t.tunPermissionHint()}</AlertDescription>
              <AlertAction>
                <Button disabled={status?.state === "running"} loading={granting} onClick={grant} size="sm">
                  {t.tunGrant()}
                </Button>
              </AlertAction>
            </Alert>
          ))}
        <SettingsGroup>
          <SettingRow description={t.heroTunDetail()} htmlFor="tun-switch" label={t.tunUse()}>
            <TunSwitch id="tun-switch" />
          </SettingRow>
        </SettingsGroup>
        {tun && (
          <Section description={t.tunAdvancedHint()} title={t.tunAdvanced()}>
            <SettingsGroup>
              <SettingRow description={tun.stack === "mips" ? t.tunStackMipsHint() : undefined} label={t.tunStack()}>
                <OptionSelect
                  label={t.tunStack()}
                  onChange={(stack) => set({ stack })}
                  options={[
                    { value: "mips", label: "MIPS" },
                    { value: "mixed", label: "Mixed" },
                    { value: "gvisor", label: "gVisor" },
                    { value: "system", label: "System" },
                  ]}
                  value={tun.stack}
                />
              </SettingRow>
              <SettingRow description={t.tunAutoRouteHint()} label={t.tunAutoRoute()}>
                <Switch checked={tun.autoRoute} onCheckedChange={(autoRoute) => set({ autoRoute })} />
              </SettingRow>
              <SettingRow label={t.tunAutoDetect()}>
                <Switch checked={tun.autoDetectInterface} onCheckedChange={(autoDetectInterface) => set({ autoDetectInterface })} />
              </SettingRow>
              <SettingRow description={t.tunStrictHint()} label={t.tunStrict()}>
                <Switch checked={tun.strictRoute} onCheckedChange={(strictRoute) => set({ strictRoute })} />
              </SettingRow>
              <SettingRow description={t.tunAutoRedirectHint()} label={t.tunAutoRedirect()}>
                <Switch checked={tun.autoRedirect} onCheckedChange={(autoRedirect) => set({ autoRedirect })} />
              </SettingRow>
              <SettingRow label={t.tunDisableICMP()}>
                <Switch checked={tun.disableICMPForwarding} onCheckedChange={(disableICMPForwarding) => set({ disableICMPForwarding })} />
              </SettingRow>
              <SettingRow label="MTU">
                <NumberInput label="MTU" max={9000} min={576} onChange={(mtu) => set({ mtu })} value={tun.mtu} />
              </SettingRow>
              <SettingRow description={t.tunDeviceHint()} label={t.tunDevice()}>
                <Input className="w-32 font-mono" onChange={(e) => set({ device: e.target.value })} placeholder="utun1024" size="sm" value={tun.device ?? ""} />
              </SettingRow>
              <SettingRow description={t.tunHijackHint()} label={t.tunHijack()} stacked>
                <ListEditor onChange={(dnsHijack) => set({ dnsHijack })} value={tun.dnsHijack} />
              </SettingRow>
              <SettingRow description={t.tunExcludeHint()} label={t.tunExclude()} stacked>
                <ListEditor onChange={(routeExcludeAddress) => set({ routeExcludeAddress })} placeholder="192.168.0.0/16" value={tun.routeExcludeAddress} />
              </SettingRow>
            </SettingsGroup>
          </Section>
        )}
        <ApplyBar applying={form.applying} dirty={form.dirty} onApply={form.apply} onDiscard={form.discard} />
      </PageBody>
    </>
  );
}
