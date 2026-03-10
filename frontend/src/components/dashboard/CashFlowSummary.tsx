import { formatCurrency } from "@/utils/formatting";

interface NetWorthBreakdown {
  assets: number;
  cash: number;
  debt: number;
}

interface CashFlowSummaryProps {
  netWorth: number;
  breakdown: NetWorthBreakdown;
  cashFlow: number;
}

function cashFlowColor(value: number): string {
  if (value > 0) return "var(--color-positive)";
  if (value < 0) return "var(--color-negative)";
  return "var(--color-text-primary)";
}

function BreakdownRow({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
        {label}
      </span>
      <span className="tabular-nums text-xs" style={{ color }}>
        {formatCurrency(value)}
      </span>
    </div>
  );
}

export default function CashFlowSummary({ netWorth, breakdown, cashFlow }: CashFlowSummaryProps) {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <div
        className="rounded-lg p-6"
        style={{
          backgroundColor: "var(--color-surface)",
          border: "1px solid var(--color-border)",
          boxShadow: "var(--glow-sm)",
        }}
      >
        <p className="mb-1 text-sm" style={{ color: "var(--color-text-secondary)" }}>
          Net Worth
        </p>
        <p
          className="tabular-nums text-3xl font-semibold"
          style={{ color: "var(--color-text-primary)" }}
        >
          {formatCurrency(netWorth)}
        </p>
        <div className="mt-3 flex flex-col gap-1">
          <BreakdownRow label="Assets" value={breakdown.assets} color="var(--color-text-primary)" />
          <BreakdownRow label="Cash" value={breakdown.cash} color="var(--color-positive)" />
          <BreakdownRow label="Debt" value={-breakdown.debt} color="var(--color-negative)" />
        </div>
      </div>
      <div
        className="rounded-lg p-6"
        style={{
          backgroundColor: "var(--color-surface)",
          border: "1px solid var(--color-border)",
          boxShadow: "var(--glow-sm)",
        }}
      >
        <p className="mb-1 text-sm" style={{ color: "var(--color-text-secondary)" }}>
          Cash Flow This Month
        </p>
        <p
          className="tabular-nums text-3xl font-semibold"
          style={{ color: cashFlowColor(cashFlow) }}
        >
          {cashFlow > 0 && <span>+</span>}
          {formatCurrency(cashFlow)}
        </p>
      </div>
    </div>
  );
}
