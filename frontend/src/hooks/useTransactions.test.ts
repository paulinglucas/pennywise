import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";
import {
  useTransactions,
  useCreateTransaction,
  useUpdateTransaction,
  useDeleteTransaction,
  useImportTransactions,
} from "./useTransactions";

const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return createElement(QueryClientProvider, { client: queryClient }, children);
  };
}

const mockTransaction = {
  id: "txn-1",
  account_id: "acc-1",
  type: "expense",
  amount: 42.5,
  category: "food",
  date: "2026-03-08",
  notes: "",
  tags: [],
  created_at: "2026-03-08T10:00:00Z",
  updated_at: "2026-03-08T10:00:00Z",
};

const mockTransactionList = {
  data: [mockTransaction],
  pagination: { page: 1, per_page: 25, total: 1, total_pages: 1 },
};

beforeEach(() => {
  mockFetch.mockReset();
});

describe("useTransactions", () => {
  it("fetches transaction list", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockTransactionList),
    });

    const { result } = renderHook(() => useTransactions(), { wrapper: createWrapper() });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.data).toHaveLength(1);
    expect(result.current.data?.data[0]?.category).toBe("food");
  });

  it("passes filters as query parameters", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockTransactionList),
    });

    const { result } = renderHook(
      () => useTransactions({ page: 2, per_page: 10, account_id: "acc-1" }),
      { wrapper: createWrapper() },
    );

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("page=2");
    expect(url).toContain("per_page=10");
    expect(url).toContain("account_id=acc-1");
  });
});

describe("useCreateTransaction", () => {
  it("creates a transaction", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 201,
      json: () => Promise.resolve(mockTransaction),
    });

    const { result } = renderHook(() => useCreateTransaction(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate({
        account_id: "acc-1",
        type: "expense",
        amount: 42.5,
        category: "food",
        date: "2026-03-08",
        tags: [],
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/transactions");
    expect(options.method).toBe("POST");
  });
});

describe("useUpdateTransaction", () => {
  it("updates a transaction", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ ...mockTransaction, amount: 50 }),
    });

    const { result } = renderHook(() => useUpdateTransaction(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate({
        id: "txn-1",
        body: {
          account_id: "acc-1",
          type: "expense",
          amount: 50,
          category: "food",
          date: "2026-03-08",
          tags: [],
        },
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/transactions/txn-1");
    expect(options.method).toBe("PUT");
  });
});

describe("useDeleteTransaction", () => {
  it("deletes a transaction", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.resolve(undefined),
    });

    const { result } = renderHook(() => useDeleteTransaction(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate("txn-1");
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/transactions/txn-1");
    expect(options.method).toBe("DELETE");
  });
});

describe("useImportTransactions", () => {
  it("imports transactions from a file", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 201,
      json: () => Promise.resolve({ imported: 5, skipped: 0, errors: [] }),
    });

    const { result } = renderHook(() => useImportTransactions(), { wrapper: createWrapper() });

    const file = new File(["csv,data"], "transactions.csv", { type: "text/csv" });

    await act(async () => {
      result.current.mutate({ file, accountId: "acc-1" });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/transactions/import");
    expect(options.method).toBe("POST");
  });
});
