import { isMyGo } from "mygo-runtime";
import { useEffect } from "react";
import { events, Profiles, type Summary } from "@/shared/api/mygo";
import { type Resource, useResource } from "@/shared/lib/use-resource";

/** Every profile, reloaded whenever any surface changes one. */
export function useProfiles(): Resource<Summary[]> {
  const resource = useResource(() => Profiles.list(), []);
  const { reload } = resource;
  useEffect(() => {
    if (!isMyGo()) return;
    return events.profilesChanged.on(() => void reload());
  }, [reload]);
  return resource;
}
