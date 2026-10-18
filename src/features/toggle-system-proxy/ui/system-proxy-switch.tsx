import { useState } from "react";
import { setStatus, useStatus } from "@/entities/status";
import { SystemProxy } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { Switch } from "@/shared/ui/switch";
import { toastManager } from "@/shared/ui/toast";

/**
 * The system proxy switch reflects what the user wants. While Kumo is
 * stopped the wish is kept, and applied on the next start.
 */
export function SystemProxySwitch({ id }: { id?: string }) {
  const status = useStatus();
  const [pending, setPending] = useState<boolean>();
  const checked = pending ?? status?.systemProxy.enabled ?? false;

  async function change(next: boolean) {
    setPending(next);
    try {
      setStatus(await SystemProxy.setEnabled(next));
    } catch (err) {
      toastManager.add({ title: t.systemProxyFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setPending(undefined);
    }
  }

  return <Switch aria-label={t.pageSystemProxy()} checked={checked} disabled={!status} id={id} onCheckedChange={change} />;
}
