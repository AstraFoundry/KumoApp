/** Every page of the app. The id also names the page's agent guidance. */
export const routes = {
  overview: "/",
  profiles: "/profiles",
  proxies: "/proxies",
  topology: "/topology",
  connections: "/connections",
  logs: "/logs",
  rules: "/rules",
  core: "/core",
  systemProxy: "/system-proxy",
  dns: "/dns",
  tun: "/tun",
  sniffer: "/sniffer",
  resources: "/resources",
  overrides: "/overrides",
  agents: "/agents",
  settings: "/settings",
  about: "/about",
  setup: "/setup",
} as const;

export type RouteId = keyof typeof routes;

/** The route id of a path, "overview" for unknown paths. */
export function routeOf(path: string): RouteId {
  const match = (Object.keys(routes) as RouteId[]).find((id) => routes[id] === path);
  return match ?? "overview";
}
