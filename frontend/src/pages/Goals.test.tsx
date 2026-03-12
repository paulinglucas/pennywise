import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "@/test-utils";
import Goals from "./Goals";

const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

const emptyGoalList = {
  data: [],
  pagination: { page: 1, per_page: 100, total: 0, total_pages: 0 },
};

const goalList = {
  data: [
    {
      id: "goal-1",
      user_id: "u1",
      name: "Emergency Fund",
      goal_type: "savings",
      target_amount: 10000,
      current_amount: 5000,
      priority_rank: 1,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
    {
      id: "goal-2",
      user_id: "u1",
      name: "Car Loan",
      goal_type: "debt_payoff",
      target_amount: 15000,
      current_amount: 8000,
      priority_rank: 2,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
  ],
  pagination: { page: 1, per_page: 100, total: 2, total_pages: 1 },
};

const mockTransactions = {
  data: [],
  pagination: { page: 1, per_page: 50, total: 0, total_pages: 0 },
};

const mockAccounts = {
  data: [
    { id: "acc-1", name: "Checking", account_type: "checking", currency: "USD", is_active: true },
  ],
  pagination: { page: 1, per_page: 100, total: 1, total_pages: 1 },
};

const mockCategories = ["food", "housing", "transportation"];

function setupFetchForGoals(goalsResponse: unknown) {
  mockFetch.mockImplementation((input: string | URL | Request) => {
    const url =
      typeof input === "string"
        ? input
        : input instanceof URL
          ? input.toString()
          : (input as Request).url;

    if (url.includes("/api/v1/goals")) {
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve(goalsResponse),
      });
    }
    if (url.includes("/api/v1/transactions")) {
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockTransactions),
      });
    }
    if (url.includes("/api/v1/accounts")) {
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockAccounts),
      });
    }
    if (url.includes("/api/v1/categories")) {
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockCategories),
      });
    }
    return Promise.resolve({
      ok: true,
      status: 200,
      json: () => Promise.resolve({}),
    });
  });
}

beforeEach(() => {
  mockFetch.mockReset();
});

describe("Goals", () => {
  it("shows loading skeleton initially", () => {
    mockFetch.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<Goals />);
    expect(document.querySelector(".animate-pulse")).toBeInTheDocument();
  });

  it("shows empty state when no goals exist", async () => {
    setupFetchForGoals(emptyGoalList);
    renderWithProviders(<Goals />);

    await waitFor(() => {
      expect(screen.getByText("No goals yet")).toBeInTheDocument();
    });
    expect(screen.getByText("Add Goal")).toBeInTheDocument();
  });

  it("renders goal list with summary", async () => {
    setupFetchForGoals(goalList);
    renderWithProviders(<Goals />);

    await waitFor(() => {
      expect(screen.getByText("Emergency Fund")).toBeInTheDocument();
    });
    expect(screen.getByText("Car Loan")).toBeInTheDocument();
  });

  it("opens create modal on Add Goal click", async () => {
    setupFetchForGoals(emptyGoalList);
    const user = userEvent.setup();
    renderWithProviders(<Goals />);

    await waitFor(() => {
      expect(screen.getByText("Add Goal")).toBeInTheDocument();
    });

    await user.click(screen.getByText("Add Goal"));
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });

  it("opens edit modal when clicking a goal", async () => {
    setupFetchForGoals(goalList);
    const user = userEvent.setup();
    renderWithProviders(<Goals />);

    await waitFor(() => {
      expect(screen.getByText("Emergency Fund")).toBeInTheDocument();
    });

    await user.click(screen.getByText("Emergency Fund"));

    await waitFor(() => {
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });
    expect(screen.getByText("Edit Goal")).toBeInTheDocument();
    expect(screen.getByText("Delete Goal")).toBeInTheDocument();
  });

  it("shows error state on fetch failure", async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
      json: () =>
        Promise.resolve({
          error: { code: "INTERNAL_ERROR", message: "Server error" },
        }),
      headers: new Headers({ "x-request-id": "req-123" }),
    });

    renderWithProviders(<Goals />);

    await waitFor(() => {
      expect(screen.getByText("Could not load your goals. Please try again.")).toBeInTheDocument();
    });
    expect(screen.getByText("Retry")).toBeInTheDocument();
  });

  it("submits create goal form", async () => {
    setupFetchForGoals(emptyGoalList);
    const user = userEvent.setup();
    renderWithProviders(<Goals />);

    await waitFor(() => {
      expect(screen.getByText("Add Goal")).toBeInTheDocument();
    });

    await user.click(screen.getByText("Add Goal"));

    await waitFor(() => {
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });

    await user.type(screen.getByLabelText("Name"), "New Goal");
    await user.type(screen.getByLabelText("Target Amount"), "5000");

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 201,
      json: () =>
        Promise.resolve({
          id: "goal-new",
          name: "New Goal",
          goal_type: "savings",
          target_amount: 5000,
          current_amount: 0,
          priority_rank: 1,
        }),
    });

    const addButtons = screen.getAllByRole("button", { name: "Add Goal" });
    const submitBtn = addButtons[addButtons.length - 1] as HTMLElement;
    await user.click(submitBtn);

    await waitFor(() => {
      const calls = mockFetch.mock.calls;
      const postCalls = calls.filter((c: unknown[]) => {
        const opts = c[1] as RequestInit | undefined;
        return opts?.method === "POST";
      });
      expect(postCalls.length).toBeGreaterThan(0);
    });
  });

  it("submits update from edit modal", async () => {
    setupFetchForGoals(goalList);
    const user = userEvent.setup();
    renderWithProviders(<Goals />);

    await waitFor(() => {
      expect(screen.getByText("Emergency Fund")).toBeInTheDocument();
    });

    await user.click(screen.getByText("Emergency Fund"));

    await waitFor(() => {
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          id: "goal-1",
          name: "Emergency Fund",
          goal_type: "savings",
          target_amount: 10000,
          current_amount: 5000,
          priority_rank: 1,
        }),
    });

    await user.click(screen.getByRole("button", { name: "Save Changes" }));

    await waitFor(() => {
      const calls = mockFetch.mock.calls;
      const putCalls = calls.filter((c: unknown[]) => {
        const opts = c[1] as RequestInit | undefined;
        return opts?.method === "PUT";
      });
      expect(putCalls.length).toBeGreaterThan(0);
    });
  });

  it("shows delete button in edit modal and closes on delete", async () => {
    setupFetchForGoals(goalList);
    const user = userEvent.setup();
    renderWithProviders(<Goals />);

    await waitFor(() => {
      expect(screen.getByText("Emergency Fund")).toBeInTheDocument();
    });

    await user.click(screen.getByText("Emergency Fund"));

    await waitFor(() => {
      expect(screen.getByText("Delete Goal")).toBeInTheDocument();
    });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.resolve(undefined),
    });

    await user.click(screen.getByText("Delete Goal"));

    await waitFor(() => {
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });
});
