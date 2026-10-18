import { CloudDownloadIcon, CpuIcon, GlobeIcon, ShieldIcon } from "lucide-react";
import { isMyGo } from "mygo-runtime";
import { useEffect, useState } from "react";
import { App, events, type UpdateInfo } from "@/shared/api/mygo";
import { mark } from "@/shared/assets";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { useResource } from "@/shared/lib/use-resource";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { PageBody, Section } from "@/shared/ui/page-body";
import { Progress, ProgressIndicator, ProgressTrack } from "@/shared/ui/progress";
import { SettingRow, SettingsGroup } from "@/shared/ui/settings";
import { toastManager } from "@/shared/ui/toast";
import { PageHeader } from "@/widgets/page-header";
import { Showcase } from "@/widgets/showcase";

function ExternalLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <button className="text-brand-foreground text-sm underline-offset-4 outline-none hover:underline focus-visible:underline" onClick={() => App.openURL(href)} type="button">
      {children}
    </button>
  );
}

export function AboutPage() {
  const info = useResource(() => App.info(), []);
  const [update, setUpdate] = useState<UpdateInfo>();
  const [checking, setChecking] = useState(false);
  const [progress, setProgress] = useState<number>();

  useEffect(() => {
    if (!isMyGo()) return;
    return events.updateProgress.on(setProgress);
  }, []);

  async function check() {
    setChecking(true);
    try {
      const found = await App.checkUpdate();
      setUpdate(found);
      if (!found.available) toastManager.add({ title: t.aboutUpToDate(), type: "success" });
    } catch (err) {
      toastManager.add({ title: t.aboutUpdateFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setChecking(false);
    }
  }

  async function install() {
    try {
      setProgress(0);
      await App.installUpdate();
    } catch (err) {
      setProgress(undefined);
      toastManager.add({ title: t.aboutUpdateFailed(), description: errorMessage(err), type: "error" });
    }
  }

  const i = info.data;
  return (
    <>
      <PageHeader id="about" />
      <PageBody className="max-w-[44rem]">
        <Card className="overflow-hidden">
          <Showcase className="h-56 rounded-t-[calc(var(--radius-2xl)-1px)]" scene="connect" />
          <div className="flex items-center gap-3 px-5 py-4">
            <img alt="" className="size-11 rounded-xl" src={mark} />
            <div className="flex flex-col">
              <span className="font-semibold text-lg tracking-tight">Kumo</span>
              <span className="text-muted-foreground text-xs" data-selectable>
                {i ? t.aboutVersion(i.version) : " "}
              </span>
            </div>
            <p className="ms-auto max-w-56 text-right text-muted-foreground text-xs">{t.aboutTagline()}</p>
          </div>
        </Card>

        <SettingsGroup>
          <SettingRow
            description={
              update?.available
                ? t.aboutUpdateAvailable(update.version ?? "")
                : i && !i.canUpdate
                  ? t.aboutUpdatesDisabled()
                  : undefined
            }
            label={t.aboutUpdates()}
          >
            {progress !== undefined ? (
              <Progress className="w-40" value={progress * 100}>
                <ProgressTrack>
                  <ProgressIndicator />
                </ProgressTrack>
              </Progress>
            ) : update?.available ? (
              <Button onClick={install} size="sm">
                <CloudDownloadIcon aria-hidden="true" />
                {t.aboutInstall(update.version ?? "")}
              </Button>
            ) : i?.canUpdate ? (
              <Button loading={checking} onClick={check} size="sm" variant="outline">
                {t.aboutCheck()}
              </Button>
            ) : null}
          </SettingRow>
        </SettingsGroup>

        <Section title={t.aboutPrivacy()}>
          <SettingsGroup>
            {[
              { icon: CpuIcon, title: t.aboutPrivacyLocal(), body: t.aboutPrivacyLocalBody() },
              { icon: GlobeIcon, title: t.aboutPrivacyNetwork(), body: t.aboutPrivacyNetworkBody() },
              { icon: ShieldIcon, title: t.aboutPrivacyProxy(), body: t.aboutPrivacyProxyBody() },
            ].map((item) => (
              <div className="flex gap-3 px-4 py-3" key={item.title}>
                <item.icon aria-hidden="true" className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                <div className="flex flex-col gap-0.5">
                  <span className="font-medium text-sm">{item.title}</span>
                  <span className="text-muted-foreground text-xs leading-relaxed">{item.body}</span>
                </div>
              </div>
            ))}
          </SettingsGroup>
        </Section>

        {i && (
          <SettingsGroup>
            <SettingRow label={t.aboutTeam()}>
              <ExternalLink href={i.teamURL}>{i.team}</ExternalLink>
            </SettingRow>
            <SettingRow label={t.aboutSource()}>
              <ExternalLink href={i.repositoryURL}>{i.repository}</ExternalLink>
            </SettingRow>
            <SettingRow label={t.aboutFeedback()}>
              <ExternalLink href={i.issuesURL}>GitHub Issues</ExternalLink>
            </SettingRow>
            <SettingRow label={t.aboutLicense()}>
              <ExternalLink href={i.licenseURL}>{i.license}</ExternalLink>
            </SettingRow>
            <SettingRow description={t.aboutThanksBody()} label={t.aboutThanks()}>
              <span className="flex gap-3">
                <ExternalLink href={i.mihomoURL}>Mihomo</ExternalLink>
                <ExternalLink href="https://github.com/egoist/mygo">MyGo</ExternalLink>
                <ExternalLink href="https://coss.com/ui">coss ui</ExternalLink>
                <ExternalLink href="https://reactflow.dev">React Flow</ExternalLink>
              </span>
            </SettingRow>
          </SettingsGroup>
        )}
      </PageBody>
    </>
  );
}
