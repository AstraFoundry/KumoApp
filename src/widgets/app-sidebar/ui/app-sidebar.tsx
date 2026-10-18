import {
  ArrowDownIcon,
  ArrowUpIcon,
  BotIcon,
  CableIcon,
  CpuIcon,
  FileStackIcon,
  GaugeIcon,
  GlobeIcon,
  InfoIcon,
  LayersIcon,
  LaptopIcon,
  ListFilterIcon,
  PackageIcon,
  RouteIcon,
  ScanSearchIcon,
  ScrollTextIcon,
  ServerIcon,
  Settings2Icon,
  WorkflowIcon,
} from "lucide-react";
import type { ComponentType } from "react";
import { StatusDot, useStatus, useTraffic } from "@/entities/status";
import { pageTitle } from "@/shared/config/page-titles";
import { type RouteId, routes } from "@/shared/config/routes";
import { t } from "@/shared/i18n";
import { formatSpeed } from "@/shared/lib/format";
import { Link } from "@/shared/lib/router";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/shared/ui/sidebar";

interface Item {
  id: RouteId;
  icon: ComponentType<{ className?: string }>;
}

const sections: { label: () => string; items: Item[] }[] = [
  {
    label: () => t.navDaily(),
    items: [
      { id: "overview", icon: GaugeIcon },
      { id: "profiles", icon: FileStackIcon },
      { id: "proxies", icon: GlobeIcon },
      { id: "topology", icon: WorkflowIcon },
    ],
  },
  {
    label: () => t.navInspect(),
    items: [
      { id: "connections", icon: CableIcon },
      { id: "logs", icon: ScrollTextIcon },
      { id: "rules", icon: ListFilterIcon },
    ],
  },
  {
    label: () => t.navConfigure(),
    items: [
      { id: "core", icon: CpuIcon },
      { id: "systemProxy", icon: LaptopIcon },
      { id: "dns", icon: ServerIcon },
      { id: "tun", icon: RouteIcon },
      { id: "sniffer", icon: ScanSearchIcon },
      { id: "resources", icon: PackageIcon },
      { id: "overrides", icon: LayersIcon },
    ],
  },
  {
    label: () => t.navApp(),
    items: [
      { id: "agents", icon: BotIcon },
      { id: "settings", icon: Settings2Icon },
      { id: "about", icon: InfoIcon },
    ],
  },
];

/** The source list: every page, grouped by how often it is needed. */
export function AppSidebar({ active }: { active: RouteId }) {
  return (
    <Sidebar className="border-sidebar-border border-e" collapsible="none">
      {/* Room for the window controls; the empty strip drags the window. */}
      <SidebarHeader className="h-3 shrink-0 p-0 in-[.mac]:h-[max(3rem,var(--mygo-titlebar-height,0px))] in-[.mac]:[--app-region:drag]" />
      <SidebarContent className="gap-0 px-1 pb-2">
        {sections.map((section) => (
          <SidebarGroup className="py-1" key={section.items[0]?.id}>
            <SidebarGroupLabel className="h-7 text-[0.6875rem] text-sidebar-foreground/70 uppercase tracking-wide">
              {section.label()}
            </SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu className="gap-px">
                {section.items.map((item) => (
                  <SidebarMenuItem key={item.id}>
                    <SidebarMenuButton
                      className="h-7.5 gap-2.5 text-[0.8125rem] data-[active=true]:text-foreground"
                      isActive={active === item.id}
                      render={<Link href={routes[item.id]} />}
                    >
                      <item.icon className="size-4 opacity-70" />
                      <span>{pageTitle(item.id)}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        ))}
      </SidebarContent>
      <SidebarFooter className="p-2">
        <StatusFooter />
      </SidebarFooter>
    </Sidebar>
  );
}

function StatusFooter() {
  const status = useStatus();
  const traffic = useTraffic();
  const last = traffic.at(-1);
  const running = status?.state === "running";
  return (
    <Link
      className="flex flex-col gap-1 rounded-lg px-2 py-1.5 text-sidebar-foreground text-xs outline-none transition-colors hover:bg-sidebar-accent focus-visible:ring-2 focus-visible:ring-ring"
      href={routes.overview}
    >
      <span className="flex items-center gap-2 font-medium text-foreground">
        <StatusDot state={status?.state} />
        {running ? t.stateRunning() : status?.state === "failed" ? t.stateFailed() : t.stateStopped()}
      </span>
      {running && (
        <span className="flex gap-3 tabular-nums">
          <span className="flex items-center gap-0.5">
            <ArrowUpIcon aria-label={t.trafficUpload()} className="size-3" />
            {formatSpeed(last?.up ?? 0)}
          </span>
          <span className="flex items-center gap-0.5">
            <ArrowDownIcon aria-label={t.trafficDownload()} className="size-3" />
            {formatSpeed(last?.down ?? 0)}
          </span>
        </span>
      )}
    </Link>
  );
}
