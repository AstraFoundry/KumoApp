import { type ComponentType, lazy, Suspense } from "react";
import { OverviewPage } from "@/pages/overview";
import type { RouteId } from "@/shared/config/routes";

// Overview loads with the app; every other page loads when first visited.
const pages: Record<RouteId, ComponentType> = {
  overview: OverviewPage,
  profiles: lazy(() => import("@/pages/profiles").then((m) => ({ default: m.ProfilesPage }))),
  proxies: lazy(() => import("@/pages/proxies").then((m) => ({ default: m.ProxiesPage }))),
  topology: lazy(() => import("@/pages/topology").then((m) => ({ default: m.TopologyPage }))),
  connections: lazy(() => import("@/pages/connections").then((m) => ({ default: m.ConnectionsPage }))),
  logs: lazy(() => import("@/pages/logs").then((m) => ({ default: m.LogsPage }))),
  rules: lazy(() => import("@/pages/rules").then((m) => ({ default: m.RulesPage }))),
  core: lazy(() => import("@/pages/core").then((m) => ({ default: m.CorePage }))),
  systemProxy: lazy(() => import("@/pages/system-proxy").then((m) => ({ default: m.SystemProxyPage }))),
  dns: lazy(() => import("@/pages/dns").then((m) => ({ default: m.DnsPage }))),
  tun: lazy(() => import("@/pages/tun").then((m) => ({ default: m.TunPage }))),
  sniffer: lazy(() => import("@/pages/sniffer").then((m) => ({ default: m.SnifferPage }))),
  resources: lazy(() => import("@/pages/resources").then((m) => ({ default: m.ResourcesPage }))),
  overrides: lazy(() => import("@/pages/overrides").then((m) => ({ default: m.OverridesPage }))),
  agents: lazy(() => import("@/pages/agents").then((m) => ({ default: m.AgentsPage }))),
  settings: lazy(() => import("@/pages/settings").then((m) => ({ default: m.SettingsPage }))),
  about: lazy(() => import("@/pages/about").then((m) => ({ default: m.AboutPage }))),
  setup: lazy(() => import("@/pages/setup").then((m) => ({ default: m.SetupPage }))),
};

export function Page({ id }: { id: RouteId }) {
  const Component = pages[id];
  return (
    <Suspense fallback={null}>
      <Component />
    </Suspense>
  );
}
