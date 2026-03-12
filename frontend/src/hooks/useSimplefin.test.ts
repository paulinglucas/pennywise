import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";
import {
  useSimplefinStatus,
  useSimplefinAccounts,
  useSetupSimplefin,
  useDisconnectSimplefin,
  useLinkSimplefinAccount,
  useUnlinkSimplefinAccount,
  useSyncSimplefin,
} from "./useSimplefin";

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

const mockStatus = {
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

const mockAccounts = {
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

beforeEach(() => {
  mockFetch.mockReset();
});

describe("useSimplefinStatus", () => {
  it("fetches status", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockStatus),
    });

    const { result } = renderHook(() => useSimplefinStatus(), { wrapper: createWrapper() });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.connected).toBe(true);
    expect(result.current.data?.linked_accounts).toHaveLength(1);
  });
});

describe("useSimplefinAccounts", () => {
  it("fetches accounts when enabled", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockAccounts),
    });

    const { result } = renderHook(() => useSimplefinAccounts(true), { wrapper: createWrapper() });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.accounts).toHaveLength(2);
  });

  it("does not fetch when disabled", () => {
    const { result } = renderHook(() => useSimplefinAccounts(false), {
      wrapper: createWrapper(),
    });
    expect(result.current.fetchStatus).toBe("idle");
  });
});

describe("useSetupSimplefin", () => {
  it("sets up simplefin with a token", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ status: "connected" }),
    });

    const { result } = renderHook(() => useSetupSimplefin(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate("my-setup-token");
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/simplefin/setup");
    expect(options.method).toBe("POST");
    const body = JSON.parse(options.body as string);
    expect(body.setup_token).toBe("my-setup-token");
  });
});

describe("useDisconnectSimplefin", () => {
  it("disconnects simplefin", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.resolve(undefined),
    });

    const { result } = renderHook(() => useDisconnectSimplefin(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate();
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/simplefin/");
    expect(options.method).toBe("DELETE");
  });
});

describe("useLinkSimplefinAccount", () => {
  it("links a simplefin account with auto-created local account", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ status: "linked", account_id: "new-acc-1" }),
    });

    const { result } = renderHook(() => useLinkSimplefinAccount(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate({
        simplefin_id: "sfin-1",
        account_type: "checking",
        name: "My Checking",
        institution: "Test Bank",
        balance: "1500.00",
        currency: "USD",
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/simplefin/link");
    expect(options.method).toBe("POST");
    const body = JSON.parse(options.body as string);
    expect(body.simplefin_id).toBe("sfin-1");
    expect(body.account_type).toBe("checking");
    expect(body.name).toBe("My Checking");
  });
});

describe("useUnlinkSimplefinAccount", () => {
  it("unlinks a simplefin account", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.resolve(undefined),
    });

    const { result } = renderHook(() => useUnlinkSimplefinAccount(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      result.current.mutate("acc-1");
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/simplefin/link/acc-1");
    expect(options.method).toBe("DELETE");
  });
});

describe("useSyncSimplefin", () => {
  it("triggers a sync", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ updated: 3, errors: [] }),
    });

    const { result } = renderHook(() => useSyncSimplefin(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate();
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/simplefin/sync");
    expect(options.method).toBe("POST");
  });
});
