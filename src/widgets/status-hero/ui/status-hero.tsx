import { AlertCircleIcon } from "lucide-react";
import { StatusDot, useStatus } from "@/entities/status";
import { ContentionAlert, SystemProxySwitch } from "@/features/toggle-system-proxy";
import { TunSwitch } from "@/features/toggle-tun";
import { routes } from "@/shared/config/routes";
import { t } from "@/shared/i18n";
import { formatUptime } from "@/shared/lib/format";
import { Link } from "@/shared/lib/router";
import { cn } from "@/shared/lib/utils";
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert";
import { Card } from "@/shared/ui/card";

function modeLabel(mode: string | undefined) {
  if (mode === "global") return t.modeGlobal();
  if (mode === "direct") return t.modeDirect();
  return t.modeRule();
}

/**
 * The first thing on Overview, answering at a glance: is Kumo connected,
 * how, and does the system send traffic to it.
 */
export function StatusHero() {
  const status = useStatus();
  const running = status?.state === "running";
  const failed = status?.state === "failed";

  return (
    <div className="flex flex-col gap-3">
      <Card className="overflow-hidden">
        <div className="relative flex items-center gap-5 px-6 py-5">
          <div
            aria-hidden="true"
            className={cn(
              "pointer-events-none absolute inset-0 opacity-0 transition-opacity duration-700 ease-[var(--ease-out)]",
              running && "opacity-100",
            )}
            style={{ background: "radial-gradient(120% 140% at 0% 0%, var(--brand-soft), transparent 60%)" }}
          />
          <div className="relative flex min-w-0 flex-1 flex-col gap-1.5">
            <div className="flex items-center gap-2.5">
              <StatusDot className="size-2.5" state={status?.state} />
              <span className="font-semibold text-xl tracking-tight">
                {running ? t.heroConnected() : failed ? t.heroFailed() : t.heroDisconnected()}
              </span>
            </div>
            <p className="truncate text-muted-foreground text-sm">
              {status ? (
                <>
                  <Link className="text-foreground underline-offset-4 hover:underline" href={routes.profiles}>
                    {status.profile.name}
                  </Link>
                  {" · "}
                  {modeLabel(status.mode)}
                  {running && (
                    <>
                      {" · "}
                      {t.heroUptime(formatUptime(status.startedAt))}
                    </>
                  )}
                </>
              ) : (
                " "
              )}
            </p>
          </div>
          {running && (
            <dl className="relative hidden shrink-0 gap-6 text-right text-xs sm:flex">
              <div>
                <dt className="text-muted-foreground">{t.heroMixedPort()}</dt>
                <dd className="font-medium font-mono tabular-nums" data-selectable>
                  {status?.mixedPort}
                </dd>
              </div>
              <div>
                <dt className="text-muted-foreground">{t.heroController()}</dt>
                <dd className="font-medium font-mono tabular-nums" data-selectable>
                  {status?.controller}
                </dd>
              </div>
            </dl>
          )}
        </div>
        <div className="grid grid-cols-2 divide-x border-t">
          <Toggle
            control={<SystemProxySwitch id="hero-system-proxy" />}
            detail={systemProxyDetail(status)}
            htmlFor="hero-system-proxy"
            label={t.pageSystemProxy()}
          />
          <Toggle control={<TunSwitch id="hero-tun" />} detail={t.heroTunDetail()} htmlFor="hero-tun" label={t.pageTun()} />
        </div>
      </Card>
      <ContentionAlert />
      {failed && status?.message && (
        <Alert variant="error">
          <AlertCircleIcon aria-hidden="true" />
          <AlertTitle>{t.heroFailedTitle()}</AlertTitle>
          <AlertDescription data-selectable>{status.message}</AlertDescription>
        </Alert>
      )}
    </div>
  );
}

function systemProxyDetail(status: ReturnType<typeof useStatus>): string {
  if (!status) return "";
  const sp = status.systemProxy;
  if (sp.active) return t.heroProxyActive(sp.applied?.service ?? "", String(status.mixedPort));
  if (sp.enabled) return t.heroProxyWaiting();
  return t.heroProxyOff();
}

function Toggle({ label, detail, control, htmlFor }: { label: string; detail: string; control: React.ReactNode; htmlFor: string }) {
  return (
    <div className="flex items-center gap-3 px-6 py-3.5">
      <label className="flex min-w-0 flex-1 flex-col gap-0.5" htmlFor={htmlFor}>
        <span className="font-medium text-sm">{label}</span>
        <span className="truncate text-muted-foreground text-xs">{detail}</span>
      </label>
      {control}
    </div>
  );
}
