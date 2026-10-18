import { useStatus } from "@/entities/status";
import { useCoreSettings } from "@/features/edit-core-settings";
import type { SnifferSettings } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { ApplyBar } from "@/shared/ui/apply-bar";
import { ListEditor, PortListEditor } from "@/shared/ui/list-editor";
import { PageBody, Section } from "@/shared/ui/page-body";
import { SettingRow, SettingsGroup } from "@/shared/ui/settings";
import { Switch } from "@/shared/ui/switch";
import { PageHeader } from "@/widgets/page-header";

export function SnifferPage() {
  const status = useStatus();
  const form = useCoreSettings();
  const sniffer = form.draft?.sniffer;
  const set = (patch: Partial<SnifferSettings>) => form.edit((d) => ({ ...d, sniffer: { ...d.sniffer, ...patch } }));

  return (
    <>
      <PageHeader id="sniffer" />
      <PageBody>
        {sniffer && (
          <>
            <SettingsGroup>
              <SettingRow description={sniffer.isEnabled ? t.snifferManagedOn() : t.snifferManagedOff()} label={t.snifferManaged()}>
                <Switch checked={sniffer.isEnabled} onCheckedChange={(isEnabled) => set({ isEnabled })} />
              </SettingRow>
            </SettingsGroup>
            {sniffer.isEnabled && (
              <>
                <Section title={t.snifferBehavior()}>
                  <SettingsGroup>
                    <SettingRow description={t.snifferPureIPHint()} label={t.snifferPureIP()}>
                      <Switch checked={sniffer.parsePureIP} onCheckedChange={(parsePureIP) => set({ parsePureIP })} />
                    </SettingRow>
                    <SettingRow description={t.snifferForceDNSHint()} label={t.snifferForceDNS()}>
                      <Switch checked={sniffer.forceDNSMapping} onCheckedChange={(forceDNSMapping) => set({ forceDNSMapping })} />
                    </SettingRow>
                    <SettingRow description={t.snifferOverrideHint()} label={t.snifferOverride()}>
                      <Switch checked={sniffer.overrideDestination} onCheckedChange={(overrideDestination) => set({ overrideDestination })} />
                    </SettingRow>
                    <SettingRow label={t.snifferHTTPOverride()}>
                      <Switch checked={sniffer.httpOverrideDestination} onCheckedChange={(httpOverrideDestination) => set({ httpOverrideDestination })} />
                    </SettingRow>
                  </SettingsGroup>
                </Section>
                <Section description={t.snifferPortsHint()} title={t.snifferPorts()}>
                  <SettingsGroup>
                    <SettingRow label="HTTP" stacked>
                      <PortListEditor onChange={(httpPorts) => set({ httpPorts })} value={sniffer.httpPorts} />
                    </SettingRow>
                    <SettingRow label="TLS" stacked>
                      <PortListEditor onChange={(tlsPorts) => set({ tlsPorts })} value={sniffer.tlsPorts} />
                    </SettingRow>
                    <SettingRow label="QUIC" stacked>
                      <PortListEditor onChange={(quicPorts) => set({ quicPorts })} value={sniffer.quicPorts} />
                    </SettingRow>
                  </SettingsGroup>
                </Section>
                <Section title={t.snifferExceptions()}>
                  <SettingsGroup>
                    <SettingRow description={t.snifferSkipDomainHint()} label={t.snifferSkipDomain()} stacked>
                      <ListEditor onChange={(skipDomain) => set({ skipDomain })} value={sniffer.skipDomain} />
                    </SettingRow>
                    <SettingRow description={t.snifferForceDomainHint()} label={t.snifferForceDomain()} stacked>
                      <ListEditor onChange={(forceDomain) => set({ forceDomain })} value={sniffer.forceDomain} />
                    </SettingRow>
                    <SettingRow label={t.snifferSkipDst()} stacked>
                      <ListEditor onChange={(skipDstAddress) => set({ skipDstAddress })} value={sniffer.skipDstAddress} />
                    </SettingRow>
                    <SettingRow label={t.snifferSkipSrc()} stacked>
                      <ListEditor onChange={(skipSrcAddress) => set({ skipSrcAddress })} value={sniffer.skipSrcAddress} />
                    </SettingRow>
                  </SettingsGroup>
                </Section>
              </>
            )}
          </>
        )}
        <ApplyBar applying={form.applying} dirty={form.dirty} note={status?.state === "running" ? t.applyReloadNote() : undefined} onApply={form.apply} onDiscard={form.discard} />
      </PageBody>
    </>
  );
}
