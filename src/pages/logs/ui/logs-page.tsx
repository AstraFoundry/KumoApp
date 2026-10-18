import { Channel, isMyGo } from "mygo-runtime";
import { FolderOpenIcon, PauseIcon, PlayIcon, Trash2Icon } from "lucide-react";
import { useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { useStatus } from "@/entities/status";
import { App, Inspect, type LogEntry, Logs } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { cn } from "@/shared/lib/utils";
import { Button } from "@/shared/ui/button";
import { Empty, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { SearchField } from "@/shared/ui/search-field";
import { ToggleGroup, ToggleGroupItem } from "@/shared/ui/toggle-group";
import { Tooltip, TooltipPopup, TooltipTrigger } from "@/shared/ui/tooltip";
import { PageHeader } from "@/widgets/page-header";

const LIMIT = 2000;
type Level = "debug" | "info" | "warning" | "error";
const rank: Record<string, number> = { debug: 0, info: 1, kumo: 1, warning: 2, error: 3 };

const tone: Record<string, string> = {
  debug: "text-muted-foreground",
  info: "text-info-foreground",
  kumo: "text-brand-foreground",
  warning: "text-warning-foreground",
  error: "text-destructive-foreground",
};

/**
 * The core's log: the file's recent lines, then live lines while Mihomo
 * runs. It sticks to the bottom unless the user scrolls up or pauses.
 */
export function LogsPage() {
  const status = useStatus();
  const running = status?.state === "running";
  const [entries, setEntries] = useState<LogEntry[]>([]);
  const [level, setLevel] = useState<Level>("info");
  const [paused, setPaused] = useState(false);
  const [query, setQuery] = useState("");
  const scroller = useRef<HTMLDivElement>(null);
  const stick = useRef(true);
  const pausedRef = useRef(paused);
  pausedRef.current = paused;

  useEffect(() => {
    void Inspect.recentLogs(500).then(setEntries);
  }, []);

  useEffect(() => {
    if (!running || !isMyGo()) return;
    const lines = new Channel<LogEntry>((entry) => {
      if (pausedRef.current) return;
      setEntries((list) => (list.length >= LIMIT ? [...list.slice(-LIMIT + 1), entry] : [...list, entry]));
    });
    void Logs.follow(level === "info" ? "info" : level, lines).catch(() => {});
    return () => lines.close();
  }, [running, level]);

  const visible = useMemo(() => {
    const q = query.trim().toLowerCase();
    return entries.filter((e) => (rank[e.level] ?? 1) >= (rank[level] ?? 1) && (!q || e.message.toLowerCase().includes(q)));
  }, [entries, level, query]);

  useLayoutEffect(() => {
    const el = scroller.current;
    if (el && stick.current) el.scrollTop = el.scrollHeight;
  }, [visible]);

  return (
    <>
      <PageHeader
        actions={
          <>
            <SearchField label={t.logsSearch()} onChange={setQuery} value={query} />
            <ToggleGroup onValueChange={(v) => v[0] && setLevel(v[0] as Level)} size="sm" value={[level]} variant="outline">
              <ToggleGroupItem value="debug">{t.logsDebug()}</ToggleGroupItem>
              <ToggleGroupItem value="info">{t.logsInfo()}</ToggleGroupItem>
              <ToggleGroupItem value="warning">{t.logsWarning()}</ToggleGroupItem>
              <ToggleGroupItem value="error">{t.logsError()}</ToggleGroupItem>
            </ToggleGroup>
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
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button aria-label={t.logsClear()} onClick={() => setEntries([])} size="icon-sm" variant="ghost">
                    <Trash2Icon aria-hidden="true" />
                  </Button>
                }
              />
              <TooltipPopup>{t.logsClear()}</TooltipPopup>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button aria-label={t.logsReveal()} onClick={() => App.reveal("coreLog")} size="icon-sm" variant="ghost">
                    <FolderOpenIcon aria-hidden="true" />
                  </Button>
                }
              />
              <TooltipPopup>{t.logsReveal()}</TooltipPopup>
            </Tooltip>
          </>
        }
        id="logs"
      />
      <div
        className="min-h-0 flex-1 overflow-y-auto bg-muted/20 py-2 font-mono text-[0.71875rem] leading-relaxed"
        onScroll={(e) => {
          const el = e.currentTarget;
          stick.current = el.scrollHeight - el.scrollTop - el.clientHeight < 24;
        }}
        ref={scroller}
      >
        {visible.length === 0 ? (
          <Empty>
            <EmptyHeader>
              <EmptyTitle>{running ? t.logsWaiting() : t.logsEmpty()}</EmptyTitle>
            </EmptyHeader>
          </Empty>
        ) : (
          <ol data-selectable>
            {visible.map((entry, index) => (
              <li className="flex gap-3 px-5 py-px hover:bg-accent/40" key={`${index}-${entry.time}`}>
                <span className="w-16 shrink-0 text-muted-foreground tabular-nums">
                  {entry.time && !entry.time.startsWith("0001") ? new Date(entry.time).toLocaleTimeString([], { hour12: false }) : ""}
                </span>
                <span className={cn("w-14 shrink-0 uppercase", tone[entry.level])}>{entry.level}</span>
                <span className="min-w-0 flex-1 break-words">{entry.message}</span>
              </li>
            ))}
          </ol>
        )}
      </div>
    </>
  );
}
