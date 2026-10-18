import { ArrowDownIcon, ArrowUpIcon } from "lucide-react";
import { useId } from "react";
import { useTraffic } from "@/entities/status";
import type { Traffic } from "@/shared/api/mygo";
import { t } from "@/shared/i18n";
import { formatBytes, formatSpeed } from "@/shared/lib/format";

const SAMPLES = 60;

/** A smooth path through points, with control points that never overshoot. */
function smooth(points: [number, number][]): string {
  if (points.length === 0) return "";
  let d = `M${points[0]?.[0]},${points[0]?.[1]}`;
  for (let i = 1; i < points.length; i++) {
    const [x0, y0] = points[i - 1] as [number, number];
    const [x1, y1] = points[i] as [number, number];
    const mx = (x0 + x1) / 2;
    d += ` C${mx},${y0} ${mx},${y1} ${x1},${y1}`;
  }
  return d;
}

function series(samples: Traffic[], pick: (s: Traffic) => number, max: number, w: number, h: number): [number, number][] {
  const padded = Array.from({ length: SAMPLES - samples.length }, () => 0).concat(samples.map(pick));
  return padded.map((v, i) => [(i / (SAMPLES - 1)) * w, h - (v / max) * (h - 4) - 2]);
}

/**
 * The last minute of throughput: download as the filled area, upload as a
 * line. It redraws once a second without animating, as a live gauge.
 */
export function TrafficChart() {
  const samples = useTraffic();
  const gradient = useId();
  const last = samples.at(-1);
  const max = Math.max(1024, ...samples.map((s) => Math.max(s.up, s.down))) * 1.15;
  const w = 600;
  const h = 120;
  const down = series(samples, (s) => s.down, max, w, h);
  const up = series(samples, (s) => s.up, max, w, h);
  const downPath = smooth(down);

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-end gap-6">
        <Reading icon={<ArrowDownIcon aria-hidden="true" className="size-3.5 text-brand" />} label={t.trafficDownload()} total={last?.downTotal ?? 0} value={last?.down ?? 0} />
        <Reading icon={<ArrowUpIcon aria-hidden="true" className="size-3.5 text-muted-foreground" />} label={t.trafficUpload()} total={last?.upTotal ?? 0} value={last?.up ?? 0} />
      </div>
      <svg aria-hidden="true" className="h-28 w-full overflow-visible" preserveAspectRatio="none" viewBox={`0 0 ${w} ${h}`}>
        <defs>
          <linearGradient id={gradient} x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor="var(--brand)" stopOpacity="0.28" />
            <stop offset="100%" stopColor="var(--brand)" stopOpacity="0" />
          </linearGradient>
        </defs>
        {[0.25, 0.5, 0.75].map((f) => (
          <line className="stroke-border" key={f} strokeDasharray="2 4" vectorEffect="non-scaling-stroke" x1="0" x2={w} y1={h * f} y2={h * f} />
        ))}
        <path d={`${downPath} L${w},${h} L0,${h} Z`} fill={`url(#${gradient})`} />
        <path d={downPath} fill="none" stroke="var(--brand)" strokeWidth="1.75" vectorEffect="non-scaling-stroke" />
        <path className="stroke-muted-foreground/60" d={smooth(up)} fill="none" strokeDasharray="3 3" strokeWidth="1.25" vectorEffect="non-scaling-stroke" />
      </svg>
    </div>
  );
}

function Reading({ icon, label, value, total }: { icon: React.ReactNode; label: string; value: number; total: number }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="flex items-center gap-1 text-muted-foreground text-xs">
        {icon}
        {label}
      </span>
      <span className="font-semibold text-xl tabular-nums tracking-tight">{formatSpeed(value)}</span>
      <span className="text-muted-foreground text-xs tabular-nums">{t.trafficTotal(formatBytes(total))}</span>
    </div>
  );
}
