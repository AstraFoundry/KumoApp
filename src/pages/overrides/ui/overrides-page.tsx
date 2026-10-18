import { ArrowDownIcon, ArrowUpIcon, PencilIcon, PlusIcon, TrashIcon } from "lucide-react";
import { useState } from "react";
import { OverrideDialog } from "@/features/edit-override";
import { type Override, Overrides } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { formatRelative } from "@/shared/lib/format";
import { useResource } from "@/shared/lib/use-resource";
import { cn } from "@/shared/lib/utils";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { PageBody } from "@/shared/ui/page-body";
import { Switch } from "@/shared/ui/switch";
import { toastManager } from "@/shared/ui/toast";
import { PageHeader } from "@/widgets/page-header";

export function OverridesPage() {
  const list = useResource(() => Overrides.list(), []);
  const [editing, setEditing] = useState<string>();

  async function run(action: () => Promise<void>) {
    try {
      await action();
    } catch (err) {
      toastManager.add({ title: t.overrideFailed(), description: errorMessage(err), type: "error" });
    } finally {
      await list.reload();
    }
  }

  function move(items: Override[], index: number, by: number) {
    const next = [...items];
    const [item] = next.splice(index, 1);
    if (!item) return;
    next.splice(index + by, 0, item);
    list.set(next);
    void run(() => Overrides.reorder(next.map((o) => o.id)));
  }

  const items = list.data ?? [];

  return (
    <>
      <PageHeader
        actions={
          <Button onClick={() => setEditing("new")} size="sm">
            <PlusIcon aria-hidden="true" />
            {t.overrideNew()}
          </Button>
        }
        id="overrides"
      />
      <PageBody>
        {list.data && items.length === 0 ? (
          <Empty>
            <EmptyHeader>
              <EmptyTitle>{t.overridesNone()}</EmptyTitle>
              <EmptyDescription>{t.overridesNoneHint()}</EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <>
            <p className="text-muted-foreground text-sm">{t.overridesIntro()}</p>
            <ol className="flex flex-col divide-y overflow-hidden rounded-xl border bg-card">
              {items.map((o, index) => (
                <li className={cn("flex items-center gap-3 px-4 py-2.5", o.disabled && "opacity-60")} key={o.id}>
                  <span className="w-5 text-center text-muted-foreground text-xs tabular-nums">{index + 1}</span>
                  <div className="flex min-w-0 flex-1 flex-col gap-0.5">
                    <span className="flex items-center gap-2 text-sm">
                      <span className="truncate font-medium">{o.name}</span>
                      <Badge size="sm" variant="outline">
                        {o.format === "javascript" ? "JS" : "YAML"}
                      </Badge>
                      {o.kind === "remote" && (
                        <Badge size="sm" variant="info">
                          {t.profileKindRemote()}
                        </Badge>
                      )}
                    </span>
                    <span className="text-muted-foreground text-xs">{t.profileUpdated(formatRelative(o.updatedAt))}</span>
                  </div>
                  <div className="flex items-center">
                    <Button aria-label={t.overrideUp()} disabled={index === 0} onClick={() => move(items, index, -1)} size="icon-xs" variant="ghost">
                      <ArrowUpIcon aria-hidden="true" />
                    </Button>
                    <Button aria-label={t.overrideDown()} disabled={index === items.length - 1} onClick={() => move(items, index, 1)} size="icon-xs" variant="ghost">
                      <ArrowDownIcon aria-hidden="true" />
                    </Button>
                    <Button aria-label={t.edit()} onClick={() => setEditing(o.id)} size="icon-xs" variant="ghost">
                      <PencilIcon aria-hidden="true" />
                    </Button>
                    <Button aria-label={t.delete()} onClick={() => run(() => Overrides.delete(o.id))} size="icon-xs" variant="ghost">
                      <TrashIcon aria-hidden="true" />
                    </Button>
                  </div>
                  <Switch aria-label={t.overrideApplied()} checked={!o.disabled} onCheckedChange={(on) => run(() => Overrides.setDisabled(o.id, !on))} />
                </li>
              ))}
            </ol>
          </>
        )}
      </PageBody>
      <OverrideDialog
        id={editing}
        onOpenChange={(open) => {
          if (!open) {
            setEditing(undefined);
            void list.reload();
          }
        }}
      />
    </>
  );
}
