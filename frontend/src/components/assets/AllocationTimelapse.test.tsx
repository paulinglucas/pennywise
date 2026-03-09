import { describe, it, expect, vi } from "vitest";
import { screen, fireEvent } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import AllocationTimelapse from "./AllocationTimelapse";
import type { AllocationResponse } from "@/api/client";

const mockData: AllocationResponse = {
  snapshots: [
    {
      date: "2025-01-01",
      allocations: [
        { asset_type: "retirement", value: 100000, percentage: 50 },
        { asset_type: "real_estate", value: 100000, percentage: 50 },
      ],
    },
    {
      date: "2025-06-01",
      allocations: [
        { asset_type: "retirement", value: 120000, percentage: 45 },
        { asset_type: "real_estate", value: 150000, percentage: 55 },
      ],
    },
  ],
};

describe("AllocationTimelapse", () => {
  it("renders heading", () => {
    renderWithProviders(
      <AllocationTimelapse data={mockData} period="1y" onPeriodChange={() => {}} />,
    );
    expect(screen.getByText("Allocation Over Time")).toBeInTheDocument();
  });

  it("renders period toggle buttons", () => {
    renderWithProviders(
      <AllocationTimelapse data={mockData} period="1y" onPeriodChange={() => {}} />,
    );
    expect(screen.getByText("6M")).toBeInTheDocument();
    expect(screen.getByText("1Y")).toBeInTheDocument();
    expect(screen.getByText("All")).toBeInTheDocument();
  });

  it("calls onPeriodChange when toggle is clicked", () => {
    const onChange = vi.fn();
    renderWithProviders(
      <AllocationTimelapse data={mockData} period="1y" onPeriodChange={onChange} />,
    );
    fireEvent.click(screen.getByText("6M"));
    expect(onChange).toHaveBeenCalledWith("6m");
  });

  it("renders empty message when no snapshots", () => {
    renderWithProviders(
      <AllocationTimelapse data={{ snapshots: [] }} period="1y" onPeriodChange={() => {}} />,
    );
    expect(screen.getByText("No historical data available")).toBeInTheDocument();
  });
});
