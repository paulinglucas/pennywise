import type { AssetResponse } from "@/api/client";
import { formatCurrency, formatPercentage } from "@/utils/formatting";
import { assetTypeLabel, assetTypeColor } from "./AssetOverview";
import { currentEquity, safeScenarios } from "@/utils/calculations";

interface AssetCardProps {
  asset: AssetResponse;
  portfolioTotal: number;
  onClick: () => void;
}

interface RealEstateMetadata {
  purchase_price: number;
  down_payment_percent: number;
  mortgage_rate: number;
  mortgage_term_years: number;
  current_valuation: number;
  monthly_payment: number;
  extra_principal_monthly: number;
  purchase_date: string;
}

interface SafeMetadata {
  company_name: string;
  ownership_percentage: number;
  valuation_cap: number;
}

function monthsSince(dateStr: string): number {
  const start = new Date(dateStr);
  const now = new Date();
  return (now.getFullYear() - start.getFullYear()) * 12 + (now.getMonth() - start.getMonth());
}

function RealEstateDetails({ metadata }: { metadata: RealEstateMetadata }) {
  const equity = currentEquity({
    purchasePrice: metadata.purchase_price,
    downPaymentPercent: metadata.down_payment_percent,
    annualRate: metadata.mortgage_rate,
    termYears: metadata.mortgage_term_years,
    extraMonthly: metadata.extra_principal_monthly,
    monthsElapsed: monthsSince(metadata.purchase_date),
    currentValuation: metadata.current_valuation,
  });

  return (
    <div className="mt-2 flex flex-col gap-1 text-xs">
      <div className="flex justify-between">
        <span style={{ color: "var(--color-text-secondary)" }}>Equity</span>
        <span className="tabular-nums" style={{ color: "var(--color-positive)" }}>
          {formatCurrency(equity.currentEquity)}
        </span>
      </div>
      <div className="flex justify-between">
        <span style={{ color: "var(--color-text-secondary)" }}>Loan Balance</span>
        <span className="tabular-nums" style={{ color: "var(--color-text-primary)" }}>
          {formatCurrency(equity.loanBalance)}
        </span>
      </div>
      <div className="flex justify-between">
        <span style={{ color: "var(--color-text-secondary)" }}>Monthly Payment</span>
        <span className="tabular-nums" style={{ color: "var(--color-text-primary)" }}>
          {formatCurrency(metadata.monthly_payment)}
        </span>
      </div>
    </div>
  );
}

function SafeDetails({ metadata }: { metadata: SafeMetadata }) {
  const scenarios = safeScenarios({
    ownershipPercentage: metadata.ownership_percentage,
    valuationCap: metadata.valuation_cap,
  });

  return (
    <div className="mt-2 flex flex-col gap-1 text-xs">
      <span
        className="inline-block rounded-full px-2 py-0.5 text-xs font-medium"
        style={{
          backgroundColor: "var(--color-negative)",
          color: "var(--color-background)",
        }}
      >
        Illiquid - Speculative
      </span>
      <div className="mt-1 flex justify-between">
        <span style={{ color: "var(--color-text-secondary)" }}>Ownership</span>
        <span className="tabular-nums" style={{ color: "var(--color-text-primary)" }}>
          {formatPercentage(metadata.ownership_percentage)}
        </span>
      </div>
      <div className="flex justify-between">
        <span style={{ color: "var(--color-text-secondary)" }}>Valuation Cap</span>
        <span className="tabular-nums" style={{ color: "var(--color-text-primary)" }}>
          {formatCurrency(metadata.valuation_cap)}
        </span>
      </div>
      <div className="mt-2 border-t pt-2" style={{ borderColor: "var(--color-border)" }}>
        <p className="mb-1 font-medium" style={{ color: "var(--color-text-secondary)" }}>
          Scenario Table
        </p>
        {scenarios.map((s) => (
          <div key={s.valuation} className="flex justify-between">
            <span style={{ color: "var(--color-text-secondary)" }}>
              @{formatCurrency(s.valuation)}
            </span>
            <span className="tabular-nums" style={{ color: "var(--color-accent)" }}>
              {formatCurrency(s.ownershipValue)}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

function hasRealEstateFields(meta: Record<string, unknown>): boolean {
  return "purchase_price" in meta && "mortgage_rate" in meta;
}

function hasSafeFields(meta: Record<string, unknown>): boolean {
  return "ownership_percentage" in meta && "valuation_cap" in meta;
}

export default function AssetCard({ asset, portfolioTotal, onClick }: AssetCardProps) {
  const percentage = portfolioTotal > 0 ? (asset.current_value / portfolioTotal) * 100 : 0;
  const meta = asset.metadata as Record<string, unknown> | undefined;

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") onClick();
      }}
      className="cursor-pointer rounded-lg p-5 transition-all"
      style={{
        backgroundColor: "var(--color-surface)",
        border: "1px solid var(--color-border)",
      }}
    >
      <div className="flex items-start justify-between">
        <div className="min-w-0 flex-1">
          <p className="truncate font-medium" style={{ color: "var(--color-text-primary)" }}>
            {asset.name}
          </p>
          <span
            className="mt-1 inline-block rounded-full px-2 py-0.5 text-xs font-medium"
            style={{
              backgroundColor: assetTypeColor(asset.asset_type) + "22",
              color: assetTypeColor(asset.asset_type),
            }}
          >
            {assetTypeLabel(asset.asset_type)}
          </span>
        </div>
        <div className="text-right">
          <p
            className="tabular-nums text-lg font-semibold"
            style={{ color: "var(--color-text-primary)" }}
          >
            {formatCurrency(asset.current_value)}
          </p>
          <p className="tabular-nums text-xs" style={{ color: "var(--color-text-secondary)" }}>
            {formatPercentage(percentage)}
          </p>
        </div>
      </div>
      {meta && hasRealEstateFields(meta) && (
        <RealEstateDetails metadata={meta as unknown as RealEstateMetadata} />
      )}
      {meta && hasSafeFields(meta) && <SafeDetails metadata={meta as unknown as SafeMetadata} />}
    </div>
  );
}
