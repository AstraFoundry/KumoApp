import { t } from "@/shared/i18n";

// The prompts themselves are English, for agents; their menu labels follow
// the app's language.
const labels: Record<string, () => string> = {
  "overview.connect": () => t.intentOverviewConnect(),
  "overview.diagnose": () => t.intentOverviewDiagnose(),
  "overview.mode": () => t.intentOverviewMode(),
  "profiles.import": () => t.intentProfilesImport(),
  "profiles.refresh": () => t.intentProfilesRefresh(),
  "profiles.switch": () => t.intentProfilesSwitch(),
  "proxies.fastest": () => t.intentProxiesFastest(),
  "proxies.select": () => t.intentProxiesSelect(),
  "topology.explain": () => t.intentTopologyExplain(),
  "topology.route": () => t.intentTopologyRoute(),
  "connections.inspect": () => t.intentConnectionsInspect(),
  "connections.close": () => t.intentConnectionsClose(),
  "logs.errors": () => t.intentLogsErrors(),
  "logs.follow": () => t.intentLogsFollow(),
  "rules.toggle": () => t.intentRulesToggle(),
  "core.install": () => t.intentCoreInstall(),
  "core.ports": () => t.intentCorePorts(),
  "systemProxy.enable": () => t.intentSystemProxyEnable(),
  "systemProxy.contention": () => t.intentSystemProxyContention(),
  "dns.manage": () => t.intentDnsManage(),
  "tun.enable": () => t.intentTunEnable(),
  "sniffer.manage": () => t.intentSnifferManage(),
  "resources.update": () => t.intentResourcesUpdate(),
  "overrides.add": () => t.intentOverridesAdd(),
  "overrides.order": () => t.intentOverridesOrder(),
  "agents.skills": () => t.intentAgentsSkills(),
  "settings.cli": () => t.intentSettingsCli(),
};

/** The menu label of a page intent, in the app's language. */
export function intentLabel(page: string, intent: string, fallback: string): string {
  return labels[`${page}.${intent}`]?.() ?? fallback;
}
