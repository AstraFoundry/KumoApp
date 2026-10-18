import { type ComponentType, lazy, Suspense } from "react";
import { OverviewPage } from "@/pages/overview";
import type { RouteId } from "@/shared/config/routes";

// Overview loads with the app; every other page loads when first visited.
const pages: Partial<Record<RouteId, ComponentType>> = {
  overview: OverviewPage,
  profiles: lazy(() => import("@/pages/profiles").then((m) => ({ default: m.ProfilesPage }))),
  proxies: lazy(() => import("@/pages/proxies").then((m) => ({ default: m.ProxiesPage }))),
  topology: lazy(() => import("@/pages/topology").then((m) => ({ default: m.TopologyPage }))),
};

export function Page({ id }: { id: RouteId }) {
  const Component = pages[id] ?? OverviewPage;
  return (
    <Suspense fallback={null}>
      <Component />
    </Suspense>
  );
}
