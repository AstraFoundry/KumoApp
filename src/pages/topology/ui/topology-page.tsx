import { PageHeader } from "@/widgets/page-header";
import { TopologyCanvas } from "@/widgets/topology-canvas";

export function TopologyPage() {
  return (
    <>
      <PageHeader id="topology" />
      <TopologyCanvas />
    </>
  );
}
