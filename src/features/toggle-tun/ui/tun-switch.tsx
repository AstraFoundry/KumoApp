import { useState } from "react";
import { setStatus, useStatus } from "@/entities/status";
import { Settings } from "@/shared/api/mygo";
import { routes } from "@/shared/config/routes";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { navigate } from "@/shared/lib/router";
import { Switch } from "@/shared/ui/switch";
import { toastManager } from "@/shared/ui/toast";

/** Turns TUN on or off; without permission it points to where to grant it. */
export function TunSwitch({ id }: { id?: string }) {
  const status = useStatus();
  const [pending, setPending] = useState<boolean>();
  const checked = pending ?? status?.tun ?? false;

  async function change(next: boolean) {
    setPending(next);
    try {
      setStatus(await Settings.setTUN(next));
    } catch (err) {
      const message = errorMessage(err);
      toastManager.add({
        title: t.tunFailed(),
        description: message,
        type: "error",
        actionProps: message.includes("permission") ? { children: t.tunOpenSettings(), onClick: () => navigate(routes.tun) } : undefined,
      });
    } finally {
      setPending(undefined);
    }
  }

  return <Switch aria-label={t.pageTun()} checked={checked} disabled={!status} id={id} onCheckedChange={change} />;
}
