import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";
import {
  useTransactionGroups,
  useTransactionGroup,
  useCreateTransactionGroup,
  useUpdateTransactionGroup,
  useDeleteTransactionGroup,
} from "./useTransactionGroups";

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

const mockGroup = {
  id: "grp-1",
  name: "Rent Split",
  description: "Monthly rent payment",
  transactions: [],
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

const mockGroupList = {
  data: [mockGroup],
  pagination: { page: 1, per_page: 20, total: 1, total_pages: 1 },
};

beforeEach(() => {
  mockFetch.mockReset();
});

describe("useTransactionGroups", () => {
  it("fetches transaction group list", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockGroupList),
    });

    const { result } = renderHook(() => useTransactionGroups(), { wrapper: createWrapper() });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.data).toHaveLength(1);
    expect(result.current.data?.data[0]?.name).toBe("Rent Split");
  });

  it("passes pagination parameters", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockGroupList),
    });

    const { result } = renderHook(() => useTransactionGroups(3, 10), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("page=3");
    expect(url).toContain("per_page=10");
  });
});

describe("useTransactionGroup", () => {
  it("fetches a single group by id", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockGroup),
    });

    const { result } = renderHook(() => useTransactionGroup("grp-1"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.name).toBe("Rent Split");
  });

  it("does not fetch when id is null", () => {
    const { result } = renderHook(() => useTransactionGroup(null), { wrapper: createWrapper() });
    expect(result.current.fetchStatus).toBe("idle");
  });
});

describe("useCreateTransactionGroup", () => {
  it("creates a transaction group", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 201,
      json: () => Promise.resolve(mockGroup),
    });

    const { result } = renderHook(() => useCreateTransactionGroup(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      result.current.mutate({
        name: "Rent Split",
        members: [
          {
            amount: 500,
            type: "expense",
            category: "housing",
            date: "2026-03-01",
            account_id: "acc-1",
          },
          {
            amount: 500,
            type: "expense",
            category: "housing",
            date: "2026-03-01",
            account_id: "acc-1",
          },
        ],
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/transaction-groups");
    expect(options.method).toBe("POST");
  });
});

describe("useUpdateTransactionGroup", () => {
  it("updates a transaction group", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ ...mockGroup, name: "Updated Split" }),
    });

    const { result } = renderHook(() => useUpdateTransactionGroup(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      result.current.mutate({
        id: "grp-1",
        body: { name: "Updated Split" },
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/transaction-groups/grp-1");
    expect(options.method).toBe("PUT");
  });
});

describe("useDeleteTransactionGroup", () => {
  it("deletes a transaction group", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.resolve(undefined),
    });

    const { result } = renderHook(() => useDeleteTransactionGroup(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      result.current.mutate("grp-1");
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/transaction-groups/grp-1");
    expect(options.method).toBe("DELETE");
  });
});
