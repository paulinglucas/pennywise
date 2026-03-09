import { useState } from "react";
import { useProjections, DEFAULT_PARAMS, type ProjectionParams } from "@/hooks/useProjections";
import { useDashboard } from "@/hooks/useDashboard";
import ScenarioSliders from "@/components/projections/ScenarioSliders";
import ProjectionChart from "@/components/projections/ProjectionChart";
import ProjectionSummary from "@/components/projections/ProjectionSummary";
import { SkeletonCard, SkeletonChart } from "@/components/shared/Skeleton";

function ProjectionSkeleton() {
  return (
    <div className="flex flex-col gap-6" data-testid="projection-skeleton">
      <SkeletonChart />
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <SkeletonCard />
        <SkeletonCard />
        <SkeletonCard />
      </div>
    </div>
  );
}

export default function Projections() {
  const [params, setParams] = useState<ProjectionParams>(DEFAULT_PARAMS);
  const projection = useProjections(params);
  const dashboard = useDashboard();
  const baseMonthlySavings = Math.max(dashboard.data?.cash_flow_this_month ?? 0, 0);

  if (projection.isError) {
    return (
      <div className="flex flex-col gap-6">
        <h1 className="text-2xl font-semibold" style={{ color: "var(--color-text-primary)" }}>
          Projections
        </h1>
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <h2 className="mb-2 text-lg font-semibold" style={{ color: "var(--color-text-primary)" }}>
            Something went wrong
          </h2>
          <p className="mb-6 max-w-sm text-sm" style={{ color: "var(--color-text-secondary)" }}>
            Could not load projections. Please try again.
          </p>
          <button
            onClick={() => projection.refetch()}
            className="rounded-md px-4 py-2 text-sm font-medium transition-all"
            style={{
              backgroundColor: "var(--color-accent)",
              color: "var(--color-background)",
              boxShadow: "var(--glow-accent)",
            }}
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold" style={{ color: "var(--color-text-primary)" }}>
        Projections
      </h1>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-[320px_1fr]">
        <ScenarioSliders
          params={params}
          onChange={setParams}
          baseMonthlySavings={baseMonthlySavings}
        />

        <div className="flex flex-col gap-6">
          {!projection.data ? (
            <ProjectionSkeleton />
          ) : (
            <>
              <ProjectionChart data={projection.data} />
              <ProjectionSummary data={projection.data} />
            </>
          )}
        </div>
      </div>
    </div>
  );
}
