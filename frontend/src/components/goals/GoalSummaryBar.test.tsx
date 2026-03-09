import { describe, it, expect } from "vitest";
import { screen } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import GoalSummaryBar from "./GoalSummaryBar";
import type { GoalResponse } from "@/api/client";

function makeGoal(overrides: Partial<GoalResponse>): GoalResponse {
  return {
    id: "g1",
    user_id: "u1",
    name: "Test Goal",
    goal_type: "savings",
    target_amount: 10000,
    current_amount: 5000,
    priority_rank: 1,
    created_at: "2025-01-01T00:00:00Z",
    updated_at: "2025-06-01T00:00:00Z",
    ...overrides,
  };
}

describe("GoalSummaryBar", () => {
  it("renders total debt remaining", () => {
    const goals = [
      makeGoal({ id: "d1", goal_type: "debt_payoff", target_amount: 20000, current_amount: 12000 }),
      makeGoal({ id: "d2", goal_type: "debt_payoff", target_amount: 8000, current_amount: 3000 }),
    ];
    renderWithProviders(<GoalSummaryBar goals={goals} />);
    expect(screen.getByText("$15,000.00")).toBeInTheDocument();
  });

  it("renders total savings progress", () => {
    const goals = [
      makeGoal({ id: "s1", goal_type: "savings", target_amount: 10000, current_amount: 4000 }),
      makeGoal({ id: "s2", goal_type: "savings", target_amount: 5000, current_amount: 2000 }),
    ];
    renderWithProviders(<GoalSummaryBar goals={goals} />);
    expect(screen.getByText("$6,000.00")).toBeInTheDocument();
  });

  it("renders section labels", () => {
    const goals = [
      makeGoal({ id: "d1", goal_type: "debt_payoff", target_amount: 5000, current_amount: 3000 }),
      makeGoal({ id: "s1", goal_type: "savings", target_amount: 10000, current_amount: 2000 }),
    ];
    renderWithProviders(<GoalSummaryBar goals={goals} />);
    expect(screen.getByText("Debt Remaining")).toBeInTheDocument();
    expect(screen.getByText("Savings Progress")).toBeInTheDocument();
  });

  it("handles empty goals list", () => {
    renderWithProviders(<GoalSummaryBar goals={[]} />);
    const zeros = screen.getAllByText("$0.00");
    expect(zeros.length).toBe(2);
  });
});
