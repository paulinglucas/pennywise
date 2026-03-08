import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import InsightCards, { generateInsights } from "./InsightCard";

describe("generateInsights", () => {
  it("generates spending insight when categories exist", () => {
    const insights = generateInsights({
      spending: [
        { category: "Food", amount: 1200, percentage: 60 },
        { category: "Housing", amount: 800, percentage: 40 },
      ],
      debts: [],
      cashFlow: 1000,
    });

    const spendingInsight = insights.find((insight) => insight.includes("Food"));
    expect(spendingInsight).toBeDefined();
  });

  it("generates positive cash flow insight", () => {
    const insights = generateInsights({
      spending: [],
      debts: [],
      cashFlow: 2500,
    });

    const cashFlowInsight = insights.find((insight) => insight.includes("positive"));
    expect(cashFlowInsight).toBeDefined();
  });

  it("generates negative cash flow warning", () => {
    const insights = generateInsights({
      spending: [],
      debts: [],
      cashFlow: -500,
    });

    const warning = insights.find((insight) => insight.includes("exceeding"));
    expect(warning).toBeDefined();
  });

  it("generates debt payoff insight when months remaining is low", () => {
    const insights = generateInsights({
      spending: [],
      debts: [{ name: "Car Loan", months_remaining: 10, monthly_payment: 400 }],
      cashFlow: 0,
    });

    const debtInsight = insights.find((insight) => insight.includes("Car Loan"));
    expect(debtInsight).toBeDefined();
  });
});

describe("InsightCards", () => {
  it("renders insight messages", () => {
    render(
      <InsightCards
        spending={[{ category: "Food", amount: 500, percentage: 100 }]}
        debts={[]}
        cashFlow={2000}
      />,
    );

    expect(screen.getByText("Insights")).toBeInTheDocument();
  });

  it("renders empty state when no insights can be generated", () => {
    render(<InsightCards spending={[]} debts={[]} cashFlow={0} />);

    expect(screen.getByText("Insights")).toBeInTheDocument();
  });
});
