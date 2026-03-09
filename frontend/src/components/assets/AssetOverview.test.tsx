import { describe, it, expect } from "vitest";
import { screen } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import AssetOverview from "./AssetOverview";
import type { PortfolioSummary } from "@/api/client";

const mockSummary: PortfolioSummary = {
  total_value: 525000,
  allocation: [
    { asset_type: "real_estate", value: 308000, percentage: 58.67 },
    { asset_type: "retirement", value: 185000, percentage: 35.24 },
    { asset_type: "speculative", value: 30000, percentage: 5.71 },
    { asset_type: "liquid", value: 2000, percentage: 0.38 },
  ],
};

describe("AssetOverview", () => {
  it("renders total portfolio value", () => {
    renderWithProviders(<AssetOverview summary={mockSummary} />);
    expect(screen.getByText("$525,000.00")).toBeInTheDocument();
  });

  it("renders portfolio value label", () => {
    renderWithProviders(<AssetOverview summary={mockSummary} />);
    expect(screen.getByText("Total Portfolio Value")).toBeInTheDocument();
  });

  it("renders allocation entries", () => {
    renderWithProviders(<AssetOverview summary={mockSummary} />);
    expect(screen.getByText("Real Estate")).toBeInTheDocument();
    expect(screen.getByText("Retirement")).toBeInTheDocument();
    expect(screen.getByText("Speculative")).toBeInTheDocument();
    expect(screen.getByText("Liquid")).toBeInTheDocument();
  });

  it("renders zero state when total is zero", () => {
    const empty: PortfolioSummary = { total_value: 0, allocation: [] };
    renderWithProviders(<AssetOverview summary={empty} />);
    expect(screen.getByText("$0.00")).toBeInTheDocument();
  });
});
