import { BotIcon, CheckIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { App, type Guide } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { Button } from "@/shared/ui/button";
import { Menu, MenuGroup, MenuGroupLabel, MenuItem, MenuPopup, MenuTrigger } from "@/shared/ui/menu";
import { toastManager } from "@/shared/ui/toast";
import { Tooltip, TooltipPopup, TooltipTrigger } from "@/shared/ui/tooltip";
import { intentLabel } from "../model/intents";

/**
 * A quiet icon button that copies a ready-made prompt for a coding agent:
 * how to do one of the page's tasks with the kumo CLI. It is an aside to
 * the page, so it stays small and unlabeled until hovered.
 */
// Guides are fixed for the app's life: load them once for every header.
let guides: Promise<Guide[]> | undefined;
function loadGuides() {
  guides ??= App.guides().catch((err) => {
    guides = undefined;
    throw err;
  });
  return guides;
}

export function AgentPromptMenu({ page }: { page: string }) {
  const [guide, setGuide] = useState<Guide>();
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    let live = true;
    loadGuides()
      .then((all) => live && setGuide(all.find((g) => g.page === page)))
      .catch(() => {});
    return () => {
      live = false;
    };
  }, [page]);

  useEffect(() => {
    if (!copied) return;
    const timer = window.setTimeout(() => setCopied(false), 1600);
    return () => window.clearTimeout(timer);
  }, [copied]);

  if (guide && guide.intents.length === 0) return null;

  async function copy(intent: string) {
    try {
      await App.copyPrompt(page, intent);
      setCopied(true);
      toastManager.add({ title: t.agentPromptCopied(), description: t.agentPromptCopiedHint(), type: "success" });
    } catch (err) {
      toastManager.add({ title: errorMessage(err), type: "error" });
    }
  }

  return (
    <Menu>
      <Tooltip>
        <TooltipTrigger
          render={
            <MenuTrigger
              render={
                <Button aria-label={t.agentPromptButton()} className="text-muted-foreground" disabled={!guide} size="icon-sm" variant="ghost">
                  {copied ? <CheckIcon aria-hidden="true" /> : <BotIcon aria-hidden="true" />}
                </Button>
              }
            />
          }
        />
        <TooltipPopup>{t.agentPromptButton()}</TooltipPopup>
      </Tooltip>
      <MenuPopup align="end" className="min-w-64">
        <MenuGroup>
          <MenuGroupLabel>{t.agentPromptMenuTitle()}</MenuGroupLabel>
          {guide?.intents.map((intent) => (
            <MenuItem key={intent.id} onClick={() => copy(intent.id)}>
              {intentLabel(page, intent.id, intent.goal)}
            </MenuItem>
          ))}
        </MenuGroup>
      </MenuPopup>
    </Menu>
  );
}
