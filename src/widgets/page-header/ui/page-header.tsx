import type { ReactNode } from "react";
import { AgentPromptMenu } from "@/features/copy-agent-prompt";
import { pageTitle } from "@/shared/config/page-titles";
import type { RouteId } from "@/shared/config/routes";

/**
 * The bar above a page: its title on the left, its actions on the right,
 * and the quiet agent-prompt button last. The bar drags the window.
 */
export function PageHeader({ id, actions }: { id: RouteId; actions?: ReactNode }) {
  return (
    <header className="flex h-12 shrink-0 items-center gap-3 border-b ps-5 pe-[calc(var(--mygo-titlebar-inset-right,0px)+1.25rem)] [--app-region:drag]">
      <h1 className="min-w-0 truncate font-semibold text-[0.9375rem]">{pageTitle(id)}</h1>
      <div className="ms-auto flex items-center gap-2 [--app-region:no-drag]">
        {actions}
        <AgentPromptMenu page={id} />
      </div>
    </header>
  );
}
