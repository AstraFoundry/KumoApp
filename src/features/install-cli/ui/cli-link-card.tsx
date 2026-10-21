import { CheckCircle2Icon, TerminalIcon } from "lucide-react";
import { useState } from "react";
import { App, type ClilinkStatus } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { useResource } from "@/shared/lib/use-resource";
import { Button } from "@/shared/ui/button";
import { toastManager } from "@/shared/ui/toast";

/** Puts kumo on PATH, or shows where it is. */
export function CliLinkCard() {
  const status = useResource(() => App.setup().then((s) => s.cli), []);
  const [busy, setBusy] = useState(false);
  const cli: ClilinkStatus | undefined = status.data;

  async function run(action: () => Promise<ClilinkStatus>) {
    setBusy(true);
    try {
      status.set(await action());
    } catch (err) {
      toastManager.add({ title: t.cliFailed(), description: errorMessage(err), type: "error" });
      await status.reload();
    } finally {
      setBusy(false);
    }
  }

  if (!cli) return <div className="h-16" />;
  const installed = cli.installed && cli.managed;
  return (
    <div className="flex items-center gap-3 rounded-xl border bg-card px-4 py-3">
      <span className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted">
        {installed ? <CheckCircle2Icon aria-hidden="true" className="size-4 text-success" /> : <TerminalIcon aria-hidden="true" className="size-4" />}
      </span>
      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <span className="font-medium text-sm">{installed ? t.cliInstalled() : cli.installed ? t.cliForeign() : t.cliNotInstalled()}</span>
        <span className="truncate font-mono text-muted-foreground text-xs" data-selectable>
          {cli.path}
        </span>
      </div>
      {cli.supported &&
        (installed ? (
          <Button loading={busy} onClick={() => run(App.uninstallCLI)} size="sm" variant="ghost">
            {t.cliRemove()}
          </Button>
        ) : (
          <Button disabled={cli.installed} loading={busy} onClick={() => run(App.installCLI)} size="sm">
            {t.cliInstall()}
          </Button>
        ))}
    </div>
  );
}
