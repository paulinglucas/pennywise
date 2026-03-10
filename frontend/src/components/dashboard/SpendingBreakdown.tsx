import { ResponsiveContainer, PieChart, Pie, Cell, Tooltip } from "recharts";
import { formatCurrency, formatPercentage } from "@/utils/formatting";

interface CategorySpending {
  category: string;
  amount: number;
  percentage: number;
}

interface SpendingBreakdownProps {
  categories: CategorySpending[];
  period: string;
  onPeriodChange: (period: string) => void;
}

const spendingPeriods = [
  { key: "7d", label: "7D" },
  { key: "30d", label: "30D" },
  { key: "90d", label: "90D" },
  { key: "1y", label: "1Y" },
] as const;

function SpendingPeriodToggle({
  activePeriod,
  onPeriodChange,
}: {
  activePeriod: string;
  onPeriodChange: (period: string) => void;
}) {
  return (
    <div className="flex gap-1">
      {spendingPeriods.map((period) => (
        <button
          key={period.key}
          onClick={() => onPeriodChange(period.key)}
          className={`btn-toggle rounded-md px-3 py-1 text-xs font-medium transition-all${activePeriod === period.key ? " active" : ""}`}
          style={
            activePeriod === period.key
              ? { backgroundColor: "var(--color-accent-muted)", color: "var(--color-accent)" }
              : { color: "var(--color-text-secondary)" }
          }
        >
          {period.label}
        </button>
      ))}
    </div>
  );
}

const CHART_COLORS = [
  "#22c55e",
  "#4ade80",
  "#86efac",
  "#a78bfa",
  "#f59e0b",
  "#fb923c",
  "#38bdf8",
  "#818cf8",
];

function SpendingTooltip({
  active,
  payload,
}: {
  active?: boolean;
  payload?: Array<{ payload: CategorySpending }>;
}) {
  const firstEntry = payload?.[0];
  if (!active || !firstEntry) return null;

  const data = firstEntry.payload;
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
        {data.category}
      </p>
      <p className="tabular-nums" style={{ color: "var(--color-accent)" }}>
        {formatCurrency(data.amount)} ({formatPercentage(data.percentage)})
      </p>
    </div>
  );
}

function CategoryLegend({ categories }: { categories: CategorySpending[] }) {
  return (
    <div className="flex flex-col gap-2">
      {categories.map((category, index) => (
        <div key={category.category} className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <div
              className="h-3 w-3 rounded-full"
              style={{ backgroundColor: CHART_COLORS[index % CHART_COLORS.length] }}
            />
            <span className="text-sm" style={{ color: "var(--color-text-secondary)" }}>
              {category.category}
            </span>
          </div>
          <div className="flex items-center gap-3">
            <span className="tabular-nums text-sm" style={{ color: "var(--color-text-primary)" }}>
              {formatCurrency(category.amount)}
            </span>
            <span className="tabular-nums text-xs" style={{ color: "var(--color-text-secondary)" }}>
              {formatPercentage(category.percentage)}
            </span>
          </div>
        </div>
      ))}
    </div>
  );
}

export default function SpendingBreakdown({
  categories,
  period,
  onPeriodChange,
}: SpendingBreakdownProps) {
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
          Spending Breakdown
        </h3>
        <SpendingPeriodToggle activePeriod={period} onPeriodChange={onPeriodChange} />
      </div>
      {categories.length === 0 ? (
        <div
          className="flex h-48 items-center justify-center text-sm"
          style={{ color: "var(--color-text-secondary)" }}
        >
          No spending data available
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          <div
            className="tabular-nums text-2xl font-bold"
            style={{ color: "var(--color-text-primary)" }}
          >
            {formatCurrency(categories.reduce((sum, c) => sum + c.amount, 0))}
            <span
              className="ml-2 text-sm font-normal"
              style={{ color: "var(--color-text-secondary)" }}
            >
              total spent
            </span>
          </div>
          <div className="flex flex-col gap-4 lg:flex-row lg:items-center">
            <div className="aspect-square w-full max-w-[260px] flex-shrink-0" aria-hidden="true">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={categories}
                    dataKey="amount"
                    nameKey="category"
                    cx="50%"
                    cy="50%"
                    innerRadius="45%"
                    outerRadius="75%"
                    strokeWidth={0}
                    animationDuration={600}
                  >
                    {categories.map((entry, index) => (
                      <Cell key={entry.category} fill={CHART_COLORS[index % CHART_COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip content={<SpendingTooltip />} />
                </PieChart>
              </ResponsiveContainer>
            </div>
            <div className="flex-1">
              <CategoryLegend categories={categories} />
            </div>
          </div>
          <table
            style={{
              position: "absolute",
              width: 1,
              height: 1,
              padding: 0,
              margin: -1,
              overflow: "hidden",
              clip: "rect(0,0,0,0)",
              whiteSpace: "nowrap",
              borderWidth: 0,
            }}
          >
            <caption>Spending breakdown by category</caption>
            <thead>
              <tr>
                <th scope="col">Category</th>
                <th scope="col">Amount</th>
                <th scope="col">Percentage</th>
              </tr>
            </thead>
            <tbody>
              {categories.map((cat) => (
                <tr key={cat.category}>
                  <td>{cat.category}</td>
                  <td>{formatCurrency(cat.amount)}</td>
                  <td>{formatPercentage(cat.percentage)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
