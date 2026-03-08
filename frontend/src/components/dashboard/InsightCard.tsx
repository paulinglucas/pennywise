import { TrendingUp, TrendingDown, AlertTriangle } from "lucide-react";
import { formatCurrency } from "@/utils/formatting";

interface InsightInput {
  spending: Array<{ category: string; amount: number; percentage: number }>;
  debts: Array<{ name: string; months_remaining?: number; monthly_payment: number }>;
  cashFlow: number;
}

type InsightCardsProps = InsightInput;

export function generateInsights(input: InsightInput): string[] {
  const insights: string[] = [];

  if (input.spending.length > 0) {
    const topCategory = input.spending.reduce((max, cat) => (cat.amount > max.amount ? cat : max));
    insights.push(
      `${topCategory.category} is your largest spending category at ${formatCurrency(topCategory.amount)}`,
    );
  }

  if (input.cashFlow > 0) {
    insights.push(`Cash flow is positive this month at ${formatCurrency(input.cashFlow)}`);
  } else if (input.cashFlow < 0) {
    insights.push(
      `Spending is exceeding income by ${formatCurrency(Math.abs(input.cashFlow))} this month`,
    );
  }

  for (const debt of input.debts) {
    if (debt.months_remaining !== undefined && debt.months_remaining <= 12) {
      insights.push(
        `${debt.name} will be paid off in ${debt.months_remaining} months, freeing up ${formatCurrency(debt.monthly_payment)}/mo`,
      );
    }
  }

  return insights;
}

function insightIcon(text: string) {
  if (text.includes("exceeding")) {
    return (
      <AlertTriangle size={16} style={{ color: "var(--color-negative)" }} aria-hidden="true" />
    );
  }
  if (text.includes("positive") || text.includes("paid off")) {
    return <TrendingUp size={16} style={{ color: "var(--color-positive)" }} aria-hidden="true" />;
  }
  return (
    <TrendingDown size={16} style={{ color: "var(--color-text-secondary)" }} aria-hidden="true" />
  );
}

export default function InsightCards({ spending, debts, cashFlow }: InsightCardsProps) {
  const insights = generateInsights({ spending, debts, cashFlow });

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
        Insights
      </h3>
      {insights.length === 0 ? (
        <p className="text-sm" style={{ color: "var(--color-text-secondary)" }}>
          Add more data to see insights
        </p>
      ) : (
        <div className="flex flex-col gap-3">
          {insights.map((insight) => (
            <div
              key={insight}
              className="flex items-start gap-3 rounded-md p-3"
              style={{ backgroundColor: "var(--color-background)" }}
            >
              <div className="mt-0.5 flex-shrink-0">{insightIcon(insight)}</div>
              <p className="text-sm" style={{ color: "var(--color-text-primary)" }}>
                {insight}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
