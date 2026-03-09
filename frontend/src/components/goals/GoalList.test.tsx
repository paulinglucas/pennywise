import { describe, it, expect, vi } from "vitest";
import { screen, fireEvent } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import GoalList from "./GoalList";
import type { GoalResponse } from "@/api/client";

function makeGoal(overrides: Partial<GoalResponse>): GoalResponse {
  return {
    id: "g1",
    user_id: "u1",
    name: "Test Goal",
    goal_type: "savings",
    target_amount: 10000,
    current_amount: 2500,
    priority_rank: 1,
    created_at: "2025-01-01T00:00:00Z",
    updated_at: "2025-01-01T00:00:00Z",
    ...overrides,
  };
}

const debtGoal = makeGoal({
  id: "d1",
  name: "Credit Card",
  goal_type: "debt_payoff",
  target_amount: 5000,
  current_amount: 3000,
  priority_rank: 1,
});

const debtGoal2 = makeGoal({
  id: "d2",
  name: "Student Loan",
  goal_type: "debt_payoff",
  target_amount: 20000,
  current_amount: 15000,
  priority_rank: 2,
});

const savingsGoal = makeGoal({
  id: "s1",
  name: "Emergency Fund",
  goal_type: "savings",
  target_amount: 10000,
  current_amount: 4000,
  priority_rank: 1,
  on_track: true,
});

const savingsGoal2 = makeGoal({
  id: "s2",
  name: "Vacation",
  goal_type: "savings",
  target_amount: 3000,
  current_amount: 500,
  priority_rank: 2,
  on_track: false,
});

describe("GoalList", () => {
  it("renders debt payoff section with debt goals", () => {
    renderWithProviders(
      <GoalList
        goals={[debtGoal, debtGoal2]}
        onGoalClick={vi.fn()}
        onContribute={vi.fn()}
        onReorder={vi.fn()}
      />,
    );
    expect(screen.getByText("Debt Payoff")).toBeInTheDocument();
    expect(screen.getByText("Credit Card")).toBeInTheDocument();
    expect(screen.getByText("Student Loan")).toBeInTheDocument();
  });

  it("renders savings section with savings goals", () => {
    renderWithProviders(
      <GoalList
        goals={[savingsGoal, savingsGoal2]}
        onGoalClick={vi.fn()}
        onContribute={vi.fn()}
        onReorder={vi.fn()}
      />,
    );
    expect(screen.getByText("Savings Goals")).toBeInTheDocument();
    expect(screen.getByText("Emergency Fund")).toBeInTheDocument();
    expect(screen.getByText("Vacation")).toBeInTheDocument();
  });

  it("renders both sections when mixed goals exist", () => {
    renderWithProviders(
      <GoalList
        goals={[debtGoal, savingsGoal]}
        onGoalClick={vi.fn()}
        onContribute={vi.fn()}
        onReorder={vi.fn()}
      />,
    );
    expect(screen.getByText("Debt Payoff")).toBeInTheDocument();
    expect(screen.getByText("Savings Goals")).toBeInTheDocument();
  });

  it("hides debt section when no debt goals exist", () => {
    renderWithProviders(
      <GoalList
        goals={[savingsGoal]}
        onGoalClick={vi.fn()}
        onContribute={vi.fn()}
        onReorder={vi.fn()}
      />,
    );
    expect(screen.queryByText("Debt Payoff")).not.toBeInTheDocument();
    expect(screen.getByText("Savings Goals")).toBeInTheDocument();
  });

  it("hides savings section when no savings goals exist", () => {
    renderWithProviders(
      <GoalList
        goals={[debtGoal]}
        onGoalClick={vi.fn()}
        onContribute={vi.fn()}
        onReorder={vi.fn()}
      />,
    );
    expect(screen.getByText("Debt Payoff")).toBeInTheDocument();
    expect(screen.queryByText("Savings Goals")).not.toBeInTheDocument();
  });

  it("sorts goals by priority_rank within each section", () => {
    const reversed = [debtGoal2, debtGoal];
    renderWithProviders(
      <GoalList
        goals={reversed}
        onGoalClick={vi.fn()}
        onContribute={vi.fn()}
        onReorder={vi.fn()}
      />,
    );
    const names = screen.getAllByText(/Credit Card|Student Loan/);
    expect(names[0]).toHaveTextContent("Credit Card");
    expect(names[1]).toHaveTextContent("Student Loan");
  });

  it("calls onGoalClick when a goal card is clicked", () => {
    const onGoalClick = vi.fn();
    renderWithProviders(
      <GoalList
        goals={[savingsGoal]}
        onGoalClick={onGoalClick}
        onContribute={vi.fn()}
        onReorder={vi.fn()}
      />,
    );
    const card = screen.getByText("Emergency Fund").closest("[role='button']");
    fireEvent.click(card!);
    expect(onGoalClick).toHaveBeenCalledWith(savingsGoal);
  });
});
