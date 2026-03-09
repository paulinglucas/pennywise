import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  listTransactionGroups,
  getTransactionGroup,
  createTransactionGroup,
  updateTransactionGroup,
  deleteTransactionGroup,
  type TransactionGroupResponse,
  type TransactionGroupListResponse,
  type CreateTransactionGroupRequest,
  type UpdateTransactionGroupRequest,
} from "@/api/client";

export function useTransactionGroups(page = 1, perPage = 20) {
  return useQuery<TransactionGroupListResponse>({
    queryKey: ["transaction-groups", page, perPage],
    queryFn: () => listTransactionGroups(page, perPage),
    staleTime: 30 * 1000,
  });
}

export function useTransactionGroup(id: string | null) {
  return useQuery<TransactionGroupResponse>({
    queryKey: ["transaction-groups", id],
    queryFn: () => getTransactionGroup(id!),
    enabled: id !== null,
    staleTime: 30 * 1000,
  });
}

export function useCreateTransactionGroup() {
  const queryClient = useQueryClient();
  return useMutation<TransactionGroupResponse, Error, CreateTransactionGroupRequest>({
    mutationFn: createTransactionGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["transaction-groups"] });
      queryClient.invalidateQueries({ queryKey: ["transactions"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}

export function useUpdateTransactionGroup() {
  const queryClient = useQueryClient();
  return useMutation<
    TransactionGroupResponse,
    Error,
    { id: string; body: UpdateTransactionGroupRequest }
  >({
    mutationFn: ({ id, body }) => updateTransactionGroup(id, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["transaction-groups"] });
      queryClient.invalidateQueries({ queryKey: ["transactions"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}

export function useDeleteTransactionGroup() {
  const queryClient = useQueryClient();
  return useMutation<void, Error, string>({
    mutationFn: deleteTransactionGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["transaction-groups"] });
      queryClient.invalidateQueries({ queryKey: ["transactions"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}
