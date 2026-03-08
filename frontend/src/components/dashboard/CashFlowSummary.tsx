import { formatCurrency } from "@/utils/formatting";

interface CashFlowSummaryProps {
  netWorth: number;
  cashFlow: number;
}

function cashFlowColor(value: number): string {
  if (value > 0) return "var(--color-positive)";
  if (value < 0) return "var(--color-negative)";
  return "var(--color-text-primary)";
}

export default function CashFlowSummary({ netWorth, cashFlow }: CashFlowSummaryProps) {
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
