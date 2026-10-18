import { type AnchorHTMLAttributes, type MouseEvent, useSyncExternalStore } from "react";

// A small history router: the app has one window and flat routes, so the
// path is the whole state. MyGo serves index.html for every path.

type Listener = () => void;
const listeners = new Set<Listener>();

function notify() {
  for (const listener of listeners) listener();
}

if (typeof window !== "undefined") {
  window.addEventListener("popstate", notify);
}

function subscribe(listener: Listener) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

function snapshot() {
  return window.location.pathname + window.location.search;
}

/** Navigates to path, replacing the entry when replace is set. */
export function navigate(path: string, replace = false) {
  if (path === snapshot()) return;
  if (replace) window.history.replaceState(null, "", path);
  else window.history.pushState(null, "", path);
  notify();
}

/** The current location, re-rendering on change. */
export function useLocation(): { path: string; query: URLSearchParams } {
  const location = useSyncExternalStore(subscribe, snapshot);
  const [path = "/", search = ""] = location.split("?");
  return { path, query: new URLSearchParams(search) };
}

export function Link({ href, onClick, ...props }: AnchorHTMLAttributes<HTMLAnchorElement> & { href: string }) {
  return (
    <a
      href={href}
      onClick={(event: MouseEvent<HTMLAnchorElement>) => {
        onClick?.(event);
        if (event.defaultPrevented || event.metaKey || event.ctrlKey || event.button !== 0) return;
        event.preventDefault();
        navigate(href);
      }}
      {...props}
    />
  );
}
