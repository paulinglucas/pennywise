import { useQuery, keepPreviousData } from "@tanstack/react-query";
import {
  getDashboard,
  getNetWorthHistory,
  type DashboardResponse,
  type NetWorthHistoryResponse,
  type SpendingPeriod,
} from "@/api/client";

export type { SpendingPeriod };

export function useDashboard(spendingPeriod?: SpendingPeriod) {
  return useQuery<DashboardResponse>({
    queryKey: ["dashboard", spendingPeriod ?? "30d"],
    queryFn: () => getDashboard(spendingPeriod),
    staleTime: 60 * 1000,
    placeholderData: keepPreviousData,
  });
}

type NetWorthPeriod = "1m" | "1y" | "5y" | "all";

export function useNetWorthHistory(period: NetWorthPeriod = "1y") {
  return useQuery<NetWorthHistoryResponse>({
    queryKey: ["dashboard", "net-worth-history", period],
    queryFn: () => getNetWorthHistory(period),
    staleTime: 60 * 1000,
    placeholderData: keepPreviousData,
  });
}
