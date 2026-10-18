import { onLocaleChange, isMyGo } from "mygo-runtime";
import { useEffect, useState } from "react";
import { messages } from "@/shared/api/locales";

/** The app's messages in the language of the moment. */
export const t = messages;

/**
 * Re-renders when the app's language changes, and keeps <html lang> in
 * step. Use it once, at the root: every message reads the language anew.
 */
export function useLocaleVersion(): number {
  const [version, setVersion] = useState(0);
  useEffect(() => {
    if (!isMyGo()) return;
    return onLocaleChange((locale, dir) => {
      document.documentElement.lang = locale;
      document.documentElement.dir = dir;
      setVersion((v) => v + 1);
    });
  }, []);
  return version;
}
