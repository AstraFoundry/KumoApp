import { Select, SelectItem, SelectPopup, SelectTrigger, SelectValue } from "@/shared/ui/select";

export interface Option<T extends string> {
  value: T;
  label: string;
}

/** A Select over a short list of string options. */
export function OptionSelect<T extends string>({
  value,
  onChange,
  options,
  label,
  className,
}: {
  value: T;
  onChange: (next: T) => void;
  options: Option<T>[];
  label: string;
  className?: string;
}) {
  const selected = options.find((o) => o.value === value) ?? options[0];
  return (
    <Select
      items={options}
      onValueChange={(item) => item && onChange((item as Option<T>).value)}
      value={selected}
      itemToStringValue={(item) => (item as Option<T>).value}
    >
      <SelectTrigger aria-label={label} className={className ?? "w-44"} size="sm">
        <SelectValue />
      </SelectTrigger>
      <SelectPopup>
        {options.map((o) => (
          <SelectItem key={o.value} value={o}>
            {o.label}
          </SelectItem>
        ))}
      </SelectPopup>
    </Select>
  );
}
