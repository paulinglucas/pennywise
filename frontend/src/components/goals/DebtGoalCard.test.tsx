import { describe, it, expect, vi } from "vitest";
import { screen, fireEvent } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import DebtGoalCard from "./DebtGoalCard";
import type { GoalResponse } from "@/api/client";

const debtGoal: GoalResponse = {
  id: "d1",
  user_id: "u1",
  name: "Pay off car loan",
  goal_type: "debt_payoff",
  target_amount: 7500,
  current_amount: 4200,
  priority_rank: 1,
  deadline: "2027-06-01",
  required_monthly_contribution: 280,
  projected_completion_date: "2027-03-15",
  on_track: true,
  created_at: "2025-01-01T00:00:00Z",
  updated_at: "2025-06-01T00:00:00Z",
};

describe("DebtGoalCard", () => {
  it("renders goal name", () => {
    renderWithProviders(<DebtGoalCard goal={debtGoal} onClick={vi.fn()} onContribute={vi.fn()} />);
    expect(screen.getByText("Pay off car loan")).toBeInTheDocument();
  });

  it("renders remaining balance", () => {
    renderWithProviders(<DebtGoalCard goal={debtGoal} onClick={vi.fn()} onContribute={vi.fn()} />);
    expect(screen.getByText("$4,200.00")).toBeInTheDocument();
  });

  it("renders progress bar", () => {
    renderWithProviders(<DebtGoalCard goal={debtGoal} onClick={vi.fn()} onContribute={vi.fn()} />);
    expect(screen.getByRole("progressbar")).toBeInTheDocument();
  });

  it("renders estimated payoff date when deadline exists", () => {
    renderWithProviders(<DebtGoalCard goal={debtGoal} onClick={vi.fn()} onContribute={vi.fn()} />);
    expect(screen.getByText(/Jun 1, 2027/)).toBeInTheDocument();
  });

  it("renders monthly payment when computed", () => {
    renderWithProviders(<DebtGoalCard goal={debtGoal} onClick={vi.fn()} onContribute={vi.fn()} />);
    expect(screen.getByText("$280.00")).toBeInTheDocument();
  });

  it("calls onClick when clicked", () => {
    const onClick = vi.fn();
    renderWithProviders(<DebtGoalCard goal={debtGoal} onClick={onClick} onContribute={vi.fn()} />);
    fireEvent.click(screen.getByText("Pay off car loan"));
    expect(onClick).toHaveBeenCalled();
  });

  it("calls onContribute when make payment button is clicked", () => {
    const onClick = vi.fn();
    const onContribute = vi.fn();
    renderWithProviders(
      <DebtGoalCard goal={debtGoal} onClick={onClick} onContribute={onContribute} />,
    );
    fireEvent.click(screen.getByText("Make Payment"));
    expect(onContribute).toHaveBeenCalled();
    expect(onClick).not.toHaveBeenCalled();
  });
});
