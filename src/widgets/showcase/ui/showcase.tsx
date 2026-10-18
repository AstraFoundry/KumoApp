import { AnimatePresence, motion } from "motion/react";
import { backdrop } from "@/shared/assets";
import { t } from "@/shared/i18n";
import { cn } from "@/shared/lib/utils";
import { AgentScene, ConnectScene, CoreScene, DoneScene, ProfileScene, ProxyScene, TerminalScene } from "./scenes";

export type Scene = "connect" | "core" | "profile" | "proxy" | "terminal" | "agent" | "done";

const scenes: Record<Scene, () => React.ReactElement> = {
  connect: ConnectScene,
  core: CoreScene,
  profile: ProfileScene,
  proxy: ProxyScene,
  terminal: TerminalScene,
  agent: AgentScene,
  done: DoneScene,
};

/**
 * Kumo at work, drawn over a photograph of clouds: a small mock of what
 * the current step enables. Scenes cross-fade with a slight blur so the
 * change reads as one surface settling, not two images swapping.
 */
export function Showcase({ scene, className }: { scene: Scene; className?: string }) {
  const Current = scenes[scene];
  return (
    <div aria-label={t.showcaseLabel()} className={cn("relative isolate overflow-hidden bg-sky-200", className)} role="img">
      <img alt="" className="absolute inset-0 size-full object-cover" draggable={false} src={backdrop} />
      <div aria-hidden="true" className="absolute inset-x-0 top-0 h-28 bg-gradient-to-b from-black/25 to-transparent" />
      <div aria-hidden="true" className="absolute inset-x-0 bottom-0 h-56 bg-gradient-to-t from-black/30 to-transparent" />
      <AnimatePresence initial={false} mode="popLayout">
        <motion.div
          animate={{ opacity: 1, scale: 1, filter: "blur(0px)", y: 0 }}
          className="absolute inset-0 flex flex-col items-center justify-center"
          exit={{ opacity: 0, scale: 0.98, filter: "blur(4px)" }}
          initial={{ opacity: 0, scale: 0.98, filter: "blur(4px)", y: 8 }}
          key={scene}
          transition={{ duration: 0.36, ease: [0.23, 1, 0.32, 1] }}
        >
          <Current />
        </motion.div>
      </AnimatePresence>
    </div>
  );
}
