import { describe, it, expect, vi } from "vitest";
import { screen, fireEvent } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import TransactionFilters from "./TransactionFilters";
import type { AccountResponse, TransactionFilters as Filters } from "@/api/client";

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
];

const emptyFilters: Filters = {};

describe("TransactionFilters", () => {
  it("renders search input", () => {
    renderWithProviders(
      <TransactionFilters
        filters={emptyFilters}
        accounts={mockAccounts}
        onFiltersChange={vi.fn()}
      />,
    );
    expect(screen.getByPlaceholderText("Search transactions...")).toBeInTheDocument();
  });

  it("renders account filter dropdown", () => {
    renderWithProviders(
      <TransactionFilters
        filters={emptyFilters}
        accounts={mockAccounts}
        onFiltersChange={vi.fn()}
      />,
    );
    expect(screen.getByLabelText("Account")).toBeInTheDocument();
  });

  it("calls onFiltersChange when search input changes", () => {
    const onChange = vi.fn();
    renderWithProviders(
      <TransactionFilters
        filters={emptyFilters}
        accounts={mockAccounts}
        onFiltersChange={onChange}
      />,
    );
    fireEvent.change(screen.getByPlaceholderText("Search transactions..."), {
      target: { value: "coffee" },
    });
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ search: "coffee" }));
  });

  it("calls onFiltersChange when account filter changes", () => {
    const onChange = vi.fn();
    renderWithProviders(
      <TransactionFilters
        filters={emptyFilters}
        accounts={mockAccounts}
        onFiltersChange={onChange}
      />,
    );
    fireEvent.change(screen.getByLabelText("Account"), { target: { value: "acc-1" } });
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ account_id: "acc-1" }));
  });

  it("has accessible search input with aria-label", () => {
    renderWithProviders(
      <TransactionFilters
        filters={emptyFilters}
        accounts={mockAccounts}
        onFiltersChange={vi.fn()}
      />,
    );
    const searchInput = screen.getByPlaceholderText("Search transactions...");
    expect(searchInput).toHaveAttribute("aria-label", "Search transactions");
  });

  it("renders type filter", () => {
    renderWithProviders(
      <TransactionFilters
        filters={emptyFilters}
        accounts={mockAccounts}
        onFiltersChange={vi.fn()}
      />,
    );
    expect(screen.getByLabelText("Type")).toBeInTheDocument();
  });
});
