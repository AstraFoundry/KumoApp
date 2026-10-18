import { AnimatePresence, motion } from "motion/react";
import { CheckCircle2Icon, FileUpIcon } from "lucide-react";
import { type ReactNode, useState } from "react";
import { setStatus } from "@/entities/status";
import { LanguageSelect } from "@/features/choose-language";
import { CliLinkCard } from "@/features/install-cli";
import { SkillTargets } from "@/features/install-skills";
import { App, Preferences, Profiles, Runtime, Settings, SystemProxy } from "@/shared/api/mygo";
import { mark } from "@/shared/assets";
import { t } from "@/shared/i18n";
import { errorMessage } from "@/shared/lib/errors";
import { navigate } from "@/shared/lib/router";
import { useResource } from "@/shared/lib/use-resource";
import { cn } from "@/shared/lib/utils";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Switch } from "@/shared/ui/switch";
import { toastManager } from "@/shared/ui/toast";
import { type Scene, Showcase } from "@/widgets/showcase";

type Step = "welcome" | "core" | "profile" | "proxy" | "cli" | "agents" | "done";
const steps: Step[] = ["welcome", "core", "profile", "proxy", "cli", "agents", "done"];
const sceneOf: Record<Step, Scene> = {
  welcome: "connect",
  core: "core",
  profile: "profile",
  proxy: "proxy",
  cli: "terminal",
  agents: "agent",
  done: "done",
};

/**
 * First run: each step beside an illustration of what it enables. Every
 * step can be skipped; nothing here is needed to start Kumo later.
 */
