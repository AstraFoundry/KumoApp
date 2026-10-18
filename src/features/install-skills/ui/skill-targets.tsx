import { CheckIcon, UsersIcon } from "lucide-react";
import { useState } from "react";
import { App, type TargetStatus } from "@/shared/api/mygo";
import { agentLogos } from "@/shared/assets";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { useResource } from "@/shared/lib/use-resource";
import { cn } from "@/shared/lib/utils";
import { Button } from "@/shared/ui/button";
import { toastManager } from "@/shared/ui/toast";

const names: Record<string, string> = {
  claude: "Claude Code",
  codex: "Codex CLI",
  cursor: "Cursor",
  gemini: "Gemini CLI",
  agents: ".agents",
};

/**
 * The coding agents Kumo can teach its CLI: pick some, install the skill
 * into each. Agents found on this machine start picked.
 */
export function SkillTargets({ compact }: { compact?: boolean }) {
  const skills = useResource(() => App.skills("global"), []);
  const [picked, setPicked] = useState<Set<string>>();
  const [busy, setBusy] = useState(false);
  const targets = skills.data?.targets ?? [];
  const chosen = picked ?? new Set(targets.filter((x) => x.detected && !x.installed).map((x) => x.target));

  function toggle(target: string) {
    const next = new Set(chosen);
    if (next.has(target)) next.delete(target);
    else next.add(target);
    setPicked(next);
  }

  async function install() {
    setBusy(true);
    try {
      await App.installSkills([...chosen], false);
      toastManager.add({ title: t.skillsInstalled(String(chosen.size)), type: "success" });
      setPicked(new Set());
      await skills.reload();
    } catch (err) {
      toastManager.add({ title: t.skillsFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setBusy(false);
    }
  }

  async function remove(target: string) {
    try {
      await App.uninstallSkills([target]);
      await skills.reload();
    } catch (err) {
      toastManager.add({ title: t.skillsFailed(), description: errorMessage(err), type: "error" });
    }
  }

  return (
    <div className="flex flex-col gap-3">
      <div className={cn("grid gap-2", compact ? "grid-cols-2" : "grid-cols-[repeat(auto-fill,minmax(14rem,1fr))]")}>
        {targets.map((target: TargetStatus) => {
          const on = chosen.has(target.target);
          const logo = agentLogos[target.target];
          return (
            <div
              className={cn(
                "flex items-center gap-2.5 rounded-xl border bg-card px-3 py-2.5 transition-[border-color,background-color] duration-150",
                on && !target.installed && "border-brand/40 bg-brand-soft",
              )}
              key={target.target}
            >
              {logo ? (
                <img alt="" className="size-6 rounded-md" src={logo} />
              ) : (
                <span className="flex size-6 items-center justify-center rounded-md bg-muted">
                  <UsersIcon aria-hidden="true" className="size-3.5" />
                </span>
              )}
              <div className="flex min-w-0 flex-1 flex-col">
                <span className="font-medium text-sm">{names[target.target] ?? target.target}</span>
                <span className="text-muted-foreground text-xs">
                  {target.installed ? (target.upToDate ? t.skillsUpToDate() : t.skillsOutdated()) : target.detected ? t.skillsDetected() : t.skillsNotDetected()}
                </span>
              </div>
              {target.installed ? (
                <Button onClick={() => remove(target.target)} size="xs" variant="ghost">
                  {t.cliRemove()}
                </Button>
              ) : (
                <button
                  aria-label={names[target.target]}
                  aria-pressed={on}
                  className={cn(
                    "flex size-5 items-center justify-center rounded-md border transition-colors outline-none focus-visible:ring-2 focus-visible:ring-ring",
                    on ? "border-brand bg-brand text-white" : "border-input bg-background",
                  )}
                  onClick={() => toggle(target.target)}
                  type="button"
                >
                  {on && <CheckIcon aria-hidden="true" className="size-3" strokeWidth={3} />}
                </button>
              )}
            </div>
          );
        })}
      </div>
      <Button className="self-start" disabled={chosen.size === 0} loading={busy} onClick={install} size="sm">
        {t.skillsInstall(String(chosen.size))}
      </Button>
    </div>
  );
}
