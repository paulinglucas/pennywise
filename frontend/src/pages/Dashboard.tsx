import { useState } from "react";
import { useDashboard, useNetWorthHistory, type SpendingPeriod } from "@/hooks/useDashboard";
import CashFlowSummary from "@/components/dashboard/CashFlowSummary";
import NetWorthChart from "@/components/dashboard/NetWorthChart";
import SpendingBreakdown from "@/components/dashboard/SpendingBreakdown";
import DebtTracker from "@/components/dashboard/DebtTracker";
import InsightCards from "@/components/dashboard/InsightCard";
import EmptyState from "@/components/shared/EmptyState";
import { SkeletonCard, SkeletonChart } from "@/components/shared/Skeleton";

function DashboardSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <SkeletonCard />
        <SkeletonCard />
      </div>
      <SkeletonChart />
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <SkeletonChart />
        <SkeletonChart />
      </div>
      <SkeletonChart />
    </div>
  );
}

function DashboardError({ onRetry }: { onRetry: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <h2 className="mb-2 text-lg font-semibold" style={{ color: "var(--color-text-primary)" }}>
        Something went wrong
      </h2>
      <p className="mb-6 max-w-sm text-sm" style={{ color: "var(--color-text-secondary)" }}>
        We could not load your dashboard data. Please try again.
      </p>
      <button
        onClick={onRetry}
        className="rounded-md px-4 py-2 text-sm font-medium transition-all"
        style={{
          backgroundColor: "var(--color-accent)",
          color: "var(--color-background)",
          boxShadow: "var(--glow-accent)",
        }}
      >
        Retry
      </button>
    </div>
  );
}

export default function Dashboard() {
  const [period, setPeriod] = useState("1y");
  const [spendingPeriod, setSpendingPeriod] = useState<SpendingPeriod>("30d");
  const dashboard = useDashboard(spendingPeriod);
  const history = useNetWorthHistory(period as "1m" | "1y" | "5y" | "all");

  if (dashboard.isError) {
    return <DashboardError onRetry={() => dashboard.refetch()} />;
  }

  const data = dashboard.data;

  if (!data) {
    return <DashboardSkeleton />;
  }

  const isEmpty =
    data.spending_by_category.length === 0 &&
    data.debts_summary.length === 0 &&
    data.net_worth === 0;

  if (isEmpty) {
    return (
      <div className="flex flex-col gap-6">
        <CashFlowSummary netWorth={data.net_worth} cashFlow={data.cash_flow_this_month} />
        <SpendingBreakdown
          categories={data.spending_by_category}
          period={spendingPeriod ?? "30d"}
          onPeriodChange={(p) => setSpendingPeriod(p as SpendingPeriod)}
        />
        <EmptyState
          title="No data yet"
          description="Link your accounts or add transactions to get started."
          actionLabel="Add Transaction"
          actionTo="/transactions"
        />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <CashFlowSummary netWorth={data.net_worth} cashFlow={data.cash_flow_this_month} />
      <NetWorthChart
        dataPoints={history.data?.data_points ?? []}
        period={period}
        onPeriodChange={setPeriod}
      />
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <SpendingBreakdown
          categories={data.spending_by_category}
          period={spendingPeriod ?? "30d"}
          onPeriodChange={(p) => setSpendingPeriod(p as SpendingPeriod)}
        />
        <DebtTracker debts={data.debts_summary} />
      </div>
      <InsightCards
        spending={data.spending_by_category}
        debts={data.debts_summary}
        cashFlow={data.cash_flow_this_month}
      />
    </div>
  );
}
