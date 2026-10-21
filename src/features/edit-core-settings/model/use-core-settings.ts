import { useEffect, useMemo, useState } from "react";
import { setStatus } from "@/entities/status";
import { type CoreSettings, type RuntimeSettings, Settings } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { toastManager } from "@/shared/ui/toast";

export interface CoreSettingsForm {
  /** What is stored, with the cores Kumo found. */
  stored: CoreSettings | undefined;
  /** The settings as edited. */
  draft: RuntimeSettings | undefined;
  /** Changes part of the draft. */
  edit: (fn: (draft: RuntimeSettings) => RuntimeSettings) => void;
  dirty: boolean;
  applying: boolean;
  apply: () => Promise<void>;
  discard: () => void;
  reload: () => Promise<void>;
}

/**
 * Edits the runtime settings as a draft: nothing reaches Mihomo until the
 * user applies, so half-typed ports and lists never run.
 */
export function useCoreSettings(): CoreSettingsForm {
  const [stored, setStored] = useState<CoreSettings>();
  const [draft, setDraft] = useState<RuntimeSettings>();
  const [applying, setApplying] = useState(false);

  async function reload() {
    const settings = await Settings.get();
    setStored(settings);
    setDraft(settings.settings);
  }

  // biome-ignore lint/correctness/useExhaustiveDependencies: load once
  useEffect(() => {
    void reload();
  }, []);

  const dirty = useMemo(() => stored !== undefined && draft !== undefined && JSON.stringify(stored.settings) !== JSON.stringify(draft), [stored, draft]);

  async function apply() {
    if (!draft) return;
    setApplying(true);
    try {
      setStatus(await Settings.update(draft));
      await reload();
      toastManager.add({ title: t.settingsApplied(), type: "success" });
    } catch (err) {
      toastManager.add({ title: t.settingsApplyFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setApplying(false);
    }
  }

  return {
    stored,
    draft,
    edit: (fn) => setDraft((d) => (d ? fn(structuredClone(d)) : d)),
    dirty,
    applying,
    apply,
    discard: () => setDraft(stored?.settings),
    reload,
  };
}
