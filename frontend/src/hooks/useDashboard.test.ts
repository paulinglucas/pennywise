import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";
import { useDashboard, useNetWorthHistory } from "./useDashboard";

const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return createElement(QueryClientProvider, { client: queryClient }, children);
  };
}

const mockDashboard = {
  net_worth: 150000,
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

const mockHistory = {
  data_points: [
    { date: "2025-01-01", value: 140000 },
    { date: "2025-02-01", value: 150000 },
  ],
};

beforeEach(() => {
  mockFetch.mockReset();
});

describe("useDashboard", () => {
  it("fetches dashboard data", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockDashboard),
    });

    const { result } = renderHook(() => useDashboard(), { wrapper: createWrapper() });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.net_worth).toBe(150000);
    expect(result.current.data?.spending_by_category).toHaveLength(2);
  });

  it("passes spending period as query parameter", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockDashboard),
    });

    const { result } = renderHook(() => useDashboard("7d"), { wrapper: createWrapper() });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("spending_period=7d");
  });
});

describe("useNetWorthHistory", () => {
  it("fetches net worth history with period", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockHistory),
    });

    const { result } = renderHook(() => useNetWorthHistory("1y"), { wrapper: createWrapper() });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.data_points).toHaveLength(2);

    const [url] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("period=1y");
  });
});
