import { ResponsiveContainer, AreaChart, Area, XAxis, YAxis, Tooltip } from "recharts";
import type { AllocationResponse } from "@/api/client";
import { formatCurrency } from "@/utils/formatting";
import { assetTypeLabel, assetTypeColor } from "./AssetOverview";

interface AllocationTimelapseProps {
  data: AllocationResponse;
  period: string;
  onPeriodChange: (period: string) => void;
}

const periods = [
  { key: "6m", label: "6M" },
  { key: "1y", label: "1Y" },
  { key: "all", label: "All" },
] as const;

interface FlatSnapshot {
  date: string;
  [assetType: string]: string | number;
}

function flattenSnapshots(data: AllocationResponse): { flat: FlatSnapshot[]; types: string[] } {
  const typeSet = new Set<string>();
  for (const snap of data.snapshots) {
    for (const alloc of snap.allocations) {
      typeSet.add(alloc.asset_type);
    }
  }

  const types = Array.from(typeSet);

  const flat: FlatSnapshot[] = data.snapshots.map((snap) => {
    const row: FlatSnapshot = { date: snap.date };
    for (const t of types) {
      row[t] = 0;
    }
    for (const alloc of snap.allocations) {
      row[alloc.asset_type] = alloc.value;
    }
    return row;
  });

  return { flat, types };
}

function TimelapseTooltip({
  active,
  payload,
  label,
}: {
  active?: boolean;
  payload?: Array<{ name: string; value: number; color: string }>;
  label?: string;
}) {
  if (!active || !payload?.length) return null;

  return (
    <div
      className="rounded-md px-3 py-2 text-sm"
      style={{
        backgroundColor: "var(--color-surface)",
        border: "1px solid var(--color-accent-muted)",
        boxShadow: "var(--glow-md)",
      }}
    >
      <p className="mb-1 font-medium" style={{ color: "var(--color-text-primary)" }}>
        {label}
      </p>
      {payload.map((entry) => (
        <p key={entry.name} className="tabular-nums" style={{ color: entry.color }}>
          {assetTypeLabel(entry.name)}: {formatCurrency(entry.value)}
        </p>
      ))}
    </div>
  );
}

export default function AllocationTimelapse({
  data,
  period,
  onPeriodChange,
}: AllocationTimelapseProps) {
  const { flat, types } = flattenSnapshots(data);

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
          Allocation Over Time
        </h3>
        <div className="flex gap-1">
          {periods.map((p) => (
            <button
              key={p.key}
              onClick={() => onPeriodChange(p.key)}
              className="btn-toggle rounded-md px-3 py-1 text-xs font-medium transition-all"
              style={
                period === p.key
                  ? { backgroundColor: "var(--color-accent-muted)", color: "var(--color-accent)" }
                  : { color: "var(--color-text-secondary)" }
              }
            >
              {p.label}
            </button>
          ))}
        </div>
      </div>
      {flat.length === 0 ? (
        <div
          className="flex h-48 items-center justify-center text-sm"
          style={{ color: "var(--color-text-secondary)" }}
        >
          No historical data available
        </div>
      ) : (
        <ResponsiveContainer width="100%" height={280}>
          <AreaChart data={flat}>
            <XAxis
              dataKey="date"
              tick={{ fill: "var(--color-text-secondary)", fontSize: 11 }}
              axisLine={false}
              tickLine={false}
            />
            <YAxis
              tick={{ fill: "var(--color-text-secondary)", fontSize: 11 }}
              axisLine={false}
              tickLine={false}
              tickFormatter={(val: number) => `$${(val / 1000).toFixed(0)}k`}
            />
            <Tooltip content={<TimelapseTooltip />} />
            {types.map((type) => (
              <Area
                key={type}
                type="monotone"
                dataKey={type}
                stackId="1"
                fill={assetTypeColor(type)}
                stroke={assetTypeColor(type)}
                fillOpacity={0.6}
                animationDuration={600}
              />
            ))}
          </AreaChart>
        </ResponsiveContainer>
      )}
    </div>
  );
}
