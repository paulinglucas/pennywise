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

    expect(screen.getByText("Food")).toBeInTheDocument();
    expect(screen.getByText("$800.00")).toBeInTheDocument();
    expect(screen.getByText("Housing")).toBeInTheDocument();
    expect(screen.getByText("Transport")).toBeInTheDocument();
  });

  it("renders percentages", () => {
    render(<SpendingBreakdown {...defaultProps} />);

    expect(screen.getByText("40.0%")).toBeInTheDocument();
    expect(screen.getByText("50.0%")).toBeInTheDocument();
    expect(screen.getByText("10.0%")).toBeInTheDocument();
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
});
