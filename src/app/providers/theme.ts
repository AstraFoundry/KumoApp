import { isMyGo } from "mygo-runtime";
import { useEffect } from "react";
import { Preferences } from "@/shared/api/mygo";

const media = window.matchMedia("(prefers-color-scheme: dark)");

/** Applies the preferred theme: the system's, light or dark. */
export function applyTheme(theme: string) {
  const dark = theme === "dark" || (theme !== "light" && media.matches);
  document.documentElement.classList.toggle("dark", dark);
  document.documentElement.style.colorScheme = dark ? "dark" : "light";
}

let current = "system";

export function setTheme(theme: string) {
  current = theme;
  applyTheme(theme);
}

export function useTheme() {
  useEffect(() => {
    applyTheme(current);
    if (isMyGo()) void Preferences.get().then((p) => setTheme(p.theme));
    const listener = () => applyTheme(current);
    media.addEventListener("change", listener);
    return () => media.removeEventListener("change", listener);
  }, []);
}
