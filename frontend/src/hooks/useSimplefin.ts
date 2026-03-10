import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getSimplefinStatus,
  setupSimplefin,
  disconnectSimplefin,
  listSimplefinAccounts,
  linkSimplefinAccount,
  unlinkSimplefinAccount,
  triggerSimplefinSync,
} from "@/api/client";

export function useSimplefinStatus() {
  return useQuery({
    queryKey: ["simplefin", "status"],
    queryFn: getSimplefinStatus,
  });
}

export function useSimplefinAccounts(enabled: boolean) {
  return useQuery({
    queryKey: ["simplefin", "accounts"],
    queryFn: listSimplefinAccounts,
    enabled,
  });
}

export function useSetupSimplefin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (setupToken: string) => setupSimplefin(setupToken),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["simplefin"] });
    },
  });
}

export function useDisconnectSimplefin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: disconnectSimplefin,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["simplefin"] });
    },
  });
}

export function useLinkSimplefinAccount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ accountId, simplefinId }: { accountId: string; simplefinId: string }) =>
      linkSimplefinAccount(accountId, simplefinId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["simplefin"] });
    },
  });
}

export function useUnlinkSimplefinAccount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (accountId: string) => unlinkSimplefinAccount(accountId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["simplefin"] });
    },
  });
}

export function useSyncSimplefin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: triggerSimplefinSync,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["simplefin"] });
      queryClient.invalidateQueries({ queryKey: ["assets"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}
