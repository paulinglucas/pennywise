import {
  formatCurrency,
  formatRelativeTime,
  formatDate,
  formatPercentage,
} from "@/utils/formatting";

interface DebtSummary {
  account_id: string;
  name: string;
  balance: number;
  monthly_payment: number;
  original_balance?: number;
  payoff_date?: string;
  months_remaining?: number;
}

interface DebtTrackerProps {
  debts: DebtSummary[];
}

function debtProgress(debt: DebtSummary): number | null {
  if (!debt.original_balance || debt.original_balance <= 0) return null;
  const paid = debt.original_balance - debt.balance;
  return Math.max(0, Math.min(100, (paid / debt.original_balance) * 100));
}

function ProgressBar({ percentage }: { percentage: number }) {
  return (
    <div className="mt-3">
      <div className="mb-1 flex items-center justify-between text-xs">
        <span style={{ color: "var(--color-text-secondary)" }}>Paid off</span>
        <span className="tabular-nums" style={{ color: "var(--color-accent)" }}>
          {formatPercentage(percentage)}
        </span>
      </div>
      <div
        className="h-2 w-full overflow-hidden rounded-full"
        style={{ backgroundColor: "var(--color-border)" }}
      >
        <div
          className="h-full rounded-full transition-all"
          style={{
            width: `${percentage}%`,
            backgroundColor: "var(--color-accent)",
            boxShadow: percentage > 0 ? "var(--glow-sm)" : "none",
          }}
        />
      </div>
    </div>
  );
}

function DebtCard({ debt }: { debt: DebtSummary }) {
  const progress = debtProgress(debt);

  return (
    <div
      className="rounded-lg p-4"
      style={{
        backgroundColor: "var(--color-background)",
        border: "1px solid var(--color-border)",
      }}
    >
      <div className="mb-3 flex items-center justify-between">
        <h4 className="text-sm font-medium" style={{ color: "var(--color-text-primary)" }}>
          {debt.name}
        </h4>
        <span className="tabular-nums text-xs" style={{ color: "var(--color-text-secondary)" }}>
          {formatCurrency(debt.monthly_payment)}/mo
        </span>
      </div>
      <p
        className="tabular-nums mb-2 text-lg font-semibold"
        style={{ color: "var(--color-negative)" }}
      >
        {formatCurrency(debt.balance)}
      </p>
      <div className="flex items-center justify-between text-xs">
        {debt.months_remaining !== undefined && (
          <span style={{ color: "var(--color-text-secondary)" }}>
            {formatRelativeTime(debt.months_remaining)}
          </span>
        )}
        {debt.payoff_date && (
          <span style={{ color: "var(--color-text-secondary)" }}>
            Payoff: {formatDate(debt.payoff_date)}
          </span>
        )}
      </div>
      {progress !== null && <ProgressBar percentage={progress} />}
    </div>
  );
}

export default function DebtTracker({ debts }: DebtTrackerProps) {
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
        Debt Tracker
      </h3>
      {debts.length === 0 ? (
        <div
          className="flex h-24 items-center justify-center text-sm"
          style={{ color: "var(--color-text-secondary)" }}
        >
          No debts tracked
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {debts.map((debt) => (
            <DebtCard key={debt.account_id} debt={debt} />
          ))}
        </div>
      )}
    </div>
  );
}
