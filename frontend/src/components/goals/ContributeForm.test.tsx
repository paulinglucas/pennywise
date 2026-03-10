import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import ContributeForm from "./ContributeForm";
import type { TransactionResponse, AccountResponse } from "@/api/client";

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

const mockAccounts: AccountResponse[] = [
  {
    id: "acc-001",
    user_id: "u1",
    name: "Checking",
    institution: "Huntington Bank",
    account_type: "checking",
    currency: "USD",
    is_active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

const mockCategories = ["Savings", "Debt Payment", "Groceries"];

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

  it("does not render transaction toggle when no transactions or accounts provided", () => {
    render(<ContributeForm onSubmit={vi.fn()} onCancel={vi.fn()} isSubmitting={false} />);
    expect(screen.queryByText("Transaction")).not.toBeInTheDocument();
    expect(screen.queryByText("Link Existing")).not.toBeInTheDocument();
    expect(screen.queryByText("Create New")).not.toBeInTheDocument();
  });

  it("shows Link Existing button when transactions provided", () => {
    render(
      <ContributeForm
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        isSubmitting={false}
        transactions={mockTransactions}
      />,
    );
    expect(screen.getByText("None")).toBeInTheDocument();
    expect(screen.getByText("Link Existing")).toBeInTheDocument();
  });

  it("shows Create New button when accounts provided", () => {
    render(
      <ContributeForm
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        isSubmitting={false}
        accounts={mockAccounts}
      />,
    );
    expect(screen.getByText("None")).toBeInTheDocument();
    expect(screen.getByText("Create New")).toBeInTheDocument();
  });

  it("shows transaction selector when Link Existing is clicked", () => {
    render(
      <ContributeForm
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        isSubmitting={false}
        transactions={mockTransactions}
      />,
    );

    fireEvent.click(screen.getByText("Link Existing"));
    expect(screen.getByLabelText("Select Transaction")).toBeInTheDocument();
  });

  it("submits with transaction_id when transaction selected via link mode", () => {
    const onSubmit = vi.fn();
    render(
      <ContributeForm
        onSubmit={onSubmit}
        onCancel={vi.fn()}
        isSubmitting={false}
        transactions={mockTransactions}
      />,
    );

    fireEvent.click(screen.getByText("Link Existing"));
    fireEvent.change(screen.getByLabelText("Select Transaction"), {
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

    fireEvent.click(screen.getByText("Link Existing"));
    fireEvent.change(screen.getByLabelText("Select Transaction"), {
      target: { value: "txn-001" },
    });

    expect(screen.getByLabelText("Amount")).toHaveValue(500);
  });

  it("shows account, category, and date fields when Create New is clicked", () => {
    render(
      <ContributeForm
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        isSubmitting={false}
        accounts={mockAccounts}
        categories={mockCategories}
      />,
    );

    fireEvent.click(screen.getByText("Create New"));
    expect(screen.getByLabelText("Account")).toBeInTheDocument();
    expect(screen.getByLabelText("Category")).toBeInTheDocument();
    expect(screen.getByLabelText("Date")).toBeInTheDocument();
  });

  it("submits with create_transaction data when in create mode", () => {
    const onSubmit = vi.fn();
    render(
      <ContributeForm
        onSubmit={onSubmit}
        onCancel={vi.fn()}
        isSubmitting={false}
        accounts={mockAccounts}
        categories={mockCategories}
      />,
    );

    fireEvent.click(screen.getByText("Create New"));
    fireEvent.change(screen.getByLabelText("Account"), { target: { value: "acc-001" } });
    fireEvent.change(screen.getByLabelText("Category"), { target: { value: "Savings" } });
    fireEvent.change(screen.getByLabelText("Date"), { target: { value: "2026-03-10" } });
    fireEvent.change(screen.getByLabelText("Amount"), { target: { value: "100" } });
    fireEvent.click(screen.getByRole("button", { name: "Add Contribution" }));

    expect(onSubmit).toHaveBeenCalledWith({
      amount: 100,
      create_transaction: {
        account_id: "acc-001",
        category: "Savings",
        date: "2026-03-10",
      },
    });
  });

  it("defaults category to Debt Payment for debt_payoff goals", () => {
    render(
      <ContributeForm
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        isSubmitting={false}
        accounts={mockAccounts}
        categories={mockCategories}
        goalType="debt_payoff"
      />,
    );

    fireEvent.click(screen.getByText("Create New"));
    expect(screen.getByLabelText("Category")).toHaveValue("Debt Payment");
  });
});
