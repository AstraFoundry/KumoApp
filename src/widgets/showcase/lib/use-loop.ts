import { useReducedMotion } from "motion/react";
import { useEffect, useState } from "react";

/**
 * The time within a repeating cycle, ticking every step ms. With reduced
 * motion it holds the cycle's last moment, so scenes show their outcome.
 */
export function useLoop(cycle: number, step = 100): number {
  const reduced = useReducedMotion();
  const [t, setT] = useState(reduced ? cycle : 0);
  useEffect(() => {
    if (reduced) {
      setT(cycle);
      return;
    }
    const start = performance.now();
    const timer = window.setInterval(() => setT((performance.now() - start) % cycle), step);
    return () => window.clearInterval(timer);
  }, [cycle, step, reduced]);
  return t;
}

/** The first fraction of text, for typewriter lines. */
export function typed(text: string, fraction: number): string {
  const chars = [...text];
  return chars.slice(0, Math.round(chars.length * Math.max(0, Math.min(1, fraction)))).join("");
}
