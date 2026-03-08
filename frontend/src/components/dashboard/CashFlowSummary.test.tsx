import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import CashFlowSummary from "./CashFlowSummary";

describe("CashFlowSummary", () => {
  it("renders net worth and cash flow", () => {
    render(<CashFlowSummary netWorth={150000} cashFlow={2500} />);

    expect(screen.getByText("Net Worth")).toBeInTheDocument();
    expect(screen.getByText("$150,000.00")).toBeInTheDocument();
    expect(screen.getByText("Cash Flow This Month")).toBeInTheDocument();
    expect(screen.getByText("$2,500.00")).toBeInTheDocument();
  });

  it("applies positive color to positive cash flow", () => {
    render(<CashFlowSummary netWorth={100000} cashFlow={1000} />);

    const cashFlowValue = screen.getByText("$1,000.00");
    expect(cashFlowValue.style.color).toBe("var(--color-positive)");
  });

  it("applies negative color to negative cash flow", () => {
    render(<CashFlowSummary netWorth={100000} cashFlow={-500} />);

    const cashFlowValue = screen.getByText("-$500.00");
    expect(cashFlowValue.style.color).toBe("var(--color-negative)");
  });

  it("renders cash flow prefix for positive values", () => {
    render(<CashFlowSummary netWorth={100000} cashFlow={1000} />);

    expect(screen.getByText("+")).toBeInTheDocument();
  });
});
