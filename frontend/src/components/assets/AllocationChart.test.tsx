import { describe, it, expect } from "vitest";
import { screen } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import AllocationChart from "./AllocationChart";
import type { AllocationEntry } from "@/api/client";

const mockAllocation: AllocationEntry[] = [
  { asset_type: "real_estate", value: 308000, percentage: 58.67 },
  { asset_type: "retirement", value: 185000, percentage: 35.24 },
  { asset_type: "speculative", value: 30000, percentage: 5.71 },
];

describe("AllocationChart", () => {
  it("renders heading", () => {
    renderWithProviders(<AllocationChart allocation={mockAllocation} />);
    expect(screen.getByText("Portfolio Allocation")).toBeInTheDocument();
  });

  it("renders legend entries for each asset type", () => {
    renderWithProviders(<AllocationChart allocation={mockAllocation} />);
    expect(screen.getByText("Real Estate")).toBeInTheDocument();
    expect(screen.getByText("Retirement")).toBeInTheDocument();
    expect(screen.getByText("Speculative")).toBeInTheDocument();
  });

  it("renders empty message when no allocation data", () => {
    renderWithProviders(<AllocationChart allocation={[]} />);
    expect(screen.getByText("No allocation data")).toBeInTheDocument();
  });
});
