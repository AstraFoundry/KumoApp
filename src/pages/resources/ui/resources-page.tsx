import { GlobeIcon, RefreshCwIcon } from "lucide-react";
import { useState } from "react";
import { UsageMeter } from "@/entities/profile";
import { Inspect } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { formatRelative } from "@/shared/lib/format";
import { useResource } from "@/shared/lib/use-resource";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { PageBody, Section } from "@/shared/ui/page-body";
import { toastManager } from "@/shared/ui/toast";
import { CoreRequired } from "@/widgets/core-required";
import { PageHeader } from "@/widgets/page-header";

export function ResourcesPage() {
  const providers = useResource(() => Inspect.providers(), []);
  const [busy, setBusy] = useState<Set<string>>(new Set());
  const [geo, setGeo] = useState(false);

  async function update(kind: "proxy" | "rule", name: string) {
    const key = `${kind}:${name}`;
    setBusy((s) => new Set(s).add(key));
    try {
      await Inspect.updateProvider(kind, name);
      await providers.reload();
    } catch (err) {
      toastManager.add({ title: t.resourcesUpdateFailed(name), description: errorMessage(err), type: "error" });
    } finally {
      setBusy((s) => {
        const next = new Set(s);
        next.delete(key);
        return next;
      });
    }
  }

  async function updateAll(kind: "proxy" | "rule") {
    const list = kind === "proxy" ? providers.data?.proxy : providers.data?.rule;
    await Promise.all((list ?? []).map((p) => update(kind, p.name)));
  }

  async function updateGeo() {
    setGeo(true);
    try {
      await Inspect.upgradeGeoData();
      toastManager.add({ title: t.resourcesGeoDone(), type: "success" });
    } catch (err) {
      toastManager.add({ title: t.resourcesGeoFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setGeo(false);
    }
  }

  return (
    <>
      <PageHeader
        actions={
          <Button loading={geo} onClick={updateGeo} size="sm" variant="outline">
            <GlobeIcon aria-hidden="true" />
            {t.resourcesGeo()}
          </Button>
        }
        id="resources"
      />
      <CoreRequired title={t.resourcesOff()}>
        <PageBody>
          {providers.data && providers.data.proxy.length === 0 && providers.data.rule.length === 0 && (
            <Empty>
              <EmptyHeader>
                <EmptyTitle>{t.resourcesNone()}</EmptyTitle>
                <EmptyDescription>{t.resourcesNoneHint()}</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
          {(providers.data?.proxy.length ?? 0) > 0 && (
            <Section
              action={
                <Button onClick={() => updateAll("proxy")} size="xs" variant="ghost">
                  {t.resourcesUpdateAll()}
                </Button>
              }
              title={t.resourcesProxy()}
            >
              <div className="flex flex-col divide-y overflow-hidden rounded-xl border bg-card">
                {providers.data?.proxy.map((p) => (
                  <div className="flex items-center gap-4 px-4 py-3" key={p.name}>
                    <div className="flex min-w-0 flex-1 flex-col gap-0.5">
                      <span className="flex items-center gap-2 font-medium text-sm">
                        {p.name}
                        <Badge size="sm" variant="outline">
                          {p.vehicleType}
                        </Badge>
                      </span>
                      <span className="text-muted-foreground text-xs">
                        {t.proxiesCount(String(p.count))} · {t.profileUpdated(formatRelative(p.updatedAt))}
                      </span>
                    </div>
                    {p.usage && p.usage.total > 0 && <UsageMeter className="w-44" usage={p.usage} />}
                    <Button aria-label={t.profileRefresh()} loading={busy.has(`proxy:${p.name}`)} onClick={() => update("proxy", p.name)} size="icon-sm" variant="ghost">
                      <RefreshCwIcon aria-hidden="true" />
                    </Button>
                  </div>
                ))}
              </div>
            </Section>
          )}
          {(providers.data?.rule.length ?? 0) > 0 && (
            <Section
              action={
                <Button onClick={() => updateAll("rule")} size="xs" variant="ghost">
                  {t.resourcesUpdateAll()}
                </Button>
              }
              title={t.resourcesRule()}
            >
              <div className="flex flex-col divide-y overflow-hidden rounded-xl border bg-card">
                {providers.data?.rule.map((p) => (
                  <div className="flex items-center gap-4 px-4 py-3" key={p.name}>
                    <div className="flex min-w-0 flex-1 flex-col gap-0.5">
                      <span className="flex items-center gap-2 font-medium text-sm">
                        {p.name}
                        <Badge size="sm" variant="outline">
                          {p.behavior}
                        </Badge>
                        <Badge size="sm" variant="secondary">
                          {p.format}
                        </Badge>
                      </span>
                      <span className="text-muted-foreground text-xs">
                        {t.resourcesRules(String(p.count))} · {t.profileUpdated(formatRelative(p.updatedAt))}
                      </span>
                    </div>
                    <Button aria-label={t.profileRefresh()} loading={busy.has(`rule:${p.name}`)} onClick={() => update("rule", p.name)} size="icon-sm" variant="ghost">
                      <RefreshCwIcon aria-hidden="true" />
                    </Button>
                  </div>
                ))}
              </div>
            </Section>
          )}
        </PageBody>
      </CoreRequired>
    </>
  );
}
