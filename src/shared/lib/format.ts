import { currentLocale } from "mygo-runtime/i18n";

const units = ["B", "KB", "MB", "GB", "TB"];

/** Formats a byte count, such as 1.2 MB. */
export function formatBytes(bytes: number): string {
  let value = Math.max(0, bytes);
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }
  const digits = value >= 100 || unit === 0 ? 0 : 1;
  return `${value.toFixed(digits)} ${units[unit]}`;
}

/** Formats a throughput, such as 1.2 MB/s. */
export function formatSpeed(bytesPerSecond: number): string {
  return `${formatBytes(bytesPerSecond)}/s`;
}

/** Formats a time relative to now: "3 minutes ago". */
export function formatRelative(iso: string | undefined | null): string {
  if (!iso) return "—";
  const date = new Date(iso);
  if (Number.isNaN(date.getTime()) || date.getFullYear() < 2000) return "—";
  const seconds = Math.round((date.getTime() - Date.now()) / 1000);
  const rtf = new Intl.RelativeTimeFormat(currentLocale(), { numeric: "auto" });
  const steps: [Intl.RelativeTimeFormatUnit, number][] = [
    ["year", 31536000],
    ["month", 2592000],
    ["day", 86400],
    ["hour", 3600],
    ["minute", 60],
  ];
  for (const [unit, size] of steps) {
    if (Math.abs(seconds) >= size) return rtf.format(Math.round(seconds / size), unit);
  }
  return rtf.format(seconds, "second");
}

/** Formats a date in the user's short style. */
export function formatDate(iso: string | number | undefined | null): string {
  if (!iso) return "—";
  const date = typeof iso === "number" ? new Date(iso * 1000) : new Date(iso);
  if (Number.isNaN(date.getTime())) return "—";
  return new Intl.DateTimeFormat(currentLocale(), { dateStyle: "medium" }).format(date);
}

/** Formats a duration since start, such as 2h 5m. */
export function formatUptime(iso: string | undefined | null): string {
  if (!iso) return "—";
  const seconds = Math.max(0, Math.floor((Date.now() - new Date(iso).getTime()) / 1000));
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  if (h > 0) return `${h}h ${m}m`;
  if (m > 0) return `${m}m`;
  return `${seconds}s`;
}

export type DelayGrade = "unknown" | "fast" | "slow" | "timeout";

/** Buckets a delay as the Go side does: under 300 ms is fast. */
export function delayGrade(delay: number | undefined): DelayGrade {
  if (delay === undefined || delay === 0) return "unknown";
  if (delay < 0) return "timeout";
  return delay < 300 ? "fast" : "slow";
}

/** The flag emoji of an ISO region code. */
export function flag(country: string | undefined): string {
  if (!country || country.length !== 2) return "";
  return String.fromCodePoint(...[...country.toUpperCase()].map((c) => 0x1f1e6 + c.charCodeAt(0) - 65));
}

/** Drops flag emoji from a node name, for rows that draw the flag apart. */
export function displayName(name: string): string {
  const stripped = name.replace(/[\u{1F1E6}-\u{1F1FF}]{2}/gu, "").replace(/\s+/g, " ").trim().replace(/^[-_·•/|｜\s]+|[-_·•/|｜\s]+$/g, "");
  return stripped || name;
}
