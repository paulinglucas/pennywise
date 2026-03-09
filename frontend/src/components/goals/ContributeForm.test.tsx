import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import ContributeForm from "./ContributeForm";
import type { TransactionResponse } from "@/api/client";

const mockTransactions: TransactionResponse[] = [
  {
    id: "txn-001",
    user_id: "u1",
    account_id: "acc-001",
    type: "expense",
    category: "savings",
    amount: 500,
    currency: "USD",
    date: "2026-03-01",
    is_recurring: false,
    tags: [],
    created_at: "2026-03-01T00:00:00Z",
    updated_at: "2026-03-01T00:00:00Z",
  },
];

describe("ContributeForm", () => {
  it("renders amount input and submit button", () => {
    render(<ContributeForm onSubmit={vi.fn()} onCancel={vi.fn()} isSubmitting={false} />);
    expect(screen.getByLabelText("Amount")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Add Contribution" })).toBeInTheDocument();
  });

  it("calls onSubmit with amount when form is submitted", () => {
    const onSubmit = vi.fn();
    render(<ContributeForm onSubmit={onSubmit} onCancel={vi.fn()} isSubmitting={false} />);

    fireEvent.change(screen.getByLabelText("Amount"), { target: { value: "500" } });
    fireEvent.click(screen.getByRole("button", { name: "Add Contribution" }));

    expect(onSubmit).toHaveBeenCalledWith({ amount: 500 });
  });

  it("calls onSubmit with notes when provided", () => {
    const onSubmit = vi.fn();
    render(<ContributeForm onSubmit={onSubmit} onCancel={vi.fn()} isSubmitting={false} />);

    fireEvent.change(screen.getByLabelText("Amount"), { target: { value: "250" } });
    fireEvent.change(screen.getByLabelText("Notes"), { target: { value: "Bonus deposit" } });
    fireEvent.click(screen.getByRole("button", { name: "Add Contribution" }));

    expect(onSubmit).toHaveBeenCalledWith({ amount: 250, notes: "Bonus deposit" });
  });

  it("does not submit with zero amount", () => {
    const onSubmit = vi.fn();
    render(<ContributeForm onSubmit={onSubmit} onCancel={vi.fn()} isSubmitting={false} />);

    fireEvent.click(screen.getByRole("button", { name: "Add Contribution" }));
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("disables submit button when isSubmitting is true", () => {
    render(<ContributeForm onSubmit={vi.fn()} onCancel={vi.fn()} isSubmitting={true} />);
    expect(screen.getByRole("button", { name: "Adding..." })).toBeDisabled();
  });

  it("calls onCancel when cancel button is clicked", () => {
    const onCancel = vi.fn();
    render(<ContributeForm onSubmit={vi.fn()} onCancel={onCancel} isSubmitting={false} />);

    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onCancel).toHaveBeenCalled();
  });

  it("does not render transaction selector when no transactions provided", () => {
    render(<ContributeForm onSubmit={vi.fn()} onCancel={vi.fn()} isSubmitting={false} />);
    expect(screen.queryByLabelText("Link to Transaction")).not.toBeInTheDocument();
  });

  it("renders transaction selector when transactions provided", () => {
    render(
      <ContributeForm
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        isSubmitting={false}
        transactions={mockTransactions}
      />,
    );
    expect(screen.getByLabelText("Link to Transaction")).toBeInTheDocument();
    expect(screen.getByText("None")).toBeInTheDocument();
  });

  it("submits with transaction_id when transaction selected", () => {
    const onSubmit = vi.fn();
    render(
      <ContributeForm
        onSubmit={onSubmit}
        onCancel={vi.fn()}
        isSubmitting={false}
        transactions={mockTransactions}
      />,
    );

    fireEvent.change(screen.getByLabelText("Link to Transaction"), {
      target: { value: "txn-001" },
    });
    fireEvent.change(screen.getByLabelText("Amount"), { target: { value: "500" } });
    fireEvent.click(screen.getByRole("button", { name: "Add Contribution" }));

    expect(onSubmit).toHaveBeenCalledWith({
      amount: 500,
      transaction_id: "txn-001",
    });
  });

  it("auto-fills amount when transaction selected and amount is empty", () => {
    render(
      <ContributeForm
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        isSubmitting={false}
        transactions={mockTransactions}
      />,
    );

    fireEvent.change(screen.getByLabelText("Link to Transaction"), {
      target: { value: "txn-001" },
    });

    expect(screen.getByLabelText("Amount")).toHaveValue(500);
  });
});
