import type { GoalResponse } from "@/api/client";
import { formatCurrency } from "@/utils/formatting";

interface GoalSummaryBarProps {
  goals: GoalResponse[];
}

export default function GoalSummaryBar({ goals }: GoalSummaryBarProps) {
  const debtGoals = goals.filter((g) => g.goal_type === "debt_payoff");
  const savingsGoals = goals.filter((g) => g.goal_type === "savings");

  const totalDebtRemaining = debtGoals.reduce((sum, g) => sum + g.current_amount, 0);
  const totalSaved = savingsGoals.reduce((sum, g) => sum + g.current_amount, 0);
  const totalSavingsTarget = savingsGoals.reduce((sum, g) => sum + g.target_amount, 0);

  const savingsPercent = totalSavingsTarget > 0 ? (totalSaved / totalSavingsTarget) * 100 : 0;

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <SummaryCard
        label="Debt Remaining"
        value={totalDebtRemaining}
        color="var(--color-negative)"
      />
      <SummaryCard
        label="Savings Progress"
        value={totalSaved}
        target={totalSavingsTarget}
        percent={savingsPercent}
        color="var(--color-positive)"
      />
    </div>
  );
}

function SummaryCard({
  label,
  value,
  target,
  percent,
  color,
}: {
  label: string;
  value: number;
  target?: number;
  percent?: number;
  color: string;
}) {
  return (
    <div
      className="rounded-lg p-5"
      style={{
        backgroundColor: "var(--color-surface)",
        border: "1px solid var(--color-border)",
        boxShadow: "var(--glow-sm)",
      }}
    >
      <p className="text-sm font-medium" style={{ color: "var(--color-text-secondary)" }}>
        {label}
      </p>
      <p className="mt-1 tabular-nums text-2xl font-bold" style={{ color }}>
        {formatCurrency(value)}
      </p>
      {target !== undefined && target > 0 && (
        <div className="mt-2">
          <div className="flex justify-between text-xs">
            <span style={{ color: "var(--color-text-secondary)" }}>
              of {formatCurrency(target)}
            </span>
            <span className="tabular-nums" style={{ color }}>
              {percent?.toFixed(0)}%
            </span>
          </div>
          <div
            className="mt-1 h-2 overflow-hidden rounded-full"
            style={{ backgroundColor: "var(--color-background)" }}
          >
            <div
              className="h-full rounded-full transition-all"
              style={{ width: `${Math.min(percent ?? 0, 100)}%`, backgroundColor: color }}
            />
          </div>
        </div>
      )}
    </div>
  );
}
