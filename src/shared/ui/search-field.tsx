import { SearchIcon } from "lucide-react";
import { cn } from "@/shared/lib/utils";
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/shared/ui/input-group";

/** The search box pages put in their header. */
export function SearchField({ value, onChange, label, className }: { value: string; onChange: (v: string) => void; label: string; className?: string }) {
  return (
    <InputGroup className={cn("w-56", className)}>
      <InputGroupAddon>
        <SearchIcon aria-hidden="true" />
      </InputGroupAddon>
      <InputGroupInput aria-label={label} onChange={(e) => onChange(e.target.value)} placeholder={label} size="sm" type="search" value={value} />
    </InputGroup>
  );
}
