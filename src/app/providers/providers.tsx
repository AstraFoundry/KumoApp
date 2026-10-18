import type { ReactNode } from "react";
import { ToastProvider } from "@/shared/ui/toast";
import { TooltipProvider } from "@/shared/ui/tooltip";
import { useTheme } from "./theme";

export function Providers({ children }: { children: ReactNode }) {
  useTheme();
  return (
    <TooltipProvider delay={500}>
      <ToastProvider position="bottom-right">{children}</ToastProvider>
    </TooltipProvider>
  );
}
