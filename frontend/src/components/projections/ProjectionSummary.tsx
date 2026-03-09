import { formatCurrency } from "@/utils/formatting";
import type { ProjectionResponse, Scenario } from "@/api/client";

interface ProjectionSummaryProps {
  data: ProjectionResponse;
}

const SCENARIO_LABELS: Record<Scenario, string> = {
  best: "Best Case",
  average: "Average",
  worst: "Worst Case",
};

const SCENARIO_COLORS: Record<Scenario, string> = {
  best: "#22c55e",
  average: "#3b82f6",
  worst: "#f59e0b",
};

function formatMillionaireDate(dateStr: string | undefined): string {
  if (!dateStr) return "--";
  const d = new Date(dateStr + "T00:00:00");
  return d.toLocaleDateString("en-US", { month: "short", year: "numeric" });
}

function getEndValue(points: Array<{ date: string; value: number }>): number | undefined {
  const last = points[points.length - 1];
  return last?.value;
}

export default function ProjectionSummary({ data }: ProjectionSummaryProps) {
  if (data.scenarios.length === 0) return null;

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
      {data.scenarios.map((scenario) => {
        const endValue = getEndValue(scenario.data_points);
        const color = SCENARIO_COLORS[scenario.scenario];

        return (
          <div
            key={scenario.scenario}
            className="flex flex-col gap-3 rounded-lg p-5"
            style={{
              backgroundColor: "var(--color-surface)",
              border: "1px solid var(--color-border)",
            }}
          >
            <div className="flex items-center gap-2">
              <div className="h-2 w-2 rounded-full" style={{ backgroundColor: color }} />
              <span
                className="text-sm font-medium"
                style={{ color: "var(--color-text-secondary)" }}
              >
                {SCENARIO_LABELS[scenario.scenario]}
              </span>
            </div>

            <div className="flex flex-col gap-1">
              <span className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
                Projected Net Worth
              </span>
              <span
                className="tabular-nums text-xl font-semibold"
                style={{ color: "var(--color-text-primary)" }}
              >
                {endValue !== undefined ? formatCurrency(endValue) : "--"}
              </span>
            </div>

            <div className="flex flex-col gap-1">
              <span className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
                Millionaire By
              </span>
              <span
                className="tabular-nums text-sm font-medium"
                style={{ color: scenario.millionaire_date ? color : "var(--color-text-secondary)" }}
              >
                {formatMillionaireDate(scenario.millionaire_date)}
              </span>
            </div>
          </div>
        );
      })}
    </div>
  );
}
