import { isMyGo } from "mygo-runtime";
import { ShieldAlertIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { setStatus, useStatus } from "@/entities/status";
import { events, SystemProxy } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { Alert, AlertAction, AlertDescription, AlertTitle } from "@/shared/ui/alert";
import { Button } from "@/shared/ui/button";
import { toastManager } from "@/shared/ui/toast";

/**
 * Shown while another app holds the system proxy Kumo had set. Kumo does
 * not fight over it: the user decides whether to take it back.
 */
export function ContentionAlert() {
  const status = useStatus();
  const [holder, setHolder] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!isMyGo()) return;
    return events.contention.on((c) => setHolder(c.holder ?? ""));
  }, []);

  if (status?.systemProxy.ownership !== "takenOver") return null;

  async function reclaim() {
    setBusy(true);
    try {
      setStatus(await SystemProxy.reclaim());
    } catch (err) {
      toastManager.add({ title: t.systemProxyFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setBusy(false);
    }
  }

  return (
    <Alert variant="warning">
      <ShieldAlertIcon aria-hidden="true" />
      <AlertTitle>{t.contentionTitle()}</AlertTitle>
      <AlertDescription>{holder ? t.contentionBodyHolder(holder) : t.contentionBody()}</AlertDescription>
      <AlertAction>
        <Button loading={busy} onClick={reclaim} size="sm" variant="outline">
          {t.contentionReclaim()}
        </Button>
      </AlertAction>
    </Alert>
  );
}
