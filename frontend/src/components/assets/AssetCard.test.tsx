import { describe, it, expect } from "vitest";
import { screen } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import AssetCard from "./AssetCard";
import type { AssetResponse } from "@/api/client";

const baseAsset: AssetResponse = {
  id: "a1",
  user_id: "u1",
  name: "Home",
  asset_type: "real_estate",
  current_value: 308000,
  currency: "USD",
  created_at: "2025-01-01T00:00:00Z",
  updated_at: "2025-06-01T00:00:00Z",
};

describe("AssetCard", () => {
  it("renders asset name and value", () => {
    renderWithProviders(<AssetCard asset={baseAsset} portfolioTotal={500000} onClick={() => {}} />);
    expect(screen.getByText("Home")).toBeInTheDocument();
    expect(screen.getByText("$308,000.00")).toBeInTheDocument();
  });

  it("renders asset type badge", () => {
    renderWithProviders(<AssetCard asset={baseAsset} portfolioTotal={500000} onClick={() => {}} />);
    expect(screen.getByText("Real Estate")).toBeInTheDocument();
  });

  it("renders portfolio percentage", () => {
    renderWithProviders(<AssetCard asset={baseAsset} portfolioTotal={500000} onClick={() => {}} />);
    expect(screen.getByText("61.6%")).toBeInTheDocument();
  });

  it("renders SAFE details when metadata is present", () => {
    const safeAsset: AssetResponse = {
      ...baseAsset,
      name: "Reve Studios SAFE",
      asset_type: "speculative",
      current_value: 30000,
      metadata: {
        company_name: "Reve Studios",
        ownership_percentage: 1.0,
        valuation_cap: 3000000,
      },
    };
    renderWithProviders(<AssetCard asset={safeAsset} portfolioTotal={500000} onClick={() => {}} />);
    expect(screen.getByText("Reve Studios SAFE")).toBeInTheDocument();
    expect(screen.getByText("Illiquid - Speculative")).toBeInTheDocument();
  });

  it("renders real estate equity info when metadata is present", () => {
    const realEstateAsset: AssetResponse = {
      ...baseAsset,
      metadata: {
        purchase_price: 301000,
        down_payment_percent: 5,
        mortgage_rate: 6.875,
        mortgage_term_years: 30,
        current_valuation: 308000,
        monthly_payment: 2500,
        extra_principal_monthly: 0,
        purchase_date: "2023-04-01",
      },
    };
    renderWithProviders(
      <AssetCard asset={realEstateAsset} portfolioTotal={500000} onClick={() => {}} />,
    );
    expect(screen.getByText(/Equity/)).toBeInTheDocument();
  });
});
