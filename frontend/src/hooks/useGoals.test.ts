import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";
import {
  useGoals,
  useCreateGoal,
  useUpdateGoal,
  useDeleteGoal,
  useReorderGoals,
  useGoalContributions,
  useCreateGoalContribution,
  useDeleteGoalContribution,
} from "./useGoals";

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

const mockGoal = {
  id: "goal-1",
  name: "Emergency Fund",
  goal_type: "savings",
  target_amount: 10000,
  current_amount: 5000,
  priority_rank: 1,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

const mockGoalList = {
  data: [mockGoal],
  pagination: { page: 1, per_page: 100, total: 1, total_pages: 1 },
};

const mockContributionList = {
  data: [
    {
      id: "contrib-1",
      goal_id: "goal-1",
      amount: 500,
      notes: "Monthly savings",
      created_at: "2026-02-01T00:00:00Z",
    },
  ],
  pagination: { page: 1, per_page: 50, total: 1, total_pages: 1 },
};

beforeEach(() => {
  mockFetch.mockReset();
});

describe("useGoals", () => {
  it("fetches goal list", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockGoalList),
    });

    const { result } = renderHook(() => useGoals(), { wrapper: createWrapper() });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.data).toHaveLength(1);
    expect(result.current.data?.data[0]?.name).toBe("Emergency Fund");
  });

  it("passes pagination parameters", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockGoalList),
    });

    const { result } = renderHook(() => useGoals(2, 50), { wrapper: createWrapper() });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("page=2");
    expect(url).toContain("per_page=50");
  });
});

describe("useCreateGoal", () => {
  it("creates a goal", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 201,
      json: () => Promise.resolve(mockGoal),
    });

    const { result } = renderHook(() => useCreateGoal(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate({
        name: "Emergency Fund",
        goal_type: "savings",
        target_amount: 10000,
        current_amount: 5000,
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/goals");
    expect(options.method).toBe("POST");
  });
});

describe("useUpdateGoal", () => {
  it("updates a goal", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ ...mockGoal, name: "Updated Fund" }),
    });

    const { result } = renderHook(() => useUpdateGoal(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate({
        id: "goal-1",
        data: { name: "Updated Fund" },
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/goals/goal-1");
    expect(options.method).toBe("PUT");
  });
});

describe("useDeleteGoal", () => {
  it("deletes a goal", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.resolve(undefined),
    });

    const { result } = renderHook(() => useDeleteGoal(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate("goal-1");
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/goals/goal-1");
    expect(options.method).toBe("DELETE");
  });
});

describe("useReorderGoals", () => {
  it("reorders goals with optimistic update", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ status: "ok" }),
    });

    const { result } = renderHook(() => useReorderGoals(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate({
        rankings: [
          { id: "goal-1", priority_rank: 2 },
          { id: "goal-2", priority_rank: 1 },
        ],
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/goals/reorder");
    expect(options.method).toBe("PUT");
  });
});

describe("useGoalContributions", () => {
  it("fetches contributions for a goal", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockContributionList),
    });

    const { result } = renderHook(() => useGoalContributions("goal-1"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.data).toHaveLength(1);
    expect(result.current.data?.data[0]?.amount).toBe(500);

    const [url] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/goals/goal-1/contributions");
  });

  it("does not fetch when goalId is empty", () => {
    const { result } = renderHook(() => useGoalContributions(""), { wrapper: createWrapper() });
    expect(result.current.fetchStatus).toBe("idle");
  });
});

describe("useCreateGoalContribution", () => {
  it("creates a contribution", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 201,
      json: () =>
        Promise.resolve({
          id: "contrib-2",
          goal_id: "goal-1",
          amount: 200,
          created_at: "2026-03-01T00:00:00Z",
        }),
    });

    const { result } = renderHook(() => useCreateGoalContribution(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      result.current.mutate({
        goalId: "goal-1",
        data: { amount: 200, notes: "Extra savings" },
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/goals/goal-1/contributions");
    expect(options.method).toBe("POST");
  });
});

describe("useDeleteGoalContribution", () => {
  it("deletes a contribution", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.resolve(undefined),
    });

    const { result } = renderHook(() => useDeleteGoalContribution(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      result.current.mutate({
        goalId: "goal-1",
        contributionId: "contrib-1",
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/v1/goals/goal-1/contributions/contrib-1");
    expect(options.method).toBe("DELETE");
  });
});
