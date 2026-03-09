import { describe, it, expect, vi } from "vitest";
import { screen, fireEvent, waitFor } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import TransactionForm from "./TransactionForm";
import type { AccountResponse, TransactionResponse } from "@/api/client";

vi.mock("@/hooks/useCategories", () => ({
  useCategories: () => ({
    data: ["food", "rent", "salary"],
    isLoading: false,
  }),
}));

const mockAccounts: AccountResponse[] = [
  {
    id: "acc-1",
    user_id: "user-1",
    name: "Checking",
    institution: "Chase",
    account_type: "checking",
    currency: "USD",
    is_active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
  {
    id: "acc-2",
    user_id: "user-1",
    name: "Savings",
    institution: "Marcus",
    account_type: "savings",
    currency: "USD",
    is_active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

describe("TransactionForm", () => {
  it("renders add form with default values", () => {
    renderWithProviders(
      <TransactionForm accounts={mockAccounts} onSubmit={vi.fn()} onCancel={vi.fn()} />,
    );
    expect(screen.getByLabelText("Amount")).toBeInTheDocument();
    expect(screen.getByLabelText("Category")).toBeInTheDocument();
    expect(screen.getByLabelText("Date")).toBeInTheDocument();
    expect(screen.getByLabelText("Account")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Add Transaction" })).toBeInTheDocument();
  });

  it("renders edit form when transaction is provided", () => {
    const transaction: TransactionResponse = {
      id: "txn-1",
      user_id: "user-1",
      account_id: "acc-1",
      type: "expense",
      category: "food",
      amount: 42.5,
      currency: "USD",
      date: "2026-03-01",
      notes: "Lunch",
      is_recurring: false,
      tags: ["dining"],
      created_at: "2026-03-01T00:00:00Z",
      updated_at: "2026-03-01T00:00:00Z",
    };
    renderWithProviders(
      <TransactionForm
        accounts={mockAccounts}
        transaction={transaction}
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
      />,
    );
    expect(screen.getByDisplayValue("42.50")).toBeInTheDocument();
    expect(screen.getByDisplayValue("food")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Lunch")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save Changes" })).toBeInTheDocument();
  });

  it("calls onSubmit with form data", async () => {
    const onSubmit = vi.fn();
    renderWithProviders(
      <TransactionForm accounts={mockAccounts} onSubmit={onSubmit} onCancel={vi.fn()} />,
    );

    fireEvent.change(screen.getByLabelText("Amount"), { target: { value: "25.00" } });
    fireEvent.change(screen.getByLabelText("Category"), { target: { value: "food" } });
    fireEvent.click(screen.getByRole("button", { name: "Add Transaction" }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledOnce();
    });
    const callArg = onSubmit.mock.calls[0]?.[0];
    expect(callArg.amount).toBe(25);
    expect(callArg.category).toBe("food");
    expect(callArg.type).toBe("expense");
  });

  it("calls onCancel when cancel is clicked", () => {
    const onCancel = vi.fn();
    renderWithProviders(
      <TransactionForm accounts={mockAccounts} onSubmit={vi.fn()} onCancel={onCancel} />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it("toggles between expense and deposit", () => {
    renderWithProviders(
      <TransactionForm accounts={mockAccounts} onSubmit={vi.fn()} onCancel={vi.fn()} />,
    );
    const depositBtn = screen.getByRole("button", { name: "Deposit" });
    fireEvent.click(depositBtn);
    expect(depositBtn).toHaveStyle({ color: "var(--color-accent)" });
  });
});
