import { useState } from "react";
import { useProjections, DEFAULT_PARAMS, type ProjectionParams } from "@/hooks/useProjections";
import { useDashboard } from "@/hooks/useDashboard";
import ScenarioSliders from "@/components/projections/ScenarioSliders";
import ProjectionChart from "@/components/projections/ProjectionChart";
import ProjectionSummary from "@/components/projections/ProjectionSummary";
import EmptyState from "@/components/shared/EmptyState";
import ErrorState, { extractRequestId } from "@/components/shared/ErrorState";
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

function hasFinancialData(dashboard: { net_worth: number; cash_flow_this_month: number }): boolean {
  return dashboard.net_worth !== 0 || dashboard.cash_flow_this_month !== 0;
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
        <ErrorState
          message="Could not load projections. Please try again."
          onRetry={() => projection.refetch()}
          requestId={extractRequestId(projection.error)}
        />
      </div>
    );
  }

  if (dashboard.data && !hasFinancialData(dashboard.data)) {
    return (
      <div className="flex flex-col gap-6">
        <h1 className="text-2xl font-semibold" style={{ color: "var(--color-text-primary)" }}>
          Projections
        </h1>
        <EmptyState
          title="No financial data yet"
          description="Add some financial data first to see your projections."
          actionLabel="Go to Dashboard"
          actionTo="/"
        />
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
