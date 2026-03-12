import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import Dashboard from "./Dashboard";

const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

const mockDashboardData = {
  net_worth: 150000,
  net_worth_breakdown: { assets: 400000, cash: 5000, debt: 255000 },
  cash_flow_this_month: 2500,
  spending_by_category: [
    { category: "Food", amount: 800, percentage: 40 },
    { category: "Housing", amount: 1200, percentage: 60 },
  ],
  debts_summary: [
    {
      account_id: "a1",
      name: "Mortgage",
      balance: 250000,
      monthly_payment: 1500,
      payoff_date: "2050-01-01",
      months_remaining: 300,
    },
  ],
};

const mockHistoryData = {
  data_points: [
    { date: "2025-01-01", value: 140000 },
    { date: "2025-02-01", value: 150000 },
  ],
};

beforeEach(() => {
  mockFetch.mockReset();
});

function setupSuccessfulResponses() {
  mockFetch
    .mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockDashboardData),
    })
    .mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockHistoryData),
    });
}

describe("Dashboard", () => {
  it("renders all dashboard components with data", async () => {
    setupSuccessfulResponses();
    renderWithProviders(<Dashboard />);

    await waitFor(() => {
      expect(screen.getAllByText("Net Worth").length).toBeGreaterThanOrEqual(1);
    });

    expect(screen.getAllByText("$150,000.00").length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText("$2,500.00")).toBeInTheDocument();
    expect(screen.getByText("Spending Breakdown")).toBeInTheDocument();
    expect(screen.getByText("Debt Tracker")).toBeInTheDocument();
    expect(screen.getByText("Insights")).toBeInTheDocument();
  });

  it("renders loading skeletons while data is fetching", () => {
    mockFetch.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<Dashboard />);

    const skeletons = document.querySelectorAll(".animate-pulse");
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it("renders empty state when dashboard returns empty data", async () => {
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () =>
          Promise.resolve({
            net_worth: 0,
            net_worth_breakdown: { assets: 0, cash: 0, debt: 0 },
            cash_flow_this_month: 0,
            spending_by_category: [],
            debts_summary: [],
          }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ data_points: [] }),
      });

    renderWithProviders(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText("Net Worth")).toBeInTheDocument();
    });

    expect(screen.getByText("Welcome to Pennywise")).toBeInTheDocument();
    expect(
      screen.getByText(
        "Connect your bank accounts to automatically sync balances and transactions.",
      ),
    ).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Connect Bank Accounts" })).toBeInTheDocument();
  });

  it("renders error state on API failure with retry button", async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
      statusText: "Internal Server Error",
      json: () => Promise.reject(new Error("not json")),
    });

    renderWithProviders(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText("Something went wrong")).toBeInTheDocument();
    });

    expect(screen.getByRole("button", { name: /retry/i })).toBeInTheDocument();
  });

  it("displays request ID in error state when available", async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
      json: () =>
        Promise.resolve({
          error: { code: "INTERNAL_ERROR", message: "Server error", request_id: "req-dash-123" },
        }),
    });

    renderWithProviders(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText("Something went wrong")).toBeInTheDocument();
    });

    expect(screen.getByText("Request ID: req-dash-123")).toBeInTheDocument();
  });
});
