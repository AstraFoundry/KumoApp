import { CheckIcon, ShieldCheckIcon } from "lucide-react";
import { t } from "@/shared/i18n";
import { cn } from "@/shared/lib/utils";
import { typed, useLoop } from "../lib/use-loop";
import { Frame, Pill, Toggle } from "./frame";

function Dot({ className }: { className?: string }) {
  return <span aria-hidden="true" className={cn("size-1.5 shrink-0 rounded-full", className)} />;
}

/** Kumo connected: status, toggles and a live line of traffic. */
export function ConnectScene() {
  const time = useLoop(6000, 120);
  const points = Array.from({ length: 24 }, (_, i) => {
    const x = i * 11;
    const y = 26 - (Math.sin((i + time / 300) / 2.2) * 0.5 + 0.5) * 14 - (Math.sin((i + time / 180) / 1.3) * 0.5 + 0.5) * 6;
    return `${x},${y.toFixed(1)}`;
  });
  return (
    <>
      <Frame title="Kumo">
        <div className="flex flex-col gap-2.5 p-3.5">
          <div className="flex items-center gap-2">
            <Dot className="bg-emerald-500" />
            <span className="font-semibold text-[0.8125rem]">{t.heroConnected()}</span>
          </div>
          <span className="-mt-1.5 text-[0.625rem] text-neutral-500">{t.showcaseProfile()} · {t.modeRule()}</span>
          <svg aria-hidden="true" className="h-8 w-full" preserveAspectRatio="none" viewBox="0 0 253 30">
            <polyline fill="none" points={points.join(" ")} stroke="oklch(0.58 0.2 258)" strokeLinejoin="round" strokeWidth="1.5" />
          </svg>
          <div className="flex flex-col divide-y divide-black/5 rounded-lg border border-black/5 text-[0.6875rem]">
            <div className="flex items-center justify-between px-2.5 py-1.5">
              {t.pageSystemProxy()}
              <Toggle on />
            </div>
            <div className="flex items-center justify-between px-2.5 py-1.5">
              TUN
              <Toggle on={false} />
            </div>
          </div>
        </div>
      </Frame>
      <Pill>
        <Dot className="bg-emerald-500" />
        <span className="font-medium">Kumo</span>
        <span className="text-neutral-500 tabular-nums">↓ {(2.1 + Math.sin(time / 900) * 0.6).toFixed(1)} MB/s</span>
      </Pill>
    </>
  );
}

/** Kumo fetches Mihomo. */
export function CoreScene() {
  const time = useLoop(3600);
  const progress = Math.min(1, time / 2600);
  return (
    <>
      <Frame title={t.pageCore()}>
        <div className="flex flex-col gap-2 p-3.5">
          <span className="font-semibold text-[0.8125rem]">Mihomo</span>
          <span className="text-[0.625rem] text-neutral-500">{progress < 1 ? t.showcaseDownloading() : "v1.19.32 · arm64"}</span>
          <div className="h-1.5 overflow-hidden rounded-full bg-neutral-100">
            <div className="h-full rounded-full bg-[oklch(0.58_0.2_258)]" style={{ width: `${progress * 100}%` }} />
          </div>
          <div className="flex items-center gap-1.5 text-[0.625rem]">
            {progress >= 1 ? (
              <>
                <CheckIcon aria-hidden="true" className="size-3 text-emerald-600" />
                {t.showcaseCoreReady()}
              </>
            ) : (
              <span className="text-neutral-400 tabular-nums">{Math.round(progress * 100)}%</span>
            )}
          </div>
        </div>
      </Frame>
      <Pill>
        <span className="font-mono">mihomo-darwin-arm64.gz</span>
      </Pill>
    </>
  );
}

/** A subscription lands in the profile list. */
export function ProfileScene() {
  const time = useLoop(4200);
  const arrived = time > 1200;
  return (
    <Frame title={t.pageProfiles()}>
      <div className="flex flex-col divide-y divide-black/5 text-[0.6875rem]">
        <div
          className={cn(
            "flex flex-col gap-1.5 px-3.5 py-2.5 transition-[opacity,transform,filter] duration-500 ease-[var(--ease-out)]",
            arrived ? "opacity-100" : "-translate-y-1 opacity-0 blur-[2px]",
          )}
        >
          <div className="flex items-center gap-2">
            <span className="flex size-3 items-center justify-center rounded-full bg-[oklch(0.58_0.2_258)] text-white">
              <CheckIcon aria-hidden="true" className="size-2" strokeWidth={4} />
            </span>
            <span className="font-medium">{t.showcaseProfile()}</span>
            <span className="rounded bg-sky-500/10 px-1 text-[0.5625rem] text-sky-700">{t.profileKindRemote()}</span>
          </div>
          <div className="ms-5 h-1 overflow-hidden rounded-full bg-neutral-100">
            <div className="h-full w-2/5 rounded-full bg-[oklch(0.58_0.2_258)]" />
          </div>
          <span className="ms-5 text-[0.5625rem] text-neutral-400 tabular-nums">82 GB / 200 GB</span>
        </div>
        <div className="flex items-center gap-2 px-3.5 py-2.5">
          <span className={cn("size-3 rounded-full border", arrived ? "border-neutral-300" : "border-[oklch(0.58_0.2_258)] bg-[oklch(0.58_0.2_258)]")} />
          <span>Default</span>
          <span className="rounded bg-neutral-100 px-1 text-[0.5625rem] text-neutral-500">{t.profileKindDefault()}</span>
        </div>
      </div>
    </Frame>
  );
}

