import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { userEvent } from "@testing-library/user-event";
import SpendingBreakdown from "./SpendingBreakdown";

const mockCategories = [
  { category: "Food", amount: 800, percentage: 40 },
  { category: "Housing", amount: 1000, percentage: 50 },
  { category: "Transport", amount: 200, percentage: 10 },
];

const defaultProps = {
  categories: mockCategories,
  period: "30d",
  onPeriodChange: vi.fn(),
};

describe("SpendingBreakdown", () => {
  it("renders the title", () => {
    render(<SpendingBreakdown {...defaultProps} />);

    expect(screen.getByText("Spending Breakdown")).toBeInTheDocument();
  });

  it("renders category names and amounts", () => {
    render(<SpendingBreakdown {...defaultProps} />);

    expect(screen.getAllByText("Food").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("$800.00").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("Housing").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("Transport").length).toBeGreaterThanOrEqual(1);
  });

  it("renders percentages", () => {
    render(<SpendingBreakdown {...defaultProps} />);

    expect(screen.getAllByText("40.0%").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("50.0%").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("10.0%").length).toBeGreaterThanOrEqual(1);
  });

  it("renders total spent amount", () => {
    render(<SpendingBreakdown {...defaultProps} />);

    expect(screen.getByText("$2,000.00")).toBeInTheDocument();
    expect(screen.getByText("total spent")).toBeInTheDocument();
  });

  it("renders empty state when no categories", () => {
    render(<SpendingBreakdown {...defaultProps} categories={[]} />);

    expect(screen.getByText("No spending data available")).toBeInTheDocument();
  });

  it("renders period toggle buttons", () => {
    render(<SpendingBreakdown {...defaultProps} />);

    expect(screen.getByText("7D")).toBeInTheDocument();
    expect(screen.getByText("30D")).toBeInTheDocument();
    expect(screen.getByText("90D")).toBeInTheDocument();
    expect(screen.getByText("1Y")).toBeInTheDocument();
  });

  it("calls onPeriodChange when a period button is clicked", async () => {
    const onPeriodChange = vi.fn();
    render(<SpendingBreakdown {...defaultProps} onPeriodChange={onPeriodChange} />);

    await userEvent.click(screen.getByText("7D"));

    expect(onPeriodChange).toHaveBeenCalledWith("7d");
  });

  it("renders a screen-reader data table with category details", () => {
    render(<SpendingBreakdown {...defaultProps} />);

    const table = screen.getByRole("table", { name: "Spending breakdown by category" });
    expect(table).toBeInTheDocument();
    expect(table.style.overflow).toBe("hidden");

    const rows = screen.getAllByRole("row");
    expect(rows.length).toBe(mockCategories.length + 1);
  });
});
