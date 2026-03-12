import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import type { LinkSimplefinAccountRequest } from "@/api/client";
import {
  getSimplefinStatus,
  setupSimplefin,
  disconnectSimplefin,
  listSimplefinAccounts,
  linkSimplefinAccount,
  unlinkSimplefinAccount,
  dismissSimplefinAccount,
  undismissSimplefinAccount,
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
    mutationFn: (req: LinkSimplefinAccountRequest) => linkSimplefinAccount(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["simplefin"] });
      queryClient.invalidateQueries({ queryKey: ["accounts"] });
      queryClient.invalidateQueries({ queryKey: ["assets"] });
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

export function useDismissSimplefinAccount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (simplefinId: string) => dismissSimplefinAccount(simplefinId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["simplefin"] });
    },
  });
}

export function useUndismissSimplefinAccount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (simplefinId: string) => undismissSimplefinAccount(simplefinId),
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
      queryClient.invalidateQueries({ queryKey: ["accounts"] });
      queryClient.invalidateQueries({ queryKey: ["assets"] });
      queryClient.invalidateQueries({ queryKey: ["transactions"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}