export function SetupPage() {
  const [index, setIndex] = useState(0);
  const [direction, setDirection] = useState(1);
  const step = steps[index] as Step;
  const setup = useResource(() => App.setup(), [index]);

  function move(by: number) {
    setDirection(by);
    setIndex((i) => Math.max(0, Math.min(steps.length - 1, i + by)));
  }

  async function finish(start: boolean) {
    await App.completeSetup();
    if (start) {
      try {
        const outcome = await Runtime.start(true);
        setStatus(outcome.status);
      } catch (err) {
        toastManager.add({ title: t.coreStartFailed(), description: errorMessage(err), type: "error" });
      }
    }
    navigate("/", true);
  }

  return (
    <div className="flex h-svh bg-background">
      <Showcase className="w-[min(40%,25rem)] shrink-0" scene={sceneOf[step]} />
      {/* One column carries the header, the step and the footer, so they
          share both edges. The pane above it drags the window. */}
      <div className="flex min-w-0 flex-1 flex-col px-10 [--app-region:drag]">
        <div className="mx-auto flex h-16 w-full max-w-md shrink-0 items-center justify-between">
          <Stepper current={index} total={steps.length} />
          <div className="-me-1.75 [--app-region:no-drag]">
            <LanguageSelect quiet />
          </div>
        </div>
        <div className="relative mx-auto min-h-0 w-full max-w-md flex-1 overflow-hidden [--app-region:no-drag]">
          <AnimatePresence custom={direction} initial={false} mode="popLayout">
            <motion.div
              animate={{ opacity: 1, x: 0 }}
              className="absolute inset-0 flex flex-col justify-center overflow-y-auto py-6"
              custom={direction}
              exit={{ opacity: 0, x: -16 * direction }}
              initial={{ opacity: 0, x: 24 * direction }}
              key={step}
              transition={{ duration: 0.28, ease: [0.23, 1, 0.32, 1] }}
            >
              <div className="flex flex-col gap-6">
                {step === "welcome" && <Welcome />}
                {step === "core" && <CoreStep found={setup.data?.coreFound} path={setup.data?.corePath} onChange={() => setup.reload()} />}
                {step === "profile" && <ProfileStep count={setup.data?.profiles ?? 1} />}
                {step === "proxy" && <ProxyStep />}
                {step === "cli" && (
                  <>
                    <Heading body={t.setupCliBody()} title={t.setupCliTitle()} />
                    <CliLinkCard />
                    <pre className="rounded-lg bg-muted/50 px-3 py-2 font-mono text-muted-foreground text-xs" data-selectable>
                      kumo status{"\n"}kumo proxies --json{"\n"}kumo mode global
                    </pre>
                  </>
                )}
                {step === "agents" && (
                  <>
                    <Heading body={t.setupAgentsBody()} title={t.setupAgentsTitle()} />
                    <SkillTargets compact />
                  </>
                )}
                {step === "done" && (
                  <>
                    <svg aria-hidden="true" className="size-14" viewBox="0 0 48 48">
                      <circle cx="24" cy="24" fill="var(--color-success)" r="22" />
                      <path className="kumo-draw-check" d="M15 25 L21.5 31.5 L34 18" fill="none" stroke="white" strokeLinecap="round" strokeLinejoin="round" strokeWidth="3.5" />
                    </svg>
                    <Heading body={t.setupDoneBody()} large title={t.setupDoneTitle()} />
                  </>
                )}
              </div>
            </motion.div>
          </AnimatePresence>
        </div>
        <div className="mx-auto flex h-20 w-full max-w-md shrink-0 items-center gap-2 [--app-region:no-drag]">
          {index > 0 && step !== "done" && (
            <Button className="-ms-2.75" onClick={() => move(-1)} variant="ghost">
              {t.setupBack()}
            </Button>
          )}
          {step === "welcome" && (
            <Button className="-ms-2.75 text-muted-foreground" onClick={() => finish(false)} variant="ghost">
              {t.setupSkipAll()}
            </Button>
          )}
          <div className="ms-auto flex gap-2">
            {step === "done" ? (
              <>
                <Button onClick={() => finish(false)} variant="outline">
                  {t.setupFinish()}
                </Button>
                <Button onClick={() => finish(true)}>{t.setupFinishStart()}</Button>
              </>
            ) : (
              <Button onClick={() => move(1)}>{step === "welcome" ? t.setupStart() : t.setupContinue()}</Button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function Stepper({ current, total }: { current: number; total: number }) {
  return (
    <div aria-label={t.setupStep(String(current + 1), String(total))} className="flex items-center gap-1.5" role="img">
      {Array.from({ length: total }, (_, i) => (
        <span
          className={cn(
            "h-1.5 rounded-full transition-[width,background-color] duration-300 ease-[var(--ease-out)]",
            i === current ? "w-5 bg-foreground" : i < current ? "w-1.5 bg-foreground/50" : "w-1.5 bg-foreground/15",
          )}
          // biome-ignore lint/suspicious/noArrayIndexKey: fixed positions
          key={i}
        />
      ))}
    </div>
  );
}

function Heading({ title, body, large }: { title: string; body: ReactNode; large?: boolean }) {
  return (
    <div className="flex flex-col gap-2">
      <h1 className={cn("font-semibold tracking-tight", large ? "text-3xl" : "text-2xl")}>{title}</h1>
      <p className="text-muted-foreground text-sm leading-relaxed">{body}</p>
    </div>
  );
}

function Welcome() {
  return (
    <>
      <img alt="" className="-m-1 size-16 rounded-[1.1rem]" src={mark} />
      <Heading body={t.setupWelcomeBody()} large title={t.setupWelcomeTitle()} />
    </>
  );
}

function CoreStep({ found, path, onChange }: { found?: boolean; path?: string; onChange: () => void }) {
  const [busy, setBusy] = useState(false);
  async function install() {
    setBusy(true);
    try {
      await Settings.installCore();
      onChange();
    } catch (err) {
      toastManager.add({ title: t.coreInstallFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setBusy(false);
    }
  }
  return (
    <>
      <Heading body={t.setupCoreBody()} title={t.setupCoreTitle()} />
      {found ? (
        <div className="flex items-center gap-3 rounded-xl border bg-card px-4 py-3">
          <CheckCircle2Icon aria-hidden="true" className="size-4 shrink-0 text-success" />
          <span className="min-w-0 flex-1 truncate font-mono text-xs" data-selectable>
            {path}
          </span>
          <Button loading={busy} onClick={install} size="xs" variant="ghost">
            {t.coreUpdate()}
          </Button>
        </div>
      ) : (
        <div className="flex gap-2">
          <Button loading={busy} onClick={install}>
            {t.coreInstall()}
          </Button>
          <Button
            onClick={async () => {
              await Settings.chooseCore();
              onChange();
            }}
            variant="outline"
          >
            {t.coreChoose()}
          </Button>
        </div>
      )}
    </>
  );
}

function ProfileStep({ count }: { count: number }) {
  const [url, setURL] = useState("");
  const [busy, setBusy] = useState(false);
  const [added, setAdded] = useState<string>();
  async function add(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    try {
      const p = await Profiles.import({ url: url.trim(), activate: true });
      setAdded(p.name);
      setURL("");
    } catch (err) {
      toastManager.add({ title: t.setupProfileFailed(), description: errorMessage(err), type: "error" });
    } finally {
      setBusy(false);
    }
  }
  return (
    <>
      <Heading body={t.setupProfileBody()} title={t.setupProfileTitle()} />
      <form className="flex gap-2" onSubmit={add}>
        <Input aria-label={t.importURL()} className="flex-1 font-mono" onChange={(e) => setURL(e.target.value)} placeholder="https://" value={url} />
        <Button disabled={!/^https?:\/\/\S+$/.test(url.trim())} loading={busy} type="submit">
          {t.importSubmit()}
        </Button>
      </form>
      <Button
        className="self-start"
        onClick={async () => {
          const p = await Profiles.chooseFile(true);
          if (p) setAdded(p.name);
        }}
        size="sm"
        variant="ghost"
      >
        <FileUpIcon aria-hidden="true" />
        {t.importChooseFile()}
      </Button>
      {(added || count > 1) && (
        <p className="flex items-center gap-2 text-sm">
          <CheckCircle2Icon aria-hidden="true" className="size-4 text-success" />
          {added ? t.importDone(added) : t.setupProfileHas()}
        </p>
      )}
    </>
  );
}

function ProxyStep() {
  const prefs = useResource(() => Preferences.get(), []);
  const status = useResource(() => Runtime.status(), []);
  return (
    <>
      <Heading body={t.setupProxyBody()} title={t.setupProxyTitle()} />
      <div className="flex flex-col divide-y rounded-xl border bg-card">
        <label className="flex items-center gap-3 px-4 py-3">
          <span className="flex flex-1 flex-col gap-0.5">
            <span className="font-medium text-sm">{t.spUse()}</span>
            <span className="text-muted-foreground text-xs">{t.setupProxyHint()}</span>
          </span>
          <Switch
            checked={status.data?.systemProxy.enabled ?? false}
            onCheckedChange={async (v) => status.set(await SystemProxy.setEnabled(v))}
          />
        </label>
        <label className="flex items-center gap-3 px-4 py-3">
          <span className="flex flex-1 flex-col gap-0.5">
            <span className="font-medium text-sm">{t.settingsStartCore()}</span>
          </span>
          <Switch
            checked={prefs.data?.startCoreOnLaunch ?? false}
            onCheckedChange={async (v) => {
              if (!prefs.data) return;
              const next = { ...prefs.data, startCoreOnLaunch: v };
              prefs.set(next);
              await Preferences.save(next);
            }}
          />
        </label>
      </div>
    </>
  );
}
