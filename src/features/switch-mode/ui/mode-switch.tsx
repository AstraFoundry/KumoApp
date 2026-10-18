import { setStatus, useStatus } from "@/entities/status";
import { type Mode, Runtime } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { toastManager } from "@/shared/ui/toast";
import { ToggleGroup, ToggleGroupItem } from "@/shared/ui/toggle-group";

const modes: { value: Mode; label: () => string }[] = [
  { value: "rule", label: () => t.modeRule() },
  { value: "global", label: () => t.modeGlobal() },
  { value: "direct", label: () => t.modeDirect() },
];

/** Switches the outbound mode; a running core changes at once. */
export function ModeSwitch() {
  const status = useStatus();
  async function change(next: Mode) {
    try {
      setStatus(await Runtime.setMode(next));
    } catch (err) {
      toastManager.add({ title: t.modeSwitchFailed(), description: errorMessage(err), type: "error" });
    }
  }
  return (
    <ToggleGroup
      aria-label={t.trayMode()}
      onValueChange={(value) => value[0] && change(value[0] as Mode)}
      size="sm"
      value={status ? [status.mode] : []}
      variant="outline"
    >
      {modes.map((m) => (
        <ToggleGroupItem className="px-3" key={m.value} value={m.value}>
          {m.label()}
        </ToggleGroupItem>
      ))}
    </ToggleGroup>
  );
}
