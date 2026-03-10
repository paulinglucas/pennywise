import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import CashFlowSummary from "./CashFlowSummary";

const defaultBreakdown = { assets: 100000, cash: 50000, debt: 0 };

describe("CashFlowSummary", () => {
  it("renders net worth and cash flow", () => {
    render(<CashFlowSummary netWorth={150000} breakdown={defaultBreakdown} cashFlow={2500} />);

    expect(screen.getByText("Net Worth")).toBeInTheDocument();
    expect(screen.getByText("$150,000.00")).toBeInTheDocument();
    expect(screen.getByText("Cash Flow This Month")).toBeInTheDocument();
    expect(screen.getByText("$2,500.00")).toBeInTheDocument();
  });

  it("applies positive color to positive cash flow", () => {
    render(<CashFlowSummary netWorth={100000} breakdown={defaultBreakdown} cashFlow={1000} />);

    const cashFlowValue = screen.getByText("$1,000.00");
    expect(cashFlowValue.style.color).toBe("var(--color-positive)");
  });

  it("applies negative color to negative cash flow", () => {
    render(<CashFlowSummary netWorth={100000} breakdown={defaultBreakdown} cashFlow={-500} />);

    const cashFlowValue = screen.getByText("-$500.00");
    expect(cashFlowValue.style.color).toBe("var(--color-negative)");
  });

  it("renders cash flow prefix for positive values", () => {
    render(<CashFlowSummary netWorth={100000} breakdown={defaultBreakdown} cashFlow={1000} />);

    expect(screen.getByText("+")).toBeInTheDocument();
  });

  it("renders net worth breakdown with assets, cash, and debt", () => {
    render(
      <CashFlowSummary
        netWorth={140000}
        breakdown={{ assets: 400000, cash: 50000, debt: 310000 }}
        cashFlow={0}
      />,
    );

    expect(screen.getByText("Assets")).toBeInTheDocument();
    expect(screen.getByText("$400,000.00")).toBeInTheDocument();
    expect(screen.getByText("Cash")).toBeInTheDocument();
    expect(screen.getByText("$50,000.00")).toBeInTheDocument();
    expect(screen.getByText("Debt")).toBeInTheDocument();
    expect(screen.getByText("-$310,000.00")).toBeInTheDocument();
  });
});
