import { useStatus } from "@/entities/status";
import { useCoreSettings } from "@/features/edit-core-settings";
import type { DNSSettings } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { ApplyBar } from "@/shared/ui/apply-bar";
import { Input } from "@/shared/ui/input";
import { ListEditor } from "@/shared/ui/list-editor";
import { MapEditor } from "@/shared/ui/map-editor";
import { NumberInput } from "@/shared/ui/number-input";
import { OptionSelect } from "@/shared/ui/option-select";
import { PageBody, Section } from "@/shared/ui/page-body";
import { SettingRow, SettingsGroup } from "@/shared/ui/settings";
import { Switch } from "@/shared/ui/switch";
import { PageHeader } from "@/widgets/page-header";

export function DnsPage() {
  const status = useStatus();
  const form = useCoreSettings();
  const dns = form.draft?.dns;
  const set = (patch: Partial<DNSSettings>) => form.edit((d) => ({ ...d, dns: { ...d.dns, ...patch } }));

  return (
    <>
      <PageHeader id="dns" />
      <PageBody>
        {dns && (
          <>
            <SettingsGroup>
              <SettingRow description={dns.isEnabled ? t.dnsManagedOn() : t.dnsManagedOff()} label={t.dnsManaged()}>
                <Switch checked={dns.isEnabled} onCheckedChange={(isEnabled) => set({ isEnabled })} />
              </SettingRow>
            </SettingsGroup>
            {dns.isEnabled && (
              <>
                <Section title={t.dnsResolution()}>
                  <SettingsGroup>
                    <SettingRow description={t.dnsEnhancedHint()} label={t.dnsEnhanced()}>
                      <OptionSelect
                        label={t.dnsEnhanced()}
                        onChange={(enhancedMode) => set({ enhancedMode })}
                        options={[
                          { value: "fake-ip", label: "fake-ip" },
                          { value: "redir-host", label: "redir-host" },
                          { value: "normal", label: "normal" },
                        ]}
                        value={dns.enhancedMode}
                      />
                    </SettingRow>
                    {dns.enhancedMode === "fake-ip" && (
                      <>
                        <SettingRow label={t.dnsFakeIPRange()}>
                          <Input className="w-40 font-mono" onChange={(e) => set({ fakeIPRange: e.target.value })} size="sm" value={dns.fakeIPRange} />
                        </SettingRow>
                        <SettingRow description={t.dnsFakeIPFilterHint()} label={t.dnsFakeIPFilter()} stacked>
                          <div className="flex w-full flex-col gap-2">
                            <OptionSelect
                              className="w-40"
                              label={t.dnsFakeIPFilterMode()}
                              onChange={(fakeIPFilterMode) => set({ fakeIPFilterMode })}
                              options={[
                                { value: "blacklist", label: t.dnsBlacklist() },
                                { value: "whitelist", label: t.dnsWhitelist() },
                              ]}
                              value={dns.fakeIPFilterMode || "blacklist"}
                            />
                            <ListEditor onChange={(fakeIPFilter) => set({ fakeIPFilter })} value={dns.fakeIPFilter} />
                          </div>
                        </SettingRow>
                      </>
                    )}
                    <SettingRow description={t.dnsListenHint()} label={t.dnsListen()}>
                      <Input className="w-40 font-mono" onChange={(e) => set({ listen: e.target.value })} placeholder="0.0.0.0:1053" size="sm" value={dns.listen} />
                    </SettingRow>
                    <SettingRow label="IPv6">
                      {dns.ipv6 && (
                        <span className="flex items-center gap-1.5 text-muted-foreground text-xs">
                          <NumberInput className="w-20" label={t.dnsIPv6Timeout()} min={0} onChange={(ipv6Timeout) => set({ ipv6Timeout })} value={dns.ipv6Timeout} />
                          ms
                        </span>
                      )}
                      <Switch checked={dns.ipv6} onCheckedChange={(ipv6) => set({ ipv6 })} />
                    </SettingRow>
                    <SettingRow description={t.dnsRespectRulesHint()} label={t.dnsRespectRules()}>
                      <Switch checked={dns.respectRules} onCheckedChange={(respectRules) => set({ respectRules })} />
                    </SettingRow>
                    <SettingRow label={t.dnsPreferH3()}>
                      <Switch checked={dns.preferH3} onCheckedChange={(preferH3) => set({ preferH3 })} />
                    </SettingRow>
                    <SettingRow label={t.dnsUseHosts()}>
                      <Switch checked={dns.useHosts} onCheckedChange={(useHosts) => set({ useHosts })} />
                    </SettingRow>
                    <SettingRow label={t.dnsUseSystemHosts()}>
                      <Switch checked={dns.useSystemHosts} onCheckedChange={(useSystemHosts) => set({ useSystemHosts })} />
                    </SettingRow>
                    <SettingRow label={t.dnsCache()}>
                      <OptionSelect
                        label={t.dnsCache()}
                        onChange={(cacheAlgorithm) => set({ cacheAlgorithm: cacheAlgorithm === "default" ? "" : cacheAlgorithm })}
                        options={[
                          { value: "default", label: t.dnsCacheDefault() },
                          { value: "lru", label: "LRU" },
                          { value: "arc", label: "ARC" },
                        ]}
                        value={dns.cacheAlgorithm || "default"}
                      />
                    </SettingRow>
                  </SettingsGroup>
                </Section>

                <Section description={t.dnsServersHint()} title={t.dnsServers()}>
                  <SettingsGroup>
                    <SettingRow description={t.dnsNameserverHint()} label={t.dnsNameserver()} stacked>
                      <ListEditor onChange={(nameserver) => set({ nameserver })} value={dns.nameserver} />
                    </SettingRow>
                    <SettingRow description={t.dnsDefaultHint()} label={t.dnsDefault()} stacked>
                      <ListEditor onChange={(defaultNameserver) => set({ defaultNameserver })} value={dns.defaultNameserver} />
                    </SettingRow>
                    <SettingRow description={t.dnsProxyServerHint()} label={t.dnsProxyServer()} stacked>
                      <ListEditor onChange={(proxyServerNameserver) => set({ proxyServerNameserver })} value={dns.proxyServerNameserver} />
                    </SettingRow>
                    <SettingRow description={t.dnsDirectHint()} label={t.dnsDirect()} stacked>
                      <ListEditor onChange={(directNameserver) => set({ directNameserver })} value={dns.directNameserver} />
                    </SettingRow>
                    <SettingRow label={t.dnsDirectFollow()}>
                      <Switch checked={dns.directNameserverFollowPolicy} onCheckedChange={(directNameserverFollowPolicy) => set({ directNameserverFollowPolicy })} />
                    </SettingRow>
                    <SettingRow description={t.dnsFallbackHint()} label={t.dnsFallback()} stacked>
                      <ListEditor onChange={(fallback) => set({ fallback })} value={dns.fallback} />
                    </SettingRow>
                  </SettingsGroup>
                </Section>

                {dns.fallback.length > 0 && (
                  <Section description={t.dnsFallbackFilterHint()} title={t.dnsFallbackFilter()}>
                    <SettingsGroup>
                      <SettingRow label="GeoIP">
                        {dns.fallbackFilter.geoip && (
                          <Input
                            className="w-20 font-mono"
                            onChange={(e) => set({ fallbackFilter: { ...dns.fallbackFilter, geoipCode: e.target.value.toUpperCase() } })}
                            placeholder="CN"
                            size="sm"
                            value={dns.fallbackFilter.geoipCode}
                          />
                        )}
                        <Switch checked={dns.fallbackFilter.geoip} onCheckedChange={(geoip) => set({ fallbackFilter: { ...dns.fallbackFilter, geoip } })} />
                      </SettingRow>
                      <SettingRow label="GeoSite" stacked>
                        <ListEditor onChange={(geosite) => set({ fallbackFilter: { ...dns.fallbackFilter, geosite } })} value={dns.fallbackFilter.geosite ?? []} />
                      </SettingRow>
                      <SettingRow label="IP-CIDR" stacked>
                        <ListEditor onChange={(ipcidr) => set({ fallbackFilter: { ...dns.fallbackFilter, ipcidr } })} value={dns.fallbackFilter.ipcidr ?? []} />
                      </SettingRow>
                      <SettingRow label={t.dnsDomain()} stacked>
                        <ListEditor onChange={(domain) => set({ fallbackFilter: { ...dns.fallbackFilter, domain } })} value={dns.fallbackFilter.domain ?? []} />
                      </SettingRow>
                    </SettingsGroup>
                  </Section>
                )}

                <Section description={t.dnsPoliciesHint()} title={t.dnsPolicies()}>
                  <SettingsGroup>
                    <SettingRow label="nameserver-policy" stacked>
                      <MapEditor keyPlaceholder="geosite:cn" onChange={(nameserverPolicy) => set({ nameserverPolicy })} value={dns.nameserverPolicy} valuePlaceholder="223.5.5.5, 119.29.29.29" />
                    </SettingRow>
                    <SettingRow label="proxy-server-nameserver-policy" stacked>
                      <MapEditor
                        keyPlaceholder="+.example.com"
                        onChange={(proxyServerNameserverPolicy) => set({ proxyServerNameserverPolicy })}
                        value={dns.proxyServerNameserverPolicy}
                        valuePlaceholder="1.1.1.1"
                      />
                    </SettingRow>
                    <SettingRow description={t.dnsHostsHint()} label="hosts" stacked>
                      <MapEditor keyPlaceholder="router.lan" onChange={(hosts) => set({ hosts })} value={dns.hosts} valuePlaceholder="192.168.1.1" />
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
