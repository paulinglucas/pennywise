import { useQuery, useMutation, useQueryClient, keepPreviousData } from "@tanstack/react-query";
import {
  listGoals,
  createGoal,
  updateGoal,
  deleteGoal,
  reorderGoals,
  listGoalContributions,
  createGoalContribution,
  deleteGoalContribution,
  type GoalListResponse,
  type GoalContributionListResponse,
  type CreateGoalRequest,
  type UpdateGoalRequest,
  type GoalReorderRequest,
  type CreateGoalContributionRequest,
} from "@/api/client";

export function useGoals(page = 1, perPage = 100) {
  return useQuery<GoalListResponse>({
    queryKey: ["goals", page, perPage],
    queryFn: () => listGoals(page, perPage),
    staleTime: 60 * 1000,
    placeholderData: keepPreviousData,
  });
}

export function useCreateGoal() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateGoalRequest) => createGoal(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["goals"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}

export function useUpdateGoal() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateGoalRequest }) => updateGoal(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["goals"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}

export function useDeleteGoal() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteGoal(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["goals"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}

export function useReorderGoals() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: GoalReorderRequest) => reorderGoals(data),
    onMutate: async (data) => {
      await queryClient.cancelQueries({ queryKey: ["goals"] });
      const previous = queryClient.getQueriesData<GoalListResponse>({ queryKey: ["goals"] });

      queryClient.setQueriesData<GoalListResponse>({ queryKey: ["goals"] }, (old) => {
        if (!old) return old;
        const rankMap = new Map(data.rankings.map((r) => [r.id, r.priority_rank]));
        const updated = old.data.map((goal) => {
          const newRank = rankMap.get(goal.id);
          return newRank !== undefined ? { ...goal, priority_rank: newRank } : goal;
        });
        return { ...old, data: updated };
      });

      return { previous };
    },
    onError: (_err, _data, context) => {
      if (context?.previous) {
        for (const [key, value] of context.previous) {
          queryClient.setQueryData(key, value);
        }
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["goals"] });
    },
  });
}

export function useGoalContributions(goalId: string, page = 1, perPage = 50) {
  return useQuery<GoalContributionListResponse>({
    queryKey: ["goal-contributions", goalId, page, perPage],
    queryFn: () => listGoalContributions(goalId, page, perPage),
    enabled: !!goalId,
    staleTime: 60 * 1000,
    placeholderData: keepPreviousData,
  });
}

export function useCreateGoalContribution() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ goalId, data }: { goalId: string; data: CreateGoalContributionRequest }) =>
      createGoalContribution(goalId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["goals"] });
      queryClient.invalidateQueries({ queryKey: ["goal-contributions"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}

export function useDeleteGoalContribution() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ goalId, contributionId }: { goalId: string; contributionId: string }) =>
      deleteGoalContribution(goalId, contributionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["goals"] });
      queryClient.invalidateQueries({ queryKey: ["goal-contributions"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}
