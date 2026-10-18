import { t } from "@/shared/i18n";
import type { RouteId } from "./routes";

const titles: Record<RouteId, () => string> = {
  overview: () => t.pageOverview(),
  profiles: () => t.pageProfiles(),
  proxies: () => t.pageProxies(),
  topology: () => t.pageTopology(),
  connections: () => t.pageConnections(),
  logs: () => t.pageLogs(),
  rules: () => t.pageRules(),
  core: () => t.pageCore(),
  systemProxy: () => t.pageSystemProxy(),
  dns: () => t.pageDns(),
  tun: () => t.pageTun(),
  sniffer: () => t.pageSniffer(),
  resources: () => t.pageResources(),
  overrides: () => t.pageOverrides(),
  agents: () => t.pageAgents(),
  settings: () => t.pageSettings(),
  about: () => t.pageAbout(),
  setup: () => t.pageSetup(),
};

/** The title a page goes by, in the sidebar and its header. */
export function pageTitle(id: RouteId): string {
  return titles[id]();
}
