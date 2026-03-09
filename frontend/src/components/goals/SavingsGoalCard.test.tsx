import { describe, it, expect, vi } from "vitest";
import { screen, fireEvent } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import SavingsGoalCard from "./SavingsGoalCard";
import type { GoalResponse } from "@/api/client";

const savingsGoal: GoalResponse = {
  id: "s1",
  user_id: "u1",
  name: "Engagement ring",
  goal_type: "savings",
  target_amount: 8000,
  current_amount: 2400,
  priority_rank: 1,
  deadline: "2026-09-01",
  required_monthly_contribution: 400,
  on_track: false,
  created_at: "2025-01-01T00:00:00Z",
  updated_at: "2025-06-01T00:00:00Z",
};

describe("SavingsGoalCard", () => {
  it("renders goal name", () => {
    renderWithProviders(
      <SavingsGoalCard goal={savingsGoal} onClick={vi.fn()} onContribute={vi.fn()} />,
    );
    expect(screen.getByText("Engagement ring")).toBeInTheDocument();
  });

  it("renders current and target amounts", () => {
    renderWithProviders(
      <SavingsGoalCard goal={savingsGoal} onClick={vi.fn()} onContribute={vi.fn()} />,
    );
    expect(screen.getByText("$2,400.00")).toBeInTheDocument();
    expect(screen.getByText(/\$8,000\.00/)).toBeInTheDocument();
  });

  it("renders progress bar", () => {
    renderWithProviders(
      <SavingsGoalCard goal={savingsGoal} onClick={vi.fn()} onContribute={vi.fn()} />,
    );
    expect(screen.getByRole("progressbar")).toBeInTheDocument();
  });

  it("renders on-track indicator when behind", () => {
    renderWithProviders(
      <SavingsGoalCard goal={savingsGoal} onClick={vi.fn()} onContribute={vi.fn()} />,
    );
    expect(screen.getByText("Behind")).toBeInTheDocument();
  });

  it("renders on-track indicator when on track", () => {
    const onTrackGoal = { ...savingsGoal, on_track: true };
    renderWithProviders(
      <SavingsGoalCard goal={onTrackGoal} onClick={vi.fn()} onContribute={vi.fn()} />,
    );
    expect(screen.getByText("On Track")).toBeInTheDocument();
  });

  it("renders required monthly contribution", () => {
    renderWithProviders(
      <SavingsGoalCard goal={savingsGoal} onClick={vi.fn()} onContribute={vi.fn()} />,
    );
    expect(screen.getByText("$400.00")).toBeInTheDocument();
  });

  it("renders deadline", () => {
    renderWithProviders(
      <SavingsGoalCard goal={savingsGoal} onClick={vi.fn()} onContribute={vi.fn()} />,
    );
    expect(screen.getByText(/Sep 1, 2026/)).toBeInTheDocument();
  });

  it("calls onClick when clicked", () => {
    const onClick = vi.fn();
    renderWithProviders(
      <SavingsGoalCard goal={savingsGoal} onClick={onClick} onContribute={vi.fn()} />,
    );
    fireEvent.click(screen.getByText("Engagement ring"));
    expect(onClick).toHaveBeenCalled();
  });

  it("calls onContribute when contribute button is clicked", () => {
    const onClick = vi.fn();
    const onContribute = vi.fn();
    renderWithProviders(
      <SavingsGoalCard goal={savingsGoal} onClick={onClick} onContribute={onContribute} />,
    );
    fireEvent.click(screen.getByText("Contribute"));
    expect(onContribute).toHaveBeenCalled();
    expect(onClick).not.toHaveBeenCalled();
  });
});
