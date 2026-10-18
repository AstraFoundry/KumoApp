import { GaugeIcon } from "lucide-react";
import { t } from "@/shared/i18n";
import { Button } from "@/shared/ui/button";
import { Tooltip, TooltipPopup, TooltipTrigger } from "@/shared/ui/tooltip";

/** Measures a group's nodes. */
export function TestDelayButton({ testing, disabled, onTest }: { testing: boolean; disabled?: boolean; onTest: () => void }) {
  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <Button aria-label={t.delayTest()} disabled={disabled} loading={testing} onClick={onTest} size="icon-sm" variant="ghost">
            <GaugeIcon aria-hidden="true" />
          </Button>
        }
      />
      <TooltipPopup>{t.delayTest()}</TooltipPopup>
    </Tooltip>
  );
}
