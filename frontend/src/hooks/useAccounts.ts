import { useQuery } from "@tanstack/react-query";
import { listAccounts, type AccountListResponse } from "@/api/client";

export function useAccounts() {
  return useQuery<AccountListResponse>({
    queryKey: ["accounts"],
    queryFn: () => listAccounts(),
    staleTime: 5 * 60 * 1000,
  });
}
