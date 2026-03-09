import {
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ReferenceLine,
} from "recharts";
import { formatCurrency } from "@/utils/formatting";
import type { ProjectionResponse, Scenario } from "@/api/client";

interface ProjectionChartProps {
  data: ProjectionResponse;
}

interface MergedDataPoint {
  timestamp: number;
  date: string;
  best?: number;
  average?: number;
  worst?: number;
}

const SCENARIO_COLORS: Record<Scenario, string> = {
  best: "#22c55e",
  average: "#3b82f6",
  worst: "#f59e0b",
};

const SCENARIO_LABELS: Record<Scenario, string> = {
  best: "Best Case",
  average: "Average",
  worst: "Worst Case",
};

interface Milestone {
  value: number;
  label: string;
}

const MILESTONES: Milestone[] = [
  { value: 500_000, label: "$500K" },
  { value: 1_000_000, label: "$1M" },
  { value: 2_000_000, label: "$2M" },
  { value: 5_000_000, label: "$5M" },
];

function mergeScenarioData(data: ProjectionResponse): MergedDataPoint[] {
  const dateMap = new Map<string, MergedDataPoint>();

  for (const scenario of data.scenarios) {
    for (const point of scenario.data_points) {
      const existing = dateMap.get(point.date);
      if (existing) {
        existing[scenario.scenario] = point.value;
      } else {
        dateMap.set(point.date, {
          timestamp: new Date(point.date + "T00:00:00").getTime(),
          date: point.date,
          [scenario.scenario]: point.value,
        });
      }
    }
  }

  return Array.from(dateMap.values()).sort((a, b) => a.timestamp - b.timestamp);
}

function getMaxValue(data: ProjectionResponse): number {
  let max = 0;
  for (const scenario of data.scenarios) {
    for (const point of scenario.data_points) {
      if (point.value > max) max = point.value;
    }
  }
  return max;
}

function getVisibleMilestones(maxValue: number): Milestone[] {
  return MILESTONES.filter((m) => m.value <= maxValue * 1.1);
}

function formatAxisDate(ts: number): string {
  const d = new Date(ts);
  return d.toLocaleDateString("en-US", { month: "short", year: "numeric" });
}

function formatAxisValue(val: number): string {
  if (val >= 1_000_000) return `$${(val / 1_000_000).toFixed(1)}M`;
  if (val >= 1_000) return `$${(val / 1_000).toFixed(0)}K`;
  return formatCurrency(val);
}

function ChartTooltip({
  active,
  payload,
}: {
  active?: boolean;
  payload?: Array<{ name: string; value: number; color: string; payload: MergedDataPoint }>;
  label?: number;
}) {
  const firstEntry = payload?.[0];
  if (!active || !firstEntry) return null;

  const dateStr = firstEntry.payload.date;
  const formatted = new Date(dateStr + "T00:00:00").toLocaleDateString("en-US", {
    month: "long",
    year: "numeric",
  });

  return (
    <div
      className="rounded-md px-3 py-2 text-sm"
      style={{
        backgroundColor: "var(--color-surface)",
        border: "1px solid var(--color-border)",
        boxShadow: "var(--glow-md)",
      }}
    >
      <p className="mb-1 text-xs" style={{ color: "var(--color-text-secondary)" }}>
        {formatted}
      </p>
      {payload?.map((entry) => (
        <p key={entry.name} className="tabular-nums text-xs" style={{ color: entry.color }}>
          {SCENARIO_LABELS[entry.name as Scenario] ?? entry.name}: {formatCurrency(entry.value)}
        </p>
      ))}
    </div>
  );
}

function Legend() {
  const scenarios: Scenario[] = ["best", "average", "worst"];
  return (
    <div className="flex gap-4">
      {scenarios.map((s) => (
        <div key={s} className="flex items-center gap-1.5">
          <div className="h-2 w-2 rounded-full" style={{ backgroundColor: SCENARIO_COLORS[s] }} />
          <span className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
            {SCENARIO_LABELS[s]}
          </span>
        </div>
      ))}
    </div>
  );
}

function ScenarioLines() {
  const scenarios: Scenario[] = ["best", "average", "worst"];
  return (
    <>
      {scenarios.map((s) => (
        <Line
          key={s}
          type="monotone"
          dataKey={s}
          stroke={SCENARIO_COLORS[s]}
          strokeWidth={2}
          dot={false}
          animationDuration={600}
        />
      ))}
    </>
  );
}

function MilestoneLines({ milestones }: { milestones: Milestone[] }) {
  return (
    <>
      {milestones.map((m) => (
        <ReferenceLine
          key={m.value}
          y={m.value}
          stroke="var(--color-border)"
          strokeDasharray="4 4"
          label={{
            value: m.label,
            position: "right",
            fill: "var(--color-text-secondary)",
            fontSize: 11,
          }}
        />
      ))}
    </>
  );
}

export default function ProjectionChart({ data }: ProjectionChartProps) {
  if (data.scenarios.length === 0) {
    return (
      <div
        className="flex h-64 items-center justify-center rounded-lg text-sm"
        style={{
          backgroundColor: "var(--color-surface)",
          border: "1px solid var(--color-border)",
          color: "var(--color-text-secondary)",
        }}
      >
        No projection data available
      </div>
    );
  }

  const merged = mergeScenarioData(data);
  const maxValue = getMaxValue(data);
  const milestones = getVisibleMilestones(maxValue);

  return (
    <div
      className="rounded-lg p-6"
      style={{
        backgroundColor: "var(--color-surface)",
        border: "1px solid var(--color-border)",
        boxShadow: "var(--glow-sm)",
      }}
    >
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-sm font-medium" style={{ color: "var(--color-text-secondary)" }}>
          Net Worth Projection
        </h3>
        <Legend />
      </div>
      <ResponsiveContainer width="100%" height={350}>
        <LineChart data={merged}>
          <XAxis
            dataKey="timestamp"
            type="number"
            scale="time"
            domain={["dataMin", "dataMax"]}
            tickFormatter={formatAxisDate}
            tick={{ fill: "var(--color-text-secondary)", fontSize: 11 }}
            axisLine={false}
            tickLine={false}
          />
          <YAxis
            tickFormatter={formatAxisValue}
            tick={{ fill: "var(--color-text-secondary)", fontSize: 11 }}
            axisLine={false}
            tickLine={false}
            width={70}
          />
          <Tooltip content={<ChartTooltip />} />
          <MilestoneLines milestones={milestones} />
          <ScenarioLines />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
