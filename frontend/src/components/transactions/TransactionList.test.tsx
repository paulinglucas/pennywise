import { describe, it, expect, vi } from "vitest";
import { screen } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import TransactionList from "./TransactionList";
import type { TransactionResponse, PaginationMeta } from "@/api/client";

const mockTransactions: TransactionResponse[] = [
  {
    id: "txn-1",
    user_id: "user-1",
    account_id: "acc-1",
    type: "expense",
    category: "food",
    amount: 42.5,
    currency: "USD",
    date: "2026-03-08",
    is_recurring: false,
    tags: ["dining"],
    created_at: "2026-03-08T12:00:00Z",
    updated_at: "2026-03-08T12:00:00Z",
  },
  {
    id: "txn-2",
    user_id: "user-1",
    account_id: "acc-1",
    type: "deposit",
    category: "salary",
    amount: 5000,
    currency: "USD",
    date: "2026-03-08",
    is_recurring: true,
    tags: [],
    created_at: "2026-03-08T12:00:00Z",
    updated_at: "2026-03-08T12:00:00Z",
  },
  {
    id: "txn-3",
    user_id: "user-1",
    account_id: "acc-1",
    type: "expense",
    category: "transport",
    amount: 15,
    currency: "USD",
    date: "2026-03-07",
    is_recurring: false,
    tags: [],
    created_at: "2026-03-07T12:00:00Z",
    updated_at: "2026-03-07T12:00:00Z",
  },
];

const mockPagination: PaginationMeta = {
  page: 1,
  per_page: 25,
  total: 3,
  total_pages: 1,
};

describe("TransactionList", () => {
  it("renders transactions grouped by date", () => {
    renderWithProviders(
      <TransactionList
        transactions={mockTransactions}
        pagination={mockPagination}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onPageChange={vi.fn()}
      />,
    );
    expect(screen.getByText("Mar 8, 2026")).toBeInTheDocument();
    expect(screen.getByText("Mar 7, 2026")).toBeInTheDocument();
  });

  it("renders transaction amounts with correct formatting", () => {
    renderWithProviders(
      <TransactionList
        transactions={mockTransactions}
        pagination={mockPagination}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onPageChange={vi.fn()}
      />,
    );
    expect(screen.getByText("$42.50")).toBeInTheDocument();
    expect(screen.getByText("+$5,000.00")).toBeInTheDocument();
  });

  it("renders category labels", () => {
    renderWithProviders(
      <TransactionList
        transactions={mockTransactions}
        pagination={mockPagination}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onPageChange={vi.fn()}
      />,
    );
    expect(screen.getByText("food")).toBeInTheDocument();
    expect(screen.getByText("salary")).toBeInTheDocument();
    expect(screen.getByText("transport")).toBeInTheDocument();
  });

  it("renders tag chips", () => {
    renderWithProviders(
      <TransactionList
        transactions={mockTransactions}
        pagination={mockPagination}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onPageChange={vi.fn()}
      />,
    );
    expect(screen.getByText("dining")).toBeInTheDocument();
  });

  it("shows daily totals per date group", () => {
    renderWithProviders(
      <TransactionList
        transactions={mockTransactions}
        pagination={mockPagination}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onPageChange={vi.fn()}
      />,
    );
    expect(screen.getByText("+$4,957.50")).toBeInTheDocument();
    expect(screen.getByText("-$15.00")).toBeInTheDocument();
  });
});
