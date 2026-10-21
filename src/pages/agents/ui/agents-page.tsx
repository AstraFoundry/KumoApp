import { CopyIcon } from "lucide-react";
import { CliLinkCard } from "@/features/install-cli";
import { SkillTargets } from "@/features/install-skills";
import { App, type Guide } from "@/shared/api/mygo";
import { pageTitle } from "@/shared/config/page-titles";
import { routes } from "@/shared/config/routes";
import { t } from "@/shared/i18n";
import { useResource } from "@/shared/lib/use-resource";
import { Button } from "@/shared/ui/button";
import { PageBody, Section } from "@/shared/ui/page-body";
import { toastManager } from "@/shared/ui/toast";
import { PageHeader } from "@/widgets/page-header";
import { intentLabel } from "@/features/copy-agent-prompt";

const pages = Object.keys(routes).filter((id) => id !== "setup" && id !== "about") as (keyof typeof routes)[];

export function AgentsPage() {
  const guides = useResource(() => Promise.all(pages.map((p) => App.guide(p))), []);
  async function copy(page: string, intent: string) {
    await App.copyPrompt(page, intent);
    toastManager.add({ title: t.agentPromptCopied(), description: t.agentPromptCopiedHint(), type: "success" });
  }
  return (
    <>
      <PageHeader id="agents" />
      <PageBody>
        <p className="max-w-prose text-muted-foreground text-sm leading-relaxed">{t.agentsIntro()}</p>
        <Section description={t.agentsSkillHint()} title={t.agentsSkill()}>
          <SkillTargets />
        </Section>
        <Section description={t.agentsCliHint()} title={t.agentsCli()}>
          <CliLinkCard />
        </Section>
        <Section description={t.agentsPromptsHint()} title={t.agentsPrompts()}>
          <div className="flex flex-col divide-y overflow-hidden rounded-xl border bg-card">
            {(guides.data ?? [])
              .filter((g: Guide) => g.intents.length > 0)
              .map((g: Guide) => (
                <div className="flex gap-4 px-4 py-3" key={g.page}>
                  <span className="w-28 shrink-0 pt-1 font-medium text-sm">{pageTitle(g.page as keyof typeof routes)}</span>
                  <div className="flex flex-wrap gap-1.5">
                    {g.intents.map((intent) => (
                      <Button key={intent.id} onClick={() => copy(g.page, intent.id)} size="xs" variant="outline">
                        <CopyIcon aria-hidden="true" />
                        {intentLabel(g.page, intent.id, intent.goal)}
                      </Button>
                    ))}
                  </div>
                </div>
              ))}
          </div>
        </Section>
      </PageBody>
    </>
  );
}
