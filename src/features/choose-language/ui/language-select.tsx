import { useEffect, useState } from "react";
import { Preferences } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { OptionSelect } from "@/shared/ui/option-select";

/** Chooses the app's language; "" follows the system. */
export function LanguageSelect({ className }: { className?: string }) {
  const [language, setLanguage] = useState<string>();
  useEffect(() => {
    void Preferences.get().then((p) => setLanguage(p.appLanguage ?? ""));
  }, []);
  if (language === undefined) return null;
  return (
    <OptionSelect
      className={className ?? "w-36"}
      label={t.settingsLanguage()}
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
