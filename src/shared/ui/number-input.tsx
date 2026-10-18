import { useEffect, useState } from "react";
import { cn } from "@/shared/lib/utils";
import { Input } from "@/shared/ui/input";

/** A numeric field that keeps what is typed until it parses. */
export function NumberInput({
  value,
  onChange,
  min,
  max,
  label,
  className,
}: {
  value: number;
  onChange: (next: number) => void;
  min?: number;
  max?: number;
  label: string;
  className?: string;
}) {
  const [text, setText] = useState(String(value));
  useEffect(() => setText(String(value)), [value]);
  const n = Number(text);
  const invalid = text.trim() === "" || !Number.isInteger(n) || (min !== undefined && n < min) || (max !== undefined && n > max);
  return (
    <Input
      aria-invalid={invalid || undefined}
      aria-label={label}
      className={cn("w-24 font-mono tabular-nums", className)}
      inputMode="numeric"
      onChange={(e) => {
        setText(e.target.value);
        const next = Number(e.target.value);
        if (e.target.value.trim() !== "" && Number.isInteger(next) && (min === undefined || next >= min) && (max === undefined || next <= max)) onChange(next);
      }}
      size="sm"
      value={text}
    />
  );
}
