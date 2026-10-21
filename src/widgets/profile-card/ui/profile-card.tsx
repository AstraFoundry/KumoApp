import { ArrowRightIcon, RefreshCwIcon } from "lucide-react";
import { useState } from "react";
import { KindBadge, UsageMeter } from "@/entities/profile";
import { useStatus } from "@/entities/status";
import { Profiles } from "@/shared/api/mygo";
import { routes } from "@/shared/config/routes";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { formatRelative } from "@/shared/lib/format";
import { Link } from "@/shared/lib/router";
import { Button } from "@/shared/ui/button";
import { Card, CardAction, CardDescription, CardHeader, CardPanel, CardTitle } from "@/shared/ui/card";
import { toastManager } from "@/shared/ui/toast";

/** The current profile: where it comes from, its usage and freshness. */
export function ProfileCard() {
  const status = useStatus();
  const [refreshing, setRefreshing] = useState(false);
  const profile = status?.profile;
  if (!profile) return <Card className="min-h-40" />;

  async function refresh() {
    if (!profile) return;
    setRefreshing(true);
    try {
      await Profiles.refresh(profile.id);
      toastManager.add({ title: t.profileRefreshed(profile.name), type: "success" });
    } catch (err) {
      toastManager.add({ title: t.profileRefreshFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setRefreshing(false);
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <span className="truncate">{profile.name}</span>
          <KindBadge profile={profile} />
        </CardTitle>
        <CardDescription className="truncate">
          {profile.sourceDescription} · {t.profileUpdated(formatRelative(profile.updatedAt))}
        </CardDescription>
        <CardAction className="flex gap-1">
          {profile.kind === "remote" && (
            <Button aria-label={t.profileRefresh()} loading={refreshing} onClick={refresh} size="icon-sm" variant="ghost">
              <RefreshCwIcon aria-hidden="true" />
            </Button>
          )}
          <Button aria-label={t.pageProfiles()} render={<Link href={routes.profiles} />} size="icon-sm" variant="ghost">
            <ArrowRightIcon aria-hidden="true" />
          </Button>
        </CardAction>
      </CardHeader>
      <CardPanel className="pt-0">
        {profile.subscriptionUserInfo && profile.subscriptionUserInfo.total > 0 ? (
          <UsageMeter usage={profile.subscriptionUserInfo} />
        ) : (
          <p className="text-muted-foreground text-xs">{profile.kind === "inline" ? t.profileDefaultHint() : t.profileNoUsage()}</p>
        )}
      </CardPanel>
    </Card>
  );
}
