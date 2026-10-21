import { ArrowRightIcon, ChevronDownIcon } from "lucide-react";
import { useMemo, useState } from "react";
import { NodeTile } from "@/entities/proxy";
import { useStatus } from "@/entities/status";
import { useSelectNode } from "@/features/select-proxy";
import { TestDelayButton, useDelays } from "@/features/test-delay";
import { type Group, Proxies } from "@/shared/api/mygo";
import { routes } from "@/shared/config/routes";
import { t } from "@/shared/i18n";
import { Link } from "@/shared/lib/router";
import { useResource } from "@/shared/lib/use-resource";
import { Button } from "@/shared/ui/button";
import { Card, CardAction, CardDescription, CardHeader, CardPanel, CardTitle } from "@/shared/ui/card";
import { Menu, MenuItem, MenuPopup, MenuTrigger } from "@/shared/ui/menu";

const COLLAPSED = 12;

/**
 * Daily node switching without leaving Overview: the first group the user
 * picks by hand, its nodes as tiles, and a delay test.
 */
export function GroupPicker() {
  const status = useStatus();
  const running = status?.state === "running";
  const groups = useResource(() => Proxies.groups(), [status?.state, status?.profile.id, status?.mode], { intervalMs: running ? 5000 : undefined });
  const { select, selected } = useSelectNode(() => void groups.reload());
  const { delays, test, isTesting } = useDelays();
  const [chosen, setChosen] = useState<string>();
  const [expanded, setExpanded] = useState(false);

  const selectable = useMemo(
    () => (groups.data ?? []).filter((g) => (g.type === "Selector" || g.type === "select") && g.name !== "GLOBAL" && !g.hidden),
    [groups.data],
  );
  const group: Group | undefined = selectable.find((g) => g.name === chosen) ?? selectable[0];

  if (!group) return null;
  const nodes = expanded ? group.nodes : group.nodes.slice(0, COLLAPSED);
  const now = selected(group.name, group.now);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">
          {selectable.length > 1 ? (
            <Menu>
              <MenuTrigger render={<Button className="-ms-2 h-7 gap-1 px-2 font-semibold text-base" variant="ghost" />}>
                {group.name}
                <ChevronDownIcon aria-hidden="true" className="opacity-60" />
              </MenuTrigger>
              <MenuPopup align="start">
                {selectable.map((g) => (
                  <MenuItem key={g.name} onClick={() => setChosen(g.name)}>
                    {g.name}
                  </MenuItem>
                ))}
              </MenuPopup>
            </Menu>
          ) : (
            group.name
          )}
        </CardTitle>
        <CardDescription>{group.live ? t.groupUsing(now ?? "—") : t.groupPreview()}</CardDescription>
        <CardAction className="flex gap-1">
          <TestDelayButton disabled={!running} onTest={() => test(group.name)} testing={isTesting(group.name)} />
          <Button aria-label={t.pageProxies()} render={<Link href={routes.proxies} />} size="icon-sm" variant="ghost">
            <ArrowRightIcon aria-hidden="true" />
          </Button>
        </CardAction>
      </CardHeader>
      <CardPanel className="flex flex-col gap-2 pt-0">
        <div className="grid grid-cols-[repeat(auto-fill,minmax(11rem,1fr))] gap-1.5">
          {nodes.map((node) => (
            <NodeTile
              delay={delays[node.name]}
              disabled={!group.live}
              key={node.name}
              node={node}
              onSelect={() => select(group.name, node.name)}
              selected={group.live && node.name === now}
            />
          ))}
        </div>
        {group.nodes.length > COLLAPSED && (
          <Button className="self-start" onClick={() => setExpanded((e) => !e)} size="xs" variant="ghost">
            {expanded ? t.groupShowLess() : t.groupShowAll(String(group.nodes.length))}
          </Button>
        )}
      </CardPanel>
    </Card>
  );
}
