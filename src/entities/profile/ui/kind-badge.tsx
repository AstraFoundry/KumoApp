import type { Summary } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { Badge } from "@/shared/ui/badge";

/** Where a profile comes from: a subscription, a file or Kumo itself. */
export function KindBadge({ profile }: { profile: Summary }) {
  const label = profile.isSubStoreManaged
    ? "Sub-Store"
    : profile.kind === "remote"
      ? t.profileKindRemote()
      : profile.kind === "inline"
        ? t.profileKindDefault()
        : t.profileKindLocal();
  return (
    <Badge size="sm" variant={profile.kind === "remote" ? "info" : "secondary"}>
      {label}
    </Badge>
  );
}
