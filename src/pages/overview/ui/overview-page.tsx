import { CoreToggleButton } from "@/features/toggle-core";
import { ModeSwitch } from "@/features/switch-mode";
import { t } from "@/shared/i18n";
import { Card, CardHeader, CardPanel, CardTitle } from "@/shared/ui/card";
import { PageBody } from "@/shared/ui/page-body";
import { GroupPicker } from "@/widgets/group-picker";
import { PageHeader } from "@/widgets/page-header";
import { ProfileCard } from "@/widgets/profile-card";
import { StatusHero } from "@/widgets/status-hero";
import { TrafficChart } from "@/widgets/traffic-chart";

export function OverviewPage() {
  return (
    <>
      <PageHeader
        actions={
          <>
            <ModeSwitch />
            <CoreToggleButton />
          </>
        }
        id="overview"
      />
      <PageBody>
        <StatusHero />
        <div className="grid grid-cols-[minmax(0,1.6fr)_minmax(0,1fr)] gap-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">{t.trafficTitle()}</CardTitle>
            </CardHeader>
            <CardPanel className="pt-0">
              <TrafficChart />
            </CardPanel>
          </Card>
          <ProfileCard />
        </div>
        <GroupPicker />
      </PageBody>
    </>
  );
}
