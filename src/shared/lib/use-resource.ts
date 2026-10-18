import { type DependencyList, useCallback, useEffect, useRef, useState } from "react";

export interface Resource<T> {
  data: T | undefined;
  error: unknown;
  loading: boolean;
  reload: () => Promise<void>;
  /** Replaces the data locally, e.g. after a mutation returned it. */
  set: (data: T) => void;
}

/**
 * Loads data from a Go service, again when deps change and, with
 * intervalMs, on a timer while the page is visible. Reloads keep the old
 * data on screen until the new data arrives, so nothing flickers.
 */
export function useResource<T>(
  load: () => Promise<T>,
  deps: DependencyList,
  options: { intervalMs?: number; enabled?: boolean } = {},
): Resource<T> {
  const { intervalMs, enabled = true } = options;
  const [data, setData] = useState<T>();
  const [error, setError] = useState<unknown>();
  const [loading, setLoading] = useState(enabled);
  const loadRef = useRef(load);
  loadRef.current = load;
  const generation = useRef(0);

  const reload = useCallback(async () => {
    const current = ++generation.current;
    setLoading(true);
    try {
      const value = await loadRef.current();
      if (current === generation.current) {
        setData(value);
        setError(undefined);
      }
    } catch (err) {
      if (current === generation.current) setError(err);
    } finally {
      if (current === generation.current) setLoading(false);
    }
  }, []);

  // biome-ignore lint/correctness/useExhaustiveDependencies: deps are the caller's
  useEffect(() => {
    if (!enabled) return;
    void reload();
    if (!intervalMs) return;
    const timer = window.setInterval(() => {
      if (document.visibilityState === "visible") void reload();
    }, intervalMs);
    return () => window.clearInterval(timer);
  }, [...deps, enabled, intervalMs]);

  return { data, error, loading, reload, set: setData };
}
