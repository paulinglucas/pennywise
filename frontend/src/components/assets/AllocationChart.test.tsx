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
    expect(screen.getAllByText("Real Estate").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("Retirement").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("Speculative").length).toBeGreaterThanOrEqual(1);
  });

  it("renders empty message when no allocation data", () => {
    renderWithProviders(<AllocationChart allocation={[]} />);
    expect(screen.getByText("No allocation data")).toBeInTheDocument();
  });

  it("renders a screen-reader data table with allocation details", () => {
    renderWithProviders(<AllocationChart allocation={mockAllocation} />);

    const table = screen.getByRole("table", { name: "Portfolio allocation by asset type" });
    expect(table).toBeInTheDocument();
    expect((table as HTMLElement).style.overflow).toBe("hidden");

    const rows = screen.getAllByRole("row");
    expect(rows.length).toBe(mockAllocation.length + 1);
  });
});
