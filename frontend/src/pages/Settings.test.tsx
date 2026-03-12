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
  dismissed: [],
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
        json: () => Promise.resolve(connectedStatus),
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
        json: () => Promise.resolve(connectedStatus),
      });

    renderWithProviders(<Settings />);

    await waitFor(() => {
      expect(screen.getByText("Link Accounts")).toBeInTheDocument();
    });

    expect(screen.getByText("Chase Checking")).toBeInTheDocument();
  });

  it("shows sync error when status has sync_error", async () => {
    const statusWithError = {
      connected: true,
      last_sync_at: "2026-03-10T06:00:00Z",
      sync_error: "Connection timeout",
      linked_accounts: [],
    };

    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(statusWithError),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(statusWithError),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ accounts: [] }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(connectedStatus),
      });

    renderWithProviders(<Settings />);

    await waitFor(() => {
      expect(screen.getByText(/Connection timeout/)).toBeInTheDocument();
    });
  });

  it("calls disconnect on confirm", async () => {
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
      });

    vi.spyOn(window, "confirm").mockReturnValue(true);

    const user = userEvent.setup();
    renderWithProviders(<Settings />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /disconnect/i })).toBeInTheDocument();
    });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.resolve(undefined),
    });

    await user.click(screen.getByRole("button", { name: /disconnect/i }));

    expect(window.confirm).toHaveBeenCalled();

    await waitFor(() => {
      const calls = mockFetch.mock.calls;
      const deleteCalls = calls.filter((c: unknown[]) => {
        const opts = c[1] as RequestInit | undefined;
        return opts?.method === "DELETE";
      });
      expect(deleteCalls.length).toBeGreaterThan(0);
    });

    vi.restoreAllMocks();
  });

  it("does not disconnect when confirm is cancelled", async () => {
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
      });

    vi.spyOn(window, "confirm").mockReturnValue(false);

    const user = userEvent.setup();
    renderWithProviders(<Settings />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /disconnect/i })).toBeInTheDocument();
    });

    const callCountBefore = mockFetch.mock.calls.length;

    await user.click(screen.getByRole("button", { name: /disconnect/i }));

    expect(window.confirm).toHaveBeenCalled();

    const deleteCalls = mockFetch.mock.calls.slice(callCountBefore).filter((c: unknown[]) => {
      const opts = c[1] as RequestInit | undefined;
      return opts?.method === "DELETE";
    });
    expect(deleteCalls).toHaveLength(0);

    vi.restoreAllMocks();
  });

  it("triggers sync and calls sync endpoint", async () => {
    mockFetch.mockImplementation((input: string | URL | Request) => {
      const url =
        typeof input === "string"
          ? input
          : input instanceof URL
            ? input.toString()
            : (input as Request).url;

      if (url.includes("/simplefin/sync")) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: () => Promise.resolve({ updated: 3, errors: [] }),
        });
      }
      if (url.includes("/simplefin/accounts")) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: () => Promise.resolve(simplefinAccounts),
        });
      }
      if (url.includes("/simplefin/status")) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: () => Promise.resolve(connectedStatus),
        });
      }
      if (url.includes("/accounts")) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: () => Promise.resolve(connectedStatus),
        });
      }
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve({}),
      });
    });

    const user = userEvent.setup();
    renderWithProviders(<Settings />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /sync now/i })).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /sync now/i }));

    await waitFor(() => {
      const calls = mockFetch.mock.calls;
      const syncCalls = calls.filter((c: unknown[]) => {
        const url = c[0] as string;
        return url.includes("/simplefin/sync");
      });
      expect(syncCalls.length).toBeGreaterThan(0);
    });
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
