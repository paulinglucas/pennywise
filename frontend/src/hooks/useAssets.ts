import { useQuery, useMutation, useQueryClient, keepPreviousData } from "@tanstack/react-query";
import {
  listAssets,
  getAsset,
  createAsset,
  updateAsset,
  deleteAsset,
  getAssetHistory,
  getAssetAllocation,
  type AssetListResponse,
  type AssetResponse,
  type CreateAssetRequest,
  type UpdateAssetRequest,
  type AllocationResponse,
} from "@/api/client";

export function useAssets(page = 1, perPage = 100) {
  return useQuery<AssetListResponse>({
    queryKey: ["assets", page, perPage],
    queryFn: () => listAssets(page, perPage),
    staleTime: 60 * 1000,
    placeholderData: keepPreviousData,
  });
}

export function useAsset(id: string | undefined) {
  return useQuery<AssetResponse>({
    queryKey: ["assets", id],
    queryFn: () => getAsset(id!),
    enabled: !!id,
    staleTime: 60 * 1000,
  });
}

export function useAssetHistory(id: string | undefined, period?: string) {
  return useQuery({
    queryKey: ["assets", id, "history", period],
    queryFn: () => getAssetHistory(id!, period as "1m" | "3m" | "6m" | "1y" | "all"),
    enabled: !!id,
    staleTime: 60 * 1000,
  });
}

export function useAssetAllocation(period?: string) {
  return useQuery<AllocationResponse>({
    queryKey: ["assets", "allocation", period],
    queryFn: () => getAssetAllocation(period as "1m" | "3m" | "6m" | "1y" | "all"),
    staleTime: 60 * 1000,
    placeholderData: keepPreviousData,
  });
}

export function useCreateAsset() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateAssetRequest) => createAsset(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["assets"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}

export function useUpdateAsset() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateAssetRequest }) => updateAsset(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["assets"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}

export function useDeleteAsset() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteAsset(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["assets"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}
