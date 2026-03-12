import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";
import {
  useAssets,
  useAsset,
  useAssetHistory,
  useAssetAllocation,
  useCreateAsset,
  useUpdateAsset,
  useDeleteAsset,
} from "./useAssets";

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

const mockAssetList = {
  data: [
    {
      id: "asset-1",
      name: "Stocks",
      asset_type: "investment",
      current_value: 50000,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
  ],
  summary: {
    total_value: 50000,
    allocation: [{ asset_type: "investment", value: 50000, percentage: 100 }],
  },
  pagination: { page: 1, per_page: 100, total: 1, total_pages: 1 },
};

const mockAsset = {
  id: "asset-1",
  name: "Stocks",
  asset_type: "investment",
  current_value: 50000,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

const mockHistory = {
  entries: [
    { date: "2026-01-01", value: 45000 },
    { date: "2026-02-01", value: 50000 },
  ],
};

const mockAllocation = {
  snapshots: [
    {
      date: "2026-01-01",
      allocations: [{ asset_type: "investment", value: 50000, percentage: 100 }],
    },
  ],
};

beforeEach(() => {
  mockFetch.mockReset();
});

describe("useAssets", () => {
  it("fetches asset list", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockAssetList),
    });

    const { result } = renderHook(() => useAssets(), { wrapper: createWrapper() });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.data).toHaveLength(1);
    expect(result.current.data?.data[0]?.name).toBe("Stocks");
  });

  it("passes pagination parameters", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockAssetList),
    });

    const { result } = renderHook(() => useAssets(2, 50), { wrapper: createWrapper() });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("page=2");
    expect(url).toContain("per_page=50");
  });
});

describe("useAsset", () => {
  it("fetches a single asset by id", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockAsset),
    });

    const { result } = renderHook(() => useAsset("asset-1"), { wrapper: createWrapper() });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.name).toBe("Stocks");
  });

  it("does not fetch when id is undefined", () => {
    const { result } = renderHook(() => useAsset(undefined), { wrapper: createWrapper() });
    expect(result.current.fetchStatus).toBe("idle");
  });
});

describe("useAssetHistory", () => {
  it("fetches asset history with period", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockHistory),
    });

    const { result } = renderHook(() => useAssetHistory("asset-1", "1y"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.entries).toHaveLength(2);

    const [url] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("period=1y");
  });

  it("does not fetch when id is undefined", () => {
    const { result } = renderHook(() => useAssetHistory(undefined), { wrapper: createWrapper() });
    expect(result.current.fetchStatus).toBe("idle");
  });
});

describe("useAssetAllocation", () => {
  it("fetches allocation data with period", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockAllocation),
    });

    const { result } = renderHook(() => useAssetAllocation("1y"), { wrapper: createWrapper() });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.snapshots).toHaveLength(1);

    const [url] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("period=1y");
  });
});

describe("useCreateAsset", () => {
  it("creates an asset and invalidates queries", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 201,
      json: () => Promise.resolve(mockAsset),
    });

    const { result } = renderHook(() => useCreateAsset(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate({
        name: "Stocks",
        asset_type: "brokerage",
        current_value: 50000,
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/assets");
    expect(options.method).toBe("POST");
  });
});

describe("useUpdateAsset", () => {
  it("updates an asset", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ ...mockAsset, name: "Updated Stocks" }),
    });

    const { result } = renderHook(() => useUpdateAsset(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate({
        id: "asset-1",
        data: { name: "Updated Stocks" },
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/assets/asset-1");
    expect(options.method).toBe("PUT");
  });
});

describe("useDeleteAsset", () => {
  it("deletes an asset", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.resolve(undefined),
    });

    const { result } = renderHook(() => useDeleteAsset(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate("asset-1");
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/assets/asset-1");
    expect(options.method).toBe("DELETE");
  });
});
