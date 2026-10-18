import { type ReactNode, useState } from "react";
import { setStatus } from "@/entities/status";
import { type PortConflict, Runtime } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import {
  AlertDialog,
  AlertDialogClose,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogPopup,
  AlertDialogTitle,
} from "@/shared/ui/alert-dialog";
import { Button } from "@/shared/ui/button";
import { toastManager } from "@/shared/ui/toast";

/**
 * When another app holds Kumo's ports, Kumo does not take them or fail
 * silently: it says who holds them and offers free ones.
 */
export function usePortConflicts(): { ask: (c: PortConflict[]) => void; dialog: ReactNode } {
  const [conflicts, setConflicts] = useState<PortConflict[]>([]);
  const [busy, setBusy] = useState(false);

  async function resolve() {
    setBusy(true);
    try {
      const outcome = await Runtime.start(true);
      setStatus(outcome.status);
      setConflicts([]);
    } catch (err) {
      toastManager.add({ title: t.coreStartFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setBusy(false);
    }
  }

  const dialog = (
    <AlertDialog onOpenChange={(open) => !open && setConflicts([])} open={conflicts.length > 0}>
      <AlertDialogPopup data-e2e="port-conflict">
        <AlertDialogHeader>
          <AlertDialogTitle>{t.portConflictTitle()}</AlertDialogTitle>
          <AlertDialogDescription render={<div />}>
            <ul className="flex flex-col gap-1.5">
              {conflicts.map((c) => (
                <li key={c.role}>
                  {t.portConflictLine(
                    c.role === "mixed" ? t.portRoleMixed() : t.portRoleController(),
                    String(c.port),
                    c.process ? `${c.process} (${c.pid})` : c.pid ? String(c.pid) : t.portHolderUnknown(),
                    String(c.suggested ?? ""),
                  )}
                </li>
              ))}
            </ul>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogClose render={<Button variant="ghost" />}>{t.cancel()}</AlertDialogClose>
          <Button data-e2e="use-free-ports" loading={busy} onClick={resolve}>
            {t.portConflictUseFree()}
          </Button>
        </AlertDialogFooter>
      </AlertDialogPopup>
    </AlertDialog>
  );
  return { ask: setConflicts, dialog };
}
