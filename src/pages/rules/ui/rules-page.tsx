import { useMemo, useState } from "react";
import { Inspect, type Rule } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { useResource } from "@/shared/lib/use-resource";
import { cn } from "@/shared/lib/utils";
import { Badge } from "@/shared/ui/badge";
import { SearchField } from "@/shared/ui/search-field";
import { Switch } from "@/shared/ui/switch";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/shared/ui/table";
import { toastManager } from "@/shared/ui/toast";
import { CoreRequired } from "@/widgets/core-required";
import { PageHeader } from "@/widgets/page-header";

/** The running rules in order; the first match decides. */
export function RulesPage() {
  const rules = useResource(() => Inspect.rules(), []);
  const [query, setQuery] = useState("");

  const rows = useMemo(() => {
    const q = query.trim().toLowerCase();
    const list = rules.data ?? [];
    return q ? list.filter((r) => [r.type, r.payload, r.proxy].some((v) => v.toLowerCase().includes(q))) : list;
  }, [rules.data, query]);

  async function toggle(rule: Rule, enabled: boolean) {
    rules.set((rules.data ?? []).map((r) => (r.index === rule.index ? { ...r, enabled } : r)));
    try {
      await Inspect.setRuleEnabled(rule.index, enabled);
    } catch (err) {
      toastManager.add({ title: t.rulesToggleFailed(), description: errorMessage(err), type: "error" });
      await rules.reload();
    }
  }

  return (
    <>
      <PageHeader actions={<SearchField label={t.rulesSearch()} onChange={setQuery} value={query} />} id="rules" />
      <CoreRequired title={t.rulesOff()}>
        <div className="border-b px-5 py-2 text-muted-foreground text-xs">{t.rulesHint(String(rules.data?.length ?? 0))}</div>
        <div className="min-h-0 flex-1 overflow-auto">
          <Table className="table-fixed text-xs">
            <TableHeader className="sticky top-0 z-10 bg-background">
              <TableRow>
                <TableHead className="w-14 ps-5">#</TableHead>
                <TableHead className="w-40">{t.rulesType()}</TableHead>
                <TableHead>{t.rulesPayload()}</TableHead>
                <TableHead className="w-44">{t.rulesPolicy()}</TableHead>
                <TableHead className="w-24 text-right">{t.rulesHits()}</TableHead>
                <TableHead className="w-16 pe-5 text-right">{t.rulesOn()}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((r) => (
                <TableRow className={cn(!r.enabled && "opacity-50")} key={r.index}>
                  <TableCell className="ps-5 text-muted-foreground tabular-nums">{r.index}</TableCell>
                  <TableCell>
                    <Badge size="sm" variant="outline">
                      {r.type}
                    </Badge>
                  </TableCell>
                  <TableCell className="truncate font-mono" data-selectable title={r.payload}>
                    {r.payload || "—"}
                    {r.size > 0 && <span className="ms-2 text-muted-foreground">{t.rulesSize(String(r.size))}</span>}
                  </TableCell>
                  <TableCell className="truncate">{r.proxy}</TableCell>
                  <TableCell className="text-right text-muted-foreground tabular-nums">{r.hitCount || ""}</TableCell>
                  <TableCell className="pe-5 text-right">
                    <Switch aria-label={t.rulesOn()} checked={r.enabled} onCheckedChange={(v) => toggle(r, v)} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </CoreRequired>
    </>
  );
}
