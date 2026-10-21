import { events } from "@/shared/api/mygo";
import { isMyGo } from "mygo-runtime";
import { useEffect } from "react";
import { routeOf } from "@/shared/config/routes";
import { useLocaleVersion } from "@/shared/i18n";
import { navigate, useLocation } from "@/shared/lib/router";
import { SidebarProvider } from "@/shared/ui/sidebar";
import { AppSidebar } from "@/widgets/app-sidebar";
import { Providers } from "./providers/providers";
import { Page } from "./router";

export function App() {
  const version = useLocaleVersion();
  const { path } = useLocation();
  const id = routeOf(path);

  useEffect(() => {
    if (!isMyGo()) return;
    return events.navigate.on((route) => navigate(route));
  }, []);

  return (
    <Providers key={version}>
      {id === "setup" ? (
        <Page id="setup" />
      ) : (
        <SidebarProvider className="h-svh min-h-0 overflow-hidden" style={{ "--sidebar-width": "13.5rem" } as React.CSSProperties}>
          <AppSidebar active={id} />
          <main className="flex min-w-0 flex-1 flex-col bg-background">
            <Page id={id} />
          </main>
        </SidebarProvider>
      )}
    </Providers>
  );
}
