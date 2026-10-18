import { useState } from "react";
import { Proxies } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { toastManager } from "@/shared/ui/toast";

/**
 * Selects a node in a group. The choice shows at once and rolls back if
 * the core refuses it.
 */
export function useSelectNode(onSelected?: () => void) {
  const [optimistic, setOptimistic] = useState<Record<string, string>>({});

  async function select(group: string, node: string) {
    setOptimistic((o) => ({ ...o, [group]: node }));
    try {
      await Proxies.select(group, node);
      onSelected?.();
    } catch (err) {
      toastManager.add({ title: t.selectFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setOptimistic(({ [group]: _, ...rest }) => rest);
    }
  }

  return { select, selected: (group: string, now: string | undefined) => optimistic[group] ?? now };
}
