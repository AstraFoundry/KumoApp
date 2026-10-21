import { useState } from "react";
import { Proxies } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { toastManager } from "@/shared/ui/toast";

/** Measures groups' delays and keeps the latest results by node. */
export function useDelays() {
  const [delays, setDelays] = useState<Record<string, number>>({});
  const [testing, setTesting] = useState<Set<string>>(new Set());

  async function test(group: string) {
    setTesting((s) => new Set(s).add(group));
    try {
      const results = await Proxies.testGroup(group);
      setDelays((d) => ({ ...d, ...results }));
    } catch (err) {
      toastManager.add({ title: t.delayFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setTesting((s) => {
        const next = new Set(s);
        next.delete(group);
        return next;
      });
    }
  }

  return { delays, test, isTesting: (group: string) => testing.has(group) };
}
