import { describe, it, expect, vi } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "@/test-utils";
import TransactionGroupForm from "./TransactionGroupForm";
import type { AccountResponse } from "@/api/client";

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
    account_type: "checking",
    currency: "USD",
    institution: "Bank",
    is_active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
  {
    id: "acc-2",
    user_id: "user-1",
    name: "401k",
    account_type: "retirement_401k",
    currency: "USD",
    institution: "Fidelity",
    is_active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

describe("TransactionGroupForm", () => {
  it("renders group name input and member rows", () => {
    renderWithProviders(
      <TransactionGroupForm
        accounts={mockAccounts}
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        isSubmitting={false}
      />,
    );
    expect(screen.getByLabelText("Group Name")).toBeInTheDocument();
    expect(screen.getAllByTestId("member-row").length).toBeGreaterThanOrEqual(2);
  });

  it("starts with two empty member rows", () => {
    renderWithProviders(
      <TransactionGroupForm
        accounts={mockAccounts}
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        isSubmitting={false}
      />,
    );
    const rows = screen.getAllByTestId("member-row");
    expect(rows).toHaveLength(2);
  });

  it("adds a new member row when clicking add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <TransactionGroupForm
        accounts={mockAccounts}
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        isSubmitting={false}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Add Split" }));

    await waitFor(() => {
      expect(screen.getAllByTestId("member-row")).toHaveLength(3);
    });
  });

  it("disables submit when fewer than 2 members have data", () => {
    renderWithProviders(
      <TransactionGroupForm
        accounts={mockAccounts}
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        isSubmitting={false}
      />,
    );

    const submitButton = screen.getByRole("button", { name: "Create Group" });
    expect(submitButton).toBeDisabled();
  });

  it("shows total that auto-sums member amounts", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <TransactionGroupForm
        accounts={mockAccounts}
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        isSubmitting={false}
      />,
    );

    const amountInputs = screen.getAllByPlaceholderText("Amount");
    await user.type(amountInputs[0]!, "4000");
    await user.type(amountInputs[1]!, "500");

    await waitFor(() => {
      expect(screen.getByTestId("group-total")).toHaveTextContent("$4,500.00");
    });
  });
});
