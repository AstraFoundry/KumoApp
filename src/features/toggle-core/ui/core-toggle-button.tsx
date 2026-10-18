import { PlayIcon, SquareIcon } from "lucide-react";
import { useState } from "react";
import { setStatus, useStatus } from "@/entities/status";
import { type PortConflict, Runtime } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { Button } from "@/shared/ui/button";
import { toastManager } from "@/shared/ui/toast";
import { usePortConflicts } from "../model/port-conflicts";

/** Starts or stops the core; a start blocked by busy ports asks first. */
export function CoreToggleButton({ size = "sm" }: { size?: "sm" | "default" | "lg" }) {
  const status = useStatus();
  const [busy, setBusy] = useState(false);
  const conflicts = usePortConflicts();
  const running = status?.state === "running";
  const transitioning = status?.state === "starting" || status?.state === "stopping";

  async function toggle() {
    setBusy(true);
    try {
      if (running) {
        setStatus(await Runtime.stop());
      } else {
        const outcome = await Runtime.start(false);
        setStatus(outcome.status);
        const found = outcome.conflicts.filter((c): c is PortConflict => c !== null);
        if (found.length > 0) conflicts.ask(found);
      }
    } catch (err) {
      toastManager.add({ title: running ? t.coreStopFailed() : t.coreStartFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <Button
        className="min-w-22"
        data-e2e="core-toggle"
        disabled={!status || transitioning}
        loading={busy}
        onClick={toggle}
        size={size}
        variant={running ? "outline" : "default"}
      >
        {running ? <SquareIcon aria-hidden="true" className="size-3 fill-current" /> : <PlayIcon aria-hidden="true" className="fill-current" />}
        {running ? t.coreStop() : t.coreStart()}
      </Button>
      {conflicts.dialog}
    </>
  );
}
