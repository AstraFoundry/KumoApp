import { FolderOpenIcon } from "lucide-react";
import { setTheme } from "@/shared/lib/theme";
import { LanguageSelect } from "@/features/choose-language";
import { App, Preferences, type Preferences as Prefs, type Theme } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { navigate } from "@/shared/lib/router";
import { useResource } from "@/shared/lib/use-resource";
import { Button } from "@/shared/ui/button";
import { OptionSelect } from "@/shared/ui/option-select";
import { PageBody, Section } from "@/shared/ui/page-body";
import { SettingRow, SettingsGroup } from "@/shared/ui/settings";
import { Switch } from "@/shared/ui/switch";
import { PageHeader } from "@/widgets/page-header";

export function SettingsPage() {
  const prefs = useResource(() => Preferences.get(), []);
  const p = prefs.data;

  async function save(patch: Partial<Prefs>) {
    if (!p) return;
    const next = { ...p, ...patch };
    prefs.set(next);
    await Preferences.save(next);
  }

  return (
    <>
      <PageHeader id="settings" />
      <PageBody>
        {p && (
          <>
            <Section title={t.settingsGeneral()}>
              <SettingsGroup>
                <SettingRow label={t.settingsLaunchAtLogin()}>
                  <Switch checked={p.launchAtLogin} onCheckedChange={(launchAtLogin) => save({ launchAtLogin })} />
                </SettingRow>
                <SettingRow description={t.settingsStartCoreHint()} label={t.settingsStartCore()}>
                  <Switch checked={p.startCoreOnLaunch} onCheckedChange={(startCoreOnLaunch) => save({ startCoreOnLaunch })} />
                </SettingRow>
                <SettingRow description={t.settingsQuitOnCloseHint()} label={t.settingsQuitOnClose()}>
                  <Switch checked={p.quitOnLastWindowClose} onCheckedChange={(quitOnLastWindowClose) => save({ quitOnLastWindowClose })} />
                </SettingRow>
                <SettingRow description={t.settingsHideTrayHint()} label={t.settingsHideTray()}>
                  <Switch checked={p.hideMenuBarIcon} onCheckedChange={(hideMenuBarIcon) => save({ hideMenuBarIcon })} />
                </SettingRow>
              </SettingsGroup>
            </Section>
            <Section title={t.settingsAppearance()}>
              <SettingsGroup>
                <SettingRow label={t.settingsLanguage()}>
                  <LanguageSelect />
                </SettingRow>
                <SettingRow label={t.settingsTheme()}>
                  <OptionSelect
                    className="w-36"
                    label={t.settingsTheme()}
                    onChange={(theme: Theme) => {
                      setTheme(theme);
                      void save({ theme });
                    }}
                    options={[
                      { value: "system" as Theme, label: t.settingsThemeSystem() },
                      { value: "light" as Theme, label: t.settingsThemeLight() },
                      { value: "dark" as Theme, label: t.settingsThemeDark() },
                    ]}
                    value={p.theme}
                  />
                </SettingRow>
              </SettingsGroup>
            </Section>
            <Section title={t.settingsUpdates()}>
              <SettingsGroup>
                <SettingRow label={t.settingsAutoCheck()}>
                  <Switch checked={p.autoCheckUpdates} onCheckedChange={(autoCheckUpdates) => save({ autoCheckUpdates })} />
                </SettingRow>
              </SettingsGroup>
            </Section>
            <Section title={t.settingsAdvanced()}>
              <SettingsGroup>
                <SettingRow description={t.settingsSetupHint()} label={t.settingsSetup()}>
                  <Button
                    onClick={async () => {
                      await App.restartSetup();
                      navigate("/setup");
                    }}
                    size="sm"
                    variant="outline"
                  >
                    {t.settingsSetupRun()}
                  </Button>
                </SettingRow>
                <SettingRow label={t.settingsData()}>
                  <Button onClick={() => App.reveal("data")} size="sm" variant="outline">
                    <FolderOpenIcon aria-hidden="true" />
                    {t.revealInFolder()}
                  </Button>
                </SettingRow>
              </SettingsGroup>
            </Section>
          </>
        )}
      </PageBody>
    </>
  );
}
