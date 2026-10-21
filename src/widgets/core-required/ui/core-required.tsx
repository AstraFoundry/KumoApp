import { PlugZapIcon } from "lucide-react";
import type { ReactNode } from "react";
import { useStatus } from "@/entities/status";
import { CoreToggleButton } from "@/features/toggle-core";
import { t } from "@/shared/i18n";
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/shared/ui/empty";

/**
 * Pages that read the running core show this while it is off, with the
 * one action that changes that.
 */
export function CoreRequired({ children, title }: { children: ReactNode; title: string }) {
  const status = useStatus();
  if (!status) return null;
  if (status.state === "running") return <>{children}</>;
  return (
    <Empty className="flex-1">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <PlugZapIcon aria-hidden="true" />
        </EmptyMedia>
        <EmptyTitle>{title}</EmptyTitle>
        <EmptyDescription>{t.coreRequiredHint()}</EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <CoreToggleButton size="default" />
      </EmptyContent>
    </Empty>
  );
}

/** Renders children only while the core runs: header actions of such pages. */
export function WhenRunning({ children }: { children: ReactNode }) {
  const status = useStatus();
  return status?.state === "running" ? <>{children}</> : null;
}
