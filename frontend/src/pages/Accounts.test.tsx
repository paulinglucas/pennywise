import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "@/test-utils";
import Accounts from "./Accounts";

const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

const emptyAccountList = {
  data: [],
  pagination: { page: 1, per_page: 100, total: 0, total_pages: 0 },
};

const accountList = {
  data: [
    {
      id: "acc-1",
      user_id: "u1",
      name: "Chase Checking",
      institution: "Chase",
      account_type: "checking",
      currency: "USD",
      is_active: true,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
    {
      id: "acc-2",
      user_id: "u1",
      name: "Fidelity 401k",
      institution: "Fidelity",
      account_type: "retirement_401k",
      currency: "USD",
      is_active: true,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
  ],
  pagination: { page: 1, per_page: 100, total: 2, total_pages: 1 },
};

beforeEach(() => {
  mockFetch.mockReset();
});

describe("Accounts", () => {
  it("shows loading skeleton initially", () => {
    mockFetch.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<Accounts />);
    expect(document.querySelector(".animate-pulse")).toBeInTheDocument();
  });

  it("shows empty state when no accounts", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(emptyAccountList),
    });

    renderWithProviders(<Accounts />);

    await waitFor(() => {
      expect(screen.getByText("No accounts yet")).toBeInTheDocument();
    });
    expect(screen.getByText("Add Account")).toBeInTheDocument();
  });

  it("renders account cards", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(accountList),
    });

    renderWithProviders(<Accounts />);

    await waitFor(() => {
      expect(screen.getByText("Chase Checking")).toBeInTheDocument();
    });
    expect(screen.getByText("Fidelity 401k")).toBeInTheDocument();
    expect(screen.getByText("Checking")).toBeInTheDocument();
    expect(screen.getByText("401(k)")).toBeInTheDocument();
  });

  it("opens create modal on Add Account click", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(emptyAccountList),
    });

    const user = userEvent.setup();
    renderWithProviders(<Accounts />);

    await waitFor(() => {
      expect(screen.getByText("Add Account")).toBeInTheDocument();
    });

    await user.click(screen.getByText("Add Account"));

    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toBeInTheDocument();
    expect(screen.getByLabelText("Institution")).toBeInTheDocument();
    expect(screen.getByLabelText("Account Type")).toBeInTheDocument();
  });

  it("opens edit modal when clicking an account card", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(accountList),
    });

    const user = userEvent.setup();
    renderWithProviders(<Accounts />);

    await waitFor(() => {
      expect(screen.getByText("Chase Checking")).toBeInTheDocument();
    });

    await user.click(screen.getByText("Chase Checking"));

    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Chase Checking")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Chase")).toBeInTheDocument();
  });

  it("submits create form", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(emptyAccountList),
    });

    const user = userEvent.setup();
    renderWithProviders(<Accounts />);

    await waitFor(() => {
      expect(screen.getByText("Add Account")).toBeInTheDocument();
    });

    await user.click(screen.getByText("Add Account"));

    await user.type(screen.getByLabelText("Name"), "New Savings");
    await user.type(screen.getByLabelText("Institution"), "Ally");

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 201,
      json: () =>
        Promise.resolve({
          id: "acc-new",
          name: "New Savings",
          institution: "Ally",
          account_type: "checking",
          currency: "USD",
          is_active: true,
        }),
    });

    const addButtons = screen.getAllByRole("button", { name: "Add Account" });
    const submitBtn = addButtons[addButtons.length - 1] as HTMLElement;
    await user.click(submitBtn);

    await waitFor(() => {
      const calls = mockFetch.mock.calls;
      const lastCall = calls[calls.length - 1] as [string, ...unknown[]];
      expect(lastCall[0]).toContain("/accounts");
    });
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

    renderWithProviders(<Accounts />);

    await waitFor(() => {
      expect(
        screen.getByText("Could not load your accounts. Please try again."),
      ).toBeInTheDocument();
    });

    expect(screen.getByText("Retry")).toBeInTheDocument();
  });

  it("shows delete button in edit modal", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(accountList),
    });

    const user = userEvent.setup();
    renderWithProviders(<Accounts />);

    await waitFor(() => {
      expect(screen.getByText("Chase Checking")).toBeInTheDocument();
    });

    await user.click(screen.getByText("Chase Checking"));

    expect(screen.getByText("Delete Account")).toBeInTheDocument();
  });

  it("submits update form from edit modal", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(accountList),
    });

    const user = userEvent.setup();
    renderWithProviders(<Accounts />);

    await waitFor(() => {
      expect(screen.getByText("Chase Checking")).toBeInTheDocument();
    });

    await user.click(screen.getByText("Chase Checking"));

    const nameInput = screen.getByDisplayValue("Chase Checking");
    await user.clear(nameInput);
    await user.type(nameInput, "Updated Checking");

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          id: "acc-1",
          name: "Updated Checking",
          institution: "Chase",
          account_type: "checking",
          currency: "USD",
          is_active: true,
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

  it("calls delete when Delete Account is clicked", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(accountList),
    });

    const user = userEvent.setup();
    renderWithProviders(<Accounts />);

    await waitFor(() => {
      expect(screen.getByText("Chase Checking")).toBeInTheDocument();
    });

    await user.click(screen.getByText("Chase Checking"));

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.resolve(undefined),
    });

    await user.click(screen.getByText("Delete Account"));

    await waitFor(() => {
      const calls = mockFetch.mock.calls;
      const deleteCalls = calls.filter((c: unknown[]) => {
        const opts = c[1] as RequestInit | undefined;
        return opts?.method === "DELETE";
      });
      expect(deleteCalls.length).toBeGreaterThan(0);
    });
  });

  it("closes edit modal after successful update", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(accountList),
    });

    const user = userEvent.setup();
    renderWithProviders(<Accounts />);

    await waitFor(() => {
      expect(screen.getByText("Chase Checking")).toBeInTheDocument();
    });

    await user.click(screen.getByText("Chase Checking"));
    expect(screen.getByRole("dialog")).toBeInTheDocument();

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          id: "acc-1",
          name: "Chase Checking",
          institution: "Chase",
          account_type: "checking",
          currency: "USD",
          is_active: true,
        }),
    });

    await user.click(screen.getByRole("button", { name: "Save Changes" }));

    await waitFor(() => {
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });
});
