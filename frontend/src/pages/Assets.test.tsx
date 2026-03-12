import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "@/test-utils";
import Assets from "./Assets";

const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

const emptyAssetList = {
  data: [],
  summary: {
    total_value: 0,
    total_gain_loss: 0,
    allocation: [],
  },
  pagination: { page: 1, per_page: 100, total: 0, total_pages: 0 },
};

const assetList = {
  data: [
    {
      id: "asset-1",
      user_id: "u1",
      name: "S&P 500 ETF",
      asset_type: "investment",
      current_value: 50000,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
    {
      id: "asset-2",
      user_id: "u1",
      name: "Bitcoin",
      asset_type: "crypto",
      current_value: 20000,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
  ],
  summary: {
    total_value: 70000,
    total_gain_loss: 5000,
    allocation: [
      { asset_type: "investment", value: 50000, percentage: 71.43 },
      { asset_type: "crypto", value: 20000, percentage: 28.57 },
    ],
  },
  pagination: { page: 1, per_page: 100, total: 2, total_pages: 1 },
};

const mockAllocation = {
  snapshots: [
    {
      date: "2026-01-01",
      allocations: [{ asset_type: "investment", value: 50000, percentage: 100 }],
    },
  ],
};

const mockHistory = {
  entries: [
    { recorded_at: "2026-01-01T00:00:00Z", value: 45000 },
    { recorded_at: "2026-02-01T00:00:00Z", value: 50000 },
  ],
};

const mockAccounts = {
  data: [
    { id: "acc-1", name: "Checking", account_type: "checking", currency: "USD", is_active: true },
  ],
  pagination: { page: 1, per_page: 100, total: 1, total_pages: 1 },
};

function setupFetch(assetsResponse: unknown) {
  mockFetch.mockImplementation((input: string | URL | Request) => {
    const url =
      typeof input === "string"
        ? input
        : input instanceof URL
          ? input.toString()
          : (input as Request).url;

    if (url.includes("/api/v1/assets/allocation")) {
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockAllocation),
      });
    }
    if (url.includes("/history")) {
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockHistory),
      });
    }
    if (url.includes("/api/v1/assets")) {
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve(assetsResponse),
      });
    }
    if (url.includes("/api/v1/accounts")) {
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockAccounts),
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

describe("Assets", () => {
  it("shows loading skeleton initially", () => {
    mockFetch.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<Assets />);
    expect(document.querySelector(".animate-pulse")).toBeInTheDocument();
  });

  it("shows empty state when no assets exist", async () => {
    setupFetch(emptyAssetList);
    renderWithProviders(<Assets />);

    await waitFor(() => {
      expect(screen.getByText("No assets yet")).toBeInTheDocument();
    });
    expect(screen.getByText("Add Asset")).toBeInTheDocument();
  });

  it("renders asset cards with overview", async () => {
    setupFetch(assetList);
    renderWithProviders(<Assets />);

    await waitFor(() => {
      expect(screen.getByText("S&P 500 ETF")).toBeInTheDocument();
    });
    expect(screen.getByText("Bitcoin")).toBeInTheDocument();
  });

  it("opens create modal on Add Asset click", async () => {
    setupFetch(emptyAssetList);
    const user = userEvent.setup();
    renderWithProviders(<Assets />);

    await waitFor(() => {
      expect(screen.getByText("Add Asset")).toBeInTheDocument();
    });

    await user.click(screen.getByText("Add Asset"));
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });

  it("opens edit modal when clicking an asset card", async () => {
    setupFetch(assetList);
    const user = userEvent.setup();
    renderWithProviders(<Assets />);

    await waitFor(() => {
      expect(screen.getByText("S&P 500 ETF")).toBeInTheDocument();
    });

    await user.click(screen.getByText("S&P 500 ETF"));

    await waitFor(() => {
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });
    expect(screen.getByText("Delete Asset")).toBeInTheDocument();
  });

  it("shows error state on fetch failure", async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
      json: () =>
        Promise.resolve({
          error: { code: "INTERNAL_ERROR", message: "Server error" },
        }),
      headers: new Headers({ "x-request-id": "req-456" }),
    });

    renderWithProviders(<Assets />);

    await waitFor(() => {
      expect(screen.getByText("Could not load your assets. Please try again.")).toBeInTheDocument();
    });
    expect(screen.getByText("Retry")).toBeInTheDocument();
  });

  it("shows period selector buttons", async () => {
    setupFetch(assetList);
    renderWithProviders(<Assets />);

    await waitFor(() => {
      expect(screen.getByText("S&P 500 ETF")).toBeInTheDocument();
    });

    expect(screen.getByText("Holdings")).toBeInTheDocument();
    expect(screen.getAllByText("1M").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("3M").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("6M").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("1Y").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("All").length).toBeGreaterThanOrEqual(1);
  });

  it("deletes asset from edit modal", async () => {
    setupFetch(assetList);
    const user = userEvent.setup();
    renderWithProviders(<Assets />);

    await waitFor(() => {
      expect(screen.getByText("S&P 500 ETF")).toBeInTheDocument();
    });

    await user.click(screen.getByText("S&P 500 ETF"));

    await waitFor(() => {
      expect(screen.getByText("Delete Asset")).toBeInTheDocument();
    });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.resolve(undefined),
    });

    await user.click(screen.getByText("Delete Asset"));

    await waitFor(() => {
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });
});
