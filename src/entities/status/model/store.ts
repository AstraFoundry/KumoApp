import { isMyGo } from "mygo-runtime";
import { useSyncExternalStore } from "react";
import { events, Runtime, type Status, type Traffic } from "@/shared/api/mygo";

// The status and traffic are app-wide: one subscription to Go feeds every
// component that reads them.

type Listener = () => void;

function store<T>(initial: T) {
  let value = initial;
  const listeners = new Set<Listener>();
  return {
    get: () => value,
    set(next: T) {
      value = next;
      for (const l of listeners) l();
    },
    subscribe(l: Listener) {
      listeners.add(l);
      return () => listeners.delete(l);
    },
  };
}

const status = store<Status | undefined>(undefined);
/** The last 60 seconds of throughput, oldest first. */
const traffic = store<Traffic[]>([]);
const quitting = store(false);

const HISTORY = 60;

if (isMyGo()) {
  void Runtime.status().then((s) => status.get() ?? status.set(s));
  events.status.on((s) => status.set(s));
  events.traffic.on((t) => {
    const next = traffic.get().slice(-(HISTORY - 1));
    next.push(t);
    traffic.set(next);
  });
  events.quitting.on(() => quitting.set(true));
}

/** Replaces the status after a call that returned one. */
export function setStatus(s: Status) {
  status.set(s);
}

/** The current status; undefined until Go answers. */
export function useStatus(): Status | undefined {
  return useSyncExternalStore(status.subscribe, status.get);
}

/** The recent throughput history. */
export function useTraffic(): Traffic[] {
  return useSyncExternalStore(traffic.subscribe, traffic.get);
}

/** True once Kumo started quitting. */
export function useQuitting(): boolean {
  return useSyncExternalStore(quitting.subscribe, quitting.get);
}
