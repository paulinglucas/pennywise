import { ResponsiveContainer, PieChart, Pie, Cell, Tooltip } from "recharts";
import type { AllocationEntry } from "@/api/client";
import { formatCurrency, formatPercentage } from "@/utils/formatting";
import { assetTypeLabel, assetTypeColor } from "./AssetOverview";

interface AllocationChartProps {
  allocation: AllocationEntry[];
}

function AllocationTooltip({
  active,
  payload,
}: {
  active?: boolean;
  payload?: Array<{ payload: AllocationEntry }>;
}) {
  const entry = payload?.[0];
  if (!active || !entry) return null;

  const data = entry.payload;
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
        {assetTypeLabel(data.asset_type)}
      </p>
      <p className="tabular-nums" style={{ color: "var(--color-accent)" }}>
        {formatCurrency(data.value)} ({formatPercentage(data.percentage)})
      </p>
    </div>
  );
}

export default function AllocationChart({ allocation }: AllocationChartProps) {
  return (
    <div
      className="rounded-lg p-6"
      style={{
        backgroundColor: "var(--color-surface)",
        border: "1px solid var(--color-border)",
        boxShadow: "var(--glow-sm)",
      }}
    >
      <h3 className="mb-4 text-sm font-medium" style={{ color: "var(--color-text-secondary)" }}>
        Portfolio Allocation
      </h3>
      {allocation.length === 0 ? (
        <div
          className="flex h-48 items-center justify-center text-sm"
          style={{ color: "var(--color-text-secondary)" }}
        >
          No allocation data
        </div>
      ) : (
        <>
          <div className="flex flex-col items-center gap-4 lg:flex-row lg:items-start">
            <div className="flex-shrink-0" aria-hidden="true">
              <ResponsiveContainer width={180} height={180}>
                <PieChart>
                  <Pie
                    data={allocation}
                    dataKey="value"
                    nameKey="asset_type"
                    cx="50%"
                    cy="50%"
                    innerRadius={50}
                    outerRadius={80}
                    strokeWidth={0}
                    animationDuration={600}
                  >
                    {allocation.map((entry) => (
                      <Cell key={entry.asset_type} fill={assetTypeColor(entry.asset_type)} />
                    ))}
                  </Pie>
                  <Tooltip content={<AllocationTooltip />} />
                </PieChart>
              </ResponsiveContainer>
            </div>
            <div className="flex flex-1 flex-col gap-2">
              {allocation.map((entry) => (
                <div key={entry.asset_type} className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <div
                      className="h-3 w-3 rounded-full"
                      style={{ backgroundColor: assetTypeColor(entry.asset_type) }}
                    />
                    <span className="text-sm" style={{ color: "var(--color-text-secondary)" }}>
                      {assetTypeLabel(entry.asset_type)}
                    </span>
                  </div>
                  <span
                    className="tabular-nums text-sm"
                    style={{ color: "var(--color-text-primary)" }}
                  >
                    {formatPercentage(entry.percentage)}
                  </span>
                </div>
              ))}
            </div>
          </div>
          <table className="sr-only">
            <caption>Portfolio allocation by asset type</caption>
            <thead>
              <tr>
                <th scope="col">Asset Type</th>
                <th scope="col">Value</th>
                <th scope="col">Percentage</th>
              </tr>
            </thead>
            <tbody>
              {allocation.map((entry) => (
                <tr key={entry.asset_type}>
                  <td>{assetTypeLabel(entry.asset_type)}</td>
                  <td>{formatCurrency(entry.value)}</td>
                  <td>{formatPercentage(entry.percentage)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      )}
    </div>
  );
}
