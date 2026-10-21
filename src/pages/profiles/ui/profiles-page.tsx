import { CheckIcon, EllipsisIcon, PencilIcon, PlusIcon, RefreshCwIcon, TrashIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { KindBadge, UsageMeter, useProfiles } from "@/entities/profile";
import { EditProfileDialog } from "@/features/edit-profile";
import { ImportProfileDialog } from "@/features/import-profile";
import { Profiles, type Summary } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { formatRelative } from "@/shared/lib/format";
import { navigate, useLocation } from "@/shared/lib/router";
import { cn } from "@/shared/lib/utils";
import {
  AlertDialog,
  AlertDialogClose,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogPopup,
  AlertDialogTitle,
} from "@/shared/ui/alert-dialog";
import { Button } from "@/shared/ui/button";
import { Menu, MenuItem, MenuPopup, MenuSeparator, MenuTrigger } from "@/shared/ui/menu";
import { PageBody } from "@/shared/ui/page-body";
import { Skeleton } from "@/shared/ui/skeleton";
import { toastManager } from "@/shared/ui/toast";
import { PageHeader } from "@/widgets/page-header";

export function ProfilesPage() {
  const profiles = useProfiles();
  const { query } = useLocation();
  const [importing, setImporting] = useState(false);
  const [importURL, setImportURL] = useState<string>();
  const [editing, setEditing] = useState<Summary>();
  const [deleting, setDeleting] = useState<Summary>();
  const [refreshing, setRefreshing] = useState<Set<string>>(new Set());

  // A clash:// or kumo:// install link opens the import dialog prefilled.
  const linked = query.get("import");
  useEffect(() => {
    if (!linked) return;
    setImportURL(linked);
    setImporting(true);
    navigate("/profiles", true);
  }, [linked]);

  async function use(p: Summary) {
    if (p.isCurrent) return;
    try {
      await Profiles.select(p.id);
      await profiles.reload();
    } catch (err) {
      toastManager.add({ title: t.profileSwitchFailed(), description: errorMessage(err), type: "error" });
      await profiles.reload();
    }
  }

  async function refresh(p: Summary) {
    setRefreshing((s) => new Set(s).add(p.id));
    try {
      await Profiles.refresh(p.id);
      toastManager.add({ title: t.profileRefreshed(p.name), type: "success" });
    } catch (err) {
      toastManager.add({ title: t.profileRefreshFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setRefreshing((s) => {
        const next = new Set(s);
        next.delete(p.id);
        return next;
      });
    }
  }

  async function remove(p: Summary) {
    try {
      await Profiles.delete(p.id);
      setDeleting(undefined);
    } catch (err) {
      toastManager.add({ title: t.profileDeleteFailed(), description: errorMessage(err), type: "error" });
    }
  }

  return (
    <>
      <PageHeader
        actions={
          <Button
            onClick={() => {
              setImportURL(undefined);
              setImporting(true);
            }}
            size="sm"
          >
            <PlusIcon aria-hidden="true" />
            {t.importTitle()}
          </Button>
        }
        id="profiles"
      />
      <PageBody>
        <ul aria-label={t.pageProfiles()} className="flex flex-col divide-y overflow-hidden rounded-xl border bg-card shadow-xs/5">
          {!profiles.data &&
            [0, 1, 2].map((i) => (
              <li className="flex items-center gap-3 px-4 py-3.5" key={i}>
                <Skeleton className="size-4 rounded-full" />
                <Skeleton className="h-4 w-48" />
              </li>
            ))}
          {profiles.data?.map((p) => (
            <li className={cn("group/profile flex items-center gap-1 pe-2 transition-colors", !p.isCurrent && "hover:bg-accent/40")} key={p.id}>
              <button
                aria-pressed={p.isCurrent}
                className="flex min-w-0 flex-1 items-center gap-3 py-3 ps-4 text-left outline-none focus-visible:bg-accent/60"
                onClick={() => use(p)}
                type="button"
              >
                <span
                  aria-hidden="true"
                  className={cn(
                    "flex size-4 shrink-0 items-center justify-center rounded-full border transition-colors duration-150",
                    p.isCurrent ? "border-brand bg-brand text-white" : "border-input",
                  )}
                >
                  {p.isCurrent && <CheckIcon className="size-2.5" strokeWidth={3.5} />}
                </span>
                <span className="flex min-w-0 flex-1 flex-col gap-0.5">
                  <span className="flex items-center gap-2">
                    <span className={cn("truncate text-sm", p.isCurrent && "font-medium")}>{p.name}</span>
                    <KindBadge profile={p} />
                  </span>
                  <span className="truncate text-muted-foreground text-xs">
                    {p.sourceDescription}
                    {p.updatedAt ? ` · ${t.profileUpdated(formatRelative(p.updatedAt))}` : ""}
                    {p.kind === "remote" && p.autoUpdate && p.updateIntervalSeconds
                      ? ` · ${t.profileEvery(String(Math.round(p.updateIntervalSeconds / 3600)))}`
                      : ""}
                  </span>
                </span>
                {p.subscriptionUserInfo && p.subscriptionUserInfo.total > 0 && (
                  <UsageMeter className="w-44 shrink-0" usage={p.subscriptionUserInfo} />
                )}
              </button>
              {p.kind === "remote" && (
                <Button aria-label={t.profileRefresh()} loading={refreshing.has(p.id)} onClick={() => refresh(p)} size="icon-sm" variant="ghost">
                  <RefreshCwIcon aria-hidden="true" />
                </Button>
              )}
              <Menu>
                <MenuTrigger render={<Button aria-label={t.more()} size="icon-sm" variant="ghost" />}>
                  <EllipsisIcon aria-hidden="true" />
                </MenuTrigger>
                <MenuPopup align="end">
                  <MenuItem onClick={() => setEditing(p)}>
                    <PencilIcon aria-hidden="true" />
                    {t.edit()}
                  </MenuItem>
                  {p.id !== "default" && (
                    <>
                      <MenuSeparator />
                      <MenuItem onClick={() => setDeleting(p)} variant="destructive">
                        <TrashIcon aria-hidden="true" />
                        {t.delete()}
                      </MenuItem>
                    </>
                  )}
                </MenuPopup>
              </Menu>
            </li>
          ))}
        </ul>
        <p className="text-muted-foreground text-xs">{t.profilesFooter()}</p>
      </PageBody>
      <ImportProfileDialog initialURL={importURL} onOpenChange={setImporting} open={importing} />
      <EditProfileDialog onOpenChange={(open) => !open && setEditing(undefined)} profile={editing} />
      <AlertDialog onOpenChange={(open) => !open && setDeleting(undefined)} open={deleting !== undefined}>
        <AlertDialogPopup>
          <AlertDialogHeader>
            <AlertDialogTitle>{t.profileDeleteTitle(deleting?.name ?? "")}</AlertDialogTitle>
            <AlertDialogDescription>{deleting?.isCurrent ? t.profileDeleteCurrent() : t.profileDeleteBody()}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogClose render={<Button variant="ghost" />}>{t.cancel()}</AlertDialogClose>
            <Button onClick={() => deleting && remove(deleting)} variant="destructive">
              {t.delete()}
            </Button>
          </AlertDialogFooter>
        </AlertDialogPopup>
      </AlertDialog>
    </>
  );
}
