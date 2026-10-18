import { PauseIcon, PlayIcon, XIcon } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { type Connection, Inspect } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { formatBytes, formatSpeed, formatUptime } from "@/shared/lib/format";
import { useResource } from "@/shared/lib/use-resource";
import { cn } from "@/shared/lib/utils";
import { Button } from "@/shared/ui/button";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { SearchField } from "@/shared/ui/search-field";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/shared/ui/table";
import { toastManager } from "@/shared/ui/toast";
import { Tooltip, TooltipPopup, TooltipTrigger } from "@/shared/ui/tooltip";
import { CoreRequired } from "@/widgets/core-required";
import { PageHeader } from "@/widgets/page-header";

const icons = new Map<string, Promise<string>>();

function useProcessIcon(path: string | undefined): string {
  const [icon, setIcon] = useState("");
  useEffect(() => {
    if (!path) return;
    let live = true;
    let pending = icons.get(path);
    if (!pending) {
      pending = Inspect.processIcon(path).catch(() => "");
      icons.set(path, pending);
    }
    void pending.then((url) => live && setIcon(url));
    return () => {
      live = false;
    };
  }, [path]);
  return icon;
}

function ProcessCell({ connection }: { connection: Connection }) {
  const icon = useProcessIcon(connection.processPath);
  return (
    <span className="flex min-w-0 items-center gap-2">
      {icon ? <img alt="" className="size-4 shrink-0" src={icon} /> : <span aria-hidden="true" className="size-4 shrink-0 rounded bg-muted" />}
      <span className="truncate">{connection.process || "—"}</span>
    </span>
  );
}

export function ConnectionsPage() {
  const [paused, setPaused] = useState(false);
  const [query, setQuery] = useState("");
  const conns = useResource(() => Inspect.connections(), [], { intervalMs: paused ? undefined : 1000 });

  const rows = useMemo(() => {
    const q = query.trim().toLowerCase();
    const list = conns.data?.connections ?? [];
    const filtered = q
      ? list.filter((c) => [c.host, c.process, c.rule, c.rulePayload, ...c.chain].some((v) => v?.toLowerCase().includes(q)))
      : list;
    return [...filtered].sort((a, b) => b.downloadSpeed + b.uploadSpeed - (a.downloadSpeed + a.uploadSpeed) || b.start.localeCompare(a.start));
  }, [conns.data, query]);

  async function close(id: string) {
    try {
      await Inspect.closeConnection(id);
      await conns.reload();
    } catch (err) {
      toastManager.add({ title: t.connectionsCloseFailed(), description: errorMessage(err), type: "error" });
    }
  }

  return (
    <>
      <PageHeader
        actions={
          <>
            <SearchField label={t.connectionsSearch()} onChange={setQuery} value={query} />
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button aria-label={paused ? t.resume() : t.pause()} onClick={() => setPaused((p) => !p)} size="icon-sm" variant="outline">
                    {paused ? <PlayIcon aria-hidden="true" /> : <PauseIcon aria-hidden="true" />}
                  </Button>
                }
              />
              <TooltipPopup>{paused ? t.resume() : t.pause()}</TooltipPopup>
            </Tooltip>
            <Button disabled={rows.length === 0} onClick={() => close("")} size="sm" variant="outline">
              {t.connectionsCloseAll()}
            </Button>
          </>
        }
        id="connections"
      />
      <CoreRequired title={t.connectionsOff()}>
        <div className="flex items-center gap-4 border-b px-5 py-2 text-muted-foreground text-xs tabular-nums">
          <span>{t.connectionsOpen(String(conns.data?.connections.length ?? 0))}</span>
          <span>↓ {formatBytes(conns.data?.downloadTotal ?? 0)}</span>
          <span>↑ {formatBytes(conns.data?.uploadTotal ?? 0)}</span>
        </div>
        <div className="min-h-0 flex-1 overflow-auto">
          {rows.length === 0 ? (
            <Empty>
              <EmptyHeader>
                <EmptyTitle>{query ? t.proxiesNoMatch() : t.connectionsNone()}</EmptyTitle>
                <EmptyDescription>{query ? "" : t.connectionsNoneHint()}</EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <Table className="table-fixed text-xs">
              <TableHeader className="sticky top-0 z-10 bg-background">
                <TableRow>
                  <TableHead className="w-[28%] ps-5">{t.connectionsHost()}</TableHead>
                  <TableHead className="w-[16%]">{t.connectionsProcess()}</TableHead>
                  <TableHead className="w-[18%]">{t.connectionsRule()}</TableHead>
                  <TableHead className="w-[18%]">{t.connectionsChain()}</TableHead>
                  <TableHead className="w-[10%] text-right">{t.connectionsSpeed()}</TableHead>
                  <TableHead className="w-[7%] text-right">{t.connectionsAge()}</TableHead>
                  <TableHead className="w-10 pe-3" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {rows.map((c) => (
                  <TableRow className="group/row" key={c.id}>
                    <TableCell className="ps-5">
                      <span className="flex min-w-0 flex-col">
                        <span className="truncate font-medium" data-selectable title={c.host}>
                          {c.host}
                        </span>
                        <span className="truncate text-muted-foreground">
                          {c.network.toUpperCase()} · {c.destination}
                        </span>
                      </span>
                    </TableCell>
                    <TableCell>
                      <ProcessCell connection={c} />
                    </TableCell>
                    <TableCell>
                      <span className="block truncate" title={`${c.rule} ${c.rulePayload ?? ""}`}>
                        {c.rule}
                        {c.rulePayload && <span className="text-muted-foreground"> {c.rulePayload}</span>}
                      </span>
                    </TableCell>
                    <TableCell>
                      <span className="block truncate" title={[...c.chain].reverse().join(" → ")}>
                        {[...c.chain].reverse().join(" → ")}
                      </span>
                    </TableCell>
                    <TableCell className="text-right tabular-nums">
                      <span className={cn("block", c.downloadSpeed > 0 ? "text-foreground" : "text-muted-foreground")}>↓ {formatSpeed(c.downloadSpeed)}</span>
                      <span className="block text-muted-foreground">↑ {formatSpeed(c.uploadSpeed)}</span>
                    </TableCell>
                    <TableCell className="text-right text-muted-foreground tabular-nums">{formatUptime(c.start)}</TableCell>
                    <TableCell className="pe-3">
                      <Button
                        aria-label={t.connectionsClose(c.host)}
                        className="opacity-0 transition-opacity group-hover/row:opacity-100 focus-visible:opacity-100"
                        onClick={() => close(c.id)}
                        size="icon-xs"
                        variant="ghost"
                      >
                        <XIcon aria-hidden="true" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </div>
      </CoreRequired>
    </>
  );
}
