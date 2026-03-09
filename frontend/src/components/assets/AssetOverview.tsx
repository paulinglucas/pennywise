import type { PortfolioSummary } from "@/api/client";
import { formatCurrency, formatPercentage } from "@/utils/formatting";

interface AssetOverviewProps {
  summary: PortfolioSummary;
}

const ASSET_TYPE_LABELS: Record<string, string> = {
  liquid: "Liquid",
  retirement: "Retirement",
  real_estate: "Real Estate",
  brokerage: "Brokerage",
  speculative: "Speculative",
  other: "Other",
};

const ASSET_TYPE_COLORS: Record<string, string> = {
  liquid: "#38bdf8",
  retirement: "#22c55e",
  real_estate: "#a78bfa",
  brokerage: "#f59e0b",
  speculative: "#fb923c",
  other: "#818cf8",
};

export function assetTypeLabel(type: string): string {
  return ASSET_TYPE_LABELS[type] ?? type;
}

export function assetTypeColor(type: string): string {
  return ASSET_TYPE_COLORS[type] ?? "#818cf8";
}

export default function AssetOverview({ summary }: AssetOverviewProps) {
  return (
    <div
      className="rounded-lg p-6"
      style={{
        backgroundColor: "var(--color-surface)",
        border: "1px solid var(--color-border)",
        boxShadow: "var(--glow-sm)",
      }}
    >
      <p className="text-sm font-medium" style={{ color: "var(--color-text-secondary)" }}>
        Total Portfolio Value
      </p>
      <p className="mt-1 tabular-nums text-3xl font-bold" style={{ color: "var(--color-accent)" }}>
        {formatCurrency(summary.total_value)}
      </p>
      {summary.allocation.length > 0 && (
        <div className="mt-4 flex flex-col gap-2">
          {summary.allocation.map((entry) => (
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
              <div className="flex items-center gap-3">
                <span
                  className="tabular-nums text-sm"
                  style={{ color: "var(--color-text-primary)" }}
                >
                  {formatCurrency(entry.value)}
                </span>
                <span
                  className="tabular-nums text-xs"
                  style={{ color: "var(--color-text-secondary)" }}
                >
                  {formatPercentage(entry.percentage)}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
