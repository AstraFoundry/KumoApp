import { GlobeIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { Preferences } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { OptionSelect } from "@/shared/ui/option-select";

/** Chooses the app's language; "" follows the system. Quiet drops the field chrome for places where it is a secondary utility. */
export function LanguageSelect({ className, quiet }: { className?: string; quiet?: boolean }) {
  const [language, setLanguage] = useState<string>();
  useEffect(() => {
    void Preferences.get().then((p) => setLanguage(p.appLanguage ?? ""));
  }, []);
  if (language === undefined) return null;
  return (
    <OptionSelect
      className={className ?? (quiet ? "w-auto min-w-0 justify-start border-transparent bg-transparent text-muted-foreground shadow-none before:hidden hover:bg-accent dark:bg-transparent" : "w-36")}
      label={t.settingsLanguage()}
      leading={quiet ? <GlobeIcon aria-hidden="true" className="-ms-0.5" /> : undefined}
      onChange={async (next) => {
        const value = next === "system" ? "" : next;
        setLanguage(value);
        const prefs = await Preferences.get();
        await Preferences.save({ ...prefs, appLanguage: value });
      }}
      options={[
        { value: "system", label: t.settingsLanguageSystem() },
        { value: "en", label: "English" },
        { value: "zh-Hans", label: "简体中文" },
      ]}
      value={language || "system"}
    />
  );
}
