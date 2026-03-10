import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "@/test-utils";
import Settings from "./Settings";

const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

const disconnectedStatus = {
  connected: false,
  linked_accounts: [],
};

const connectedStatus = {
  connected: true,
  last_sync_at: "2026-03-10T06:00:00Z",
  linked_accounts: [
    {
      account_id: "acc-1",
      simplefin_id: "sfin-1",
      account_name: "Checking",
      institution: "Chase",
    },
  ],
};

const simplefinAccounts = {
  accounts: [
    {
      id: "sfin-1",
      name: "Chase Checking",
      institution: "Chase",
      balance: "5000.00",
      currency: "USD",
    },
    { id: "sfin-2", name: "Savings", institution: "Ally", balance: "20000.00", currency: "USD" },
  ],
};

const accountsList = {
  data: [
    { id: "acc-1", name: "Checking", type: "checking" },
    { id: "acc-2", name: "Savings", type: "savings" },
  ],
  pagination: { page: 1, per_page: 100, total: 2, total_pages: 1 },
};

beforeEach(() => {
  mockFetch.mockReset();
});

describe("Settings", () => {
  it("shows loading state initially", () => {
    mockFetch.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<Settings />);

    expect(screen.getByText("Settings")).toBeInTheDocument();
    expect(screen.getByText("Loading...")).toBeInTheDocument();
  });

  it("shows connect form when not connected", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(disconnectedStatus),
    });

    renderWithProviders(<Settings />);

    await waitFor(() => {
      expect(screen.getByText("Connect SimpleFIN")).toBeInTheDocument();
    });

    expect(screen.getByPlaceholderText("Paste your SimpleFIN setup token")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /connect/i })).toBeInTheDocument();
  });

  it("disables connect button when token is empty", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(disconnectedStatus),
    });

    renderWithProviders(<Settings />);

    await waitFor(() => {
      expect(screen.getByText("Connect SimpleFIN")).toBeInTheDocument();
    });

    const button = screen.getByRole("button", { name: /connect/i });
    expect(button).toBeDisabled();
  });

  it("enables connect button when token is entered", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(disconnectedStatus),
    });

    const user = userEvent.setup();
    renderWithProviders(<Settings />);

    await waitFor(() => {
      expect(screen.getByText("Connect SimpleFIN")).toBeInTheDocument();
    });

    const input = screen.getByPlaceholderText("Paste your SimpleFIN setup token");
    await user.type(input, "my-setup-token");

    const button = screen.getByRole("button", { name: /connect/i });
    expect(button).not.toBeDisabled();
  });

  it("shows connected status with sync info", async () => {
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(connectedStatus),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(connectedStatus),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(simplefinAccounts),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(accountsList),
      });

    renderWithProviders(<Settings />);

    await waitFor(() => {
      expect(screen.getByText("SimpleFIN")).toBeInTheDocument();
    });

    expect(screen.getByText("Connected")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /sync now/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /disconnect/i })).toBeInTheDocument();
    expect(screen.getByText("1 account linked")).toBeInTheDocument();
  });

  it("shows linked accounts in account mapper", async () => {
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(connectedStatus),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(connectedStatus),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(simplefinAccounts),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(accountsList),
      });

    renderWithProviders(<Settings />);

    await waitFor(() => {
      expect(screen.getByText("Link Accounts")).toBeInTheDocument();
    });

    expect(screen.getByText("Chase Checking")).toBeInTheDocument();
  });

  it("shows error message when setup fails", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(disconnectedStatus),
    });

    const user = userEvent.setup();
    renderWithProviders(<Settings />);

    await waitFor(() => {
      expect(screen.getByText("Connect SimpleFIN")).toBeInTheDocument();
    });

    const input = screen.getByPlaceholderText("Paste your SimpleFIN setup token");
    await user.type(input, "bad-token");

    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 400,
      json: () =>
        Promise.resolve({
          error: { code: "VALIDATION_ERROR", message: "Invalid token" },
        }),
    });

    const button = screen.getByRole("button", { name: /connect/i });
    await user.click(button);

    await waitFor(() => {
      expect(screen.getByText(/failed to connect/i)).toBeInTheDocument();
    });
  });
});
