import { useQuery, useMutation, useQueryClient, keepPreviousData } from "@tanstack/react-query";
import {
  listTransactions,
  createTransaction,
  updateTransaction,
  deleteTransaction,
  importTransactions,
  type TransactionFilters,
  type TransactionListResponse,
  type TransactionResponse,
  type CreateTransactionRequest,
  type UpdateTransactionRequest,
  type ImportResponse,
} from "@/api/client";

export type { TransactionFilters };

export function useTransactions(filters?: TransactionFilters) {
  return useQuery<TransactionListResponse>({
    queryKey: ["transactions", filters ?? {}],
    queryFn: () => listTransactions(filters),
    staleTime: 30 * 1000,
    placeholderData: keepPreviousData,
  });
}

export function useCreateTransaction() {
  const queryClient = useQueryClient();
  return useMutation<TransactionResponse, Error, CreateTransactionRequest>({
    mutationFn: createTransaction,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["transactions"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      queryClient.invalidateQueries({ queryKey: ["categories"] });
    },
  });
}

export function useUpdateTransaction() {
  const queryClient = useQueryClient();
  return useMutation<TransactionResponse, Error, { id: string; body: UpdateTransactionRequest }>({
    mutationFn: ({ id, body }) => updateTransaction(id, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["transactions"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      queryClient.invalidateQueries({ queryKey: ["categories"] });
    },
  });
}

export function useDeleteTransaction() {
  const queryClient = useQueryClient();
  return useMutation<void, Error, string>({
    mutationFn: deleteTransaction,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["transactions"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      queryClient.invalidateQueries({ queryKey: ["categories"] });
    },
  });
}

export function useImportTransactions() {
  const queryClient = useQueryClient();
  return useMutation<ImportResponse, Error, { file: File; accountId: string }>({
    mutationFn: ({ file, accountId }) => importTransactions(file, accountId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["transactions"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      queryClient.invalidateQueries({ queryKey: ["categories"] });
    },
  });
}
