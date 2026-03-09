import { describe, it, expect, vi } from "vitest";
import { screen, fireEvent } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import GoalForm from "./GoalForm";

describe("GoalForm", () => {
  it("renders name and target amount fields", () => {
    renderWithProviders(<GoalForm onSubmit={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByLabelText("Name")).toBeInTheDocument();
    expect(screen.getByLabelText("Target Amount")).toBeInTheDocument();
  });

  it("renders goal type selector", () => {
    renderWithProviders(<GoalForm onSubmit={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByLabelText("Goal Type")).toBeInTheDocument();
  });

  it("calls onSubmit with savings goal data", () => {
    const onSubmit = vi.fn();
    renderWithProviders(<GoalForm onSubmit={onSubmit} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "New Car" } });
    fireEvent.change(screen.getByLabelText("Target Amount"), { target: { value: "25000" } });
    fireEvent.submit(screen.getByRole("form"));
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "New Car",
        target_amount: 25000,
        goal_type: "savings",
      }),
    );
  });

  it("calls onSubmit with debt goal data", () => {
    const onSubmit = vi.fn();
    renderWithProviders(<GoalForm onSubmit={onSubmit} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByLabelText("Goal Type"), { target: { value: "debt_payoff" } });
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Credit Card" } });
    fireEvent.change(screen.getByLabelText("Target Amount"), { target: { value: "5000" } });
    fireEvent.change(screen.getByLabelText("Current Amount"), { target: { value: "3000" } });
    fireEvent.submit(screen.getByRole("form"));
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "Credit Card",
        goal_type: "debt_payoff",
        target_amount: 5000,
        current_amount: 3000,
      }),
    );
  });

  it("calls onCancel when cancel button clicked", () => {
    const onCancel = vi.fn();
    renderWithProviders(<GoalForm onSubmit={vi.fn()} onCancel={onCancel} />);
    fireEvent.click(screen.getByText("Cancel"));
    expect(onCancel).toHaveBeenCalled();
  });

  it("pre-fills fields when editing", () => {
    renderWithProviders(
      <GoalForm
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        initialValues={{
          name: "Engagement ring",
          goal_type: "savings",
          target_amount: 8000,
          current_amount: 2400,
          deadline: "2026-09-01",
        }}
      />,
    );
    expect(screen.getByDisplayValue("Engagement ring")).toBeInTheDocument();
    expect(screen.getByDisplayValue("8000")).toBeInTheDocument();
    expect(screen.getByDisplayValue("2400")).toBeInTheDocument();
  });
});
