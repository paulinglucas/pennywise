import { useQuery } from "@tanstack/react-query";
import { listCategories } from "@/api/client";

export function useCategories() {
  return useQuery<string[]>({
    queryKey: ["categories"],
    queryFn: listCategories,
    staleTime: 60 * 1000,
  });
}
