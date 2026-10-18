import { PageHeader } from "@/widgets/page-header";
import { TopologyCanvas } from "@/widgets/topology-canvas";

export function TopologyPage() {
  return <TopologyCanvas header={(actions) => <PageHeader actions={actions} id="topology" />} />;
}