/** The system's proxy settings point at Kumo. */
export function ProxyScene() {
  const time = useLoop(4000);
  const rows = [t.showcaseWebProxy(), t.showcaseSecureProxy(), "SOCKS"];
  return (
    <>
      <Frame title={t.showcaseSystemSettings()}>
        <div className="flex flex-col divide-y divide-black/5 text-[0.6875rem]">
          {rows.map((label, i) => (
            <div className="flex items-center justify-between gap-2 px-3.5 py-2" key={label}>
              <span>{label}</span>
              <span className="flex items-center gap-2">
                <span className={cn("font-mono text-[0.625rem] text-neutral-500 transition-opacity duration-300", time > 600 + i * 350 ? "opacity-100" : "opacity-0")}>
                  127.0.0.1:7890
                </span>
                <Toggle on={time > 600 + i * 350} />
              </span>
            </div>
          ))}
        </div>
      </Frame>
      <Pill>
        <ShieldCheckIcon aria-hidden="true" className="size-3.5 text-emerald-600" />
        {t.showcaseOwned()}
      </Pill>
    </>
  );
}

const terminalLines = [
  { text: "$ kumo status", at: 0, prompt: true },
  { text: "[ok] Mihomo is running (pid 4821)", at: 1400 },
  { text: "mode           rule", at: 1600 },
  { text: "system proxy   on (manual)", at: 1800 },
  { text: "$ kumo mode global --json", at: 2600, prompt: true },
  { text: '{ "ok": true, "data": { "mode": "global" } }', at: 4000 },
];

/** The kumo command at work. */
export function TerminalScene() {
  const time = useLoop(6200, 60);
  return (
    <Frame className="w-[19rem]" dark title="zsh">
      <div className="flex min-h-36 flex-col gap-1 p-3 font-mono text-[0.625rem] leading-relaxed">
        {terminalLines.map((line) => {
          if (time < line.at) return null;
          const shown = line.prompt ? typed(line.text, (time - line.at) / 900) : line.text;
          return (
            <span className={cn(line.prompt ? "text-neutral-100" : line.text.startsWith("[ok]") ? "text-emerald-400" : "text-neutral-400")} key={line.text}>
              {shown}
              {line.prompt && shown.length < line.text.length && <span className="ms-px inline-block h-2.5 w-1 translate-y-0.5 bg-neutral-300" />}
            </span>
          );
        })}
      </div>
    </Frame>
  );
}

/** A coding agent drives Kumo through its CLI. */
export function AgentScene() {
  const time = useLoop(7000, 80);
  const step = time < 900 ? 0 : time < 2800 ? 1 : time < 4400 ? 2 : 3;
  return (
    <Frame className="w-[19rem]" title="Claude Code">
      <div className="flex min-h-40 flex-col gap-2 p-3 text-[0.6875rem]">
        <div className="self-end rounded-xl rounded-br-sm bg-neutral-100 px-2.5 py-1.5">{t.showcaseAgentAsk()}</div>
        {step >= 1 && (
          <div className="flex items-center gap-1.5 rounded-md bg-neutral-50 px-2 py-1 font-mono text-[0.5625rem] text-neutral-500">
            <span className="text-neutral-400">⏺</span>
            {typed("kumo delay 节点选择 --json", (time - 900) / 1200)}
          </div>
        )}
        {step >= 2 && (
          <div className="flex items-center gap-1.5 rounded-md bg-neutral-50 px-2 py-1 font-mono text-[0.5625rem] text-neutral-500">
            <span className="text-neutral-400">⏺</span>
            {typed("kumo select 节点选择 '🇯🇵 东京 02'", (time - 2800) / 1100)}
          </div>
        )}
        {step >= 3 && (
          <div className="flex items-center gap-1.5">
            <CheckIcon aria-hidden="true" className="size-3 text-emerald-600" />
            {t.showcaseAgentDone()}
          </div>
        )}
      </div>
    </Frame>
  );
}

/** Setup is complete. */
export function DoneScene() {
  return (
    <>
      <Frame title="Kumo">
        <div className="flex flex-col items-center gap-2 px-4 py-6">
          <svg aria-hidden="true" className="size-12" viewBox="0 0 48 48">
            <circle cx="24" cy="24" fill="rgb(16 185 129)" r="22" />
            <path className="kumo-draw-check" d="M15 25 L21.5 31.5 L34 18" fill="none" stroke="white" strokeLinecap="round" strokeLinejoin="round" strokeWidth="3.5" />
          </svg>
          <span className="font-semibold text-[0.8125rem]">{t.heroConnected()}</span>
        </div>
      </Frame>
      <Pill>
        <Dot className="bg-emerald-500" />
        <span className="font-medium">Kumo</span>
        <span className="text-neutral-500">{t.modeRule()}</span>
      </Pill>
    </>
  );
}
