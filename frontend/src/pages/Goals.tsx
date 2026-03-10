import { useState } from "react";
import { Plus } from "lucide-react";
import {
  useGoals,
  useCreateGoal,
  useUpdateGoal,
  useDeleteGoal,
  useReorderGoals,
  useCreateGoalContribution,
} from "@/hooks/useGoals";
import { useTransactions } from "@/hooks/useTransactions";
import type {
  GoalResponse,
  CreateGoalRequest,
  UpdateGoalRequest,
  CreateGoalContributionRequest,
} from "@/api/client";
import GoalSummaryBar from "@/components/goals/GoalSummaryBar";
import GoalList from "@/components/goals/GoalList";
import GoalForm from "@/components/goals/GoalForm";
import ContributeForm from "@/components/goals/ContributeForm";
import Modal from "@/components/shared/Modal";
import EmptyState from "@/components/shared/EmptyState";
import ErrorState, { extractRequestId } from "@/components/shared/ErrorState";
import { SkeletonCard } from "@/components/shared/Skeleton";

type FormMode = { kind: "closed" } | { kind: "create" } | { kind: "edit"; goal: GoalResponse };
type ContributeTarget = GoalResponse | null;

function GoalsSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <SkeletonCard />
        <SkeletonCard />
      </div>
      <SkeletonCard />
      <SkeletonCard />
      <SkeletonCard />
    </div>
  );
}

export default function Goals() {
  const [formMode, setFormMode] = useState<FormMode>({ kind: "closed" });
  const [contributeTarget, setContributeTarget] = useState<ContributeTarget>(null);
  const goals = useGoals();
  const createGoal = useCreateGoal();
  const updateGoal = useUpdateGoal();
  const deleteGoal = useDeleteGoal();
  const reorderGoals = useReorderGoals();
  const contributeGoal = useCreateGoalContribution();
  const recentTransactions = useTransactions({ per_page: 50 });

  if (goals.isError) {
    return (
      <ErrorState
        message="Could not load your goals. Please try again."
        onRetry={() => goals.refetch()}
        requestId={extractRequestId(goals.error)}
      />
    );
  }

  if (!goals.data) {
    return <GoalsSkeleton />;
  }

  const goalList = goals.data.data;

  function handleCreate(data: CreateGoalRequest) {
    createGoal.mutate(data, {
      onSuccess: () => setFormMode({ kind: "closed" }),
    });
  }

  function handleUpdate(data: CreateGoalRequest) {
    if (formMode.kind !== "edit") return;
    const updateData: UpdateGoalRequest = {
      name: data.name,
      goal_type: data.goal_type,
      target_amount: data.target_amount,
      current_amount: data.current_amount,
      deadline: data.deadline,
    };
    updateGoal.mutate(
      { id: formMode.goal.id, data: updateData },
      { onSuccess: () => setFormMode({ kind: "closed" }) },
    );
  }

  function handleDelete(id: string) {
    deleteGoal.mutate(id);
  }

  function handleContribute(data: { amount: number; notes?: string; transaction_id?: string }) {
    if (!contributeTarget) return;
    const req: CreateGoalContributionRequest = {
      amount: data.amount,
      notes: data.notes,
      transaction_id: data.transaction_id,
    };
    contributeGoal.mutate(
      { goalId: contributeTarget.id, data: req },
      { onSuccess: () => setContributeTarget(null) },
    );
  }

  const addButton = (
    <button
      onClick={() => setFormMode({ kind: "create" })}
      className="btn-primary flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-all"
      style={{
        backgroundColor: "var(--color-accent)",
        color: "var(--color-background)",
        boxShadow: "var(--glow-accent)",
      }}
    >
      <Plus size={16} />
      Add Goal
    </button>
  );

  const header = (
    <div className="flex items-center justify-between">
      <h1 className="text-2xl font-semibold" style={{ color: "var(--color-text-primary)" }}>
        Goals
      </h1>
      {addButton}
    </div>
  );

  const createModal = (
    <Modal
      isOpen={formMode.kind === "create"}
      onClose={() => setFormMode({ kind: "closed" })}
      title="Add Goal"
    >
      <GoalForm
        onSubmit={handleCreate}
        onCancel={() => setFormMode({ kind: "closed" })}
        isSubmitting={createGoal.isPending}
      />
    </Modal>
  );

  if (goalList.length === 0) {
    return (
      <div className="flex flex-col gap-6">
        {header}
        <EmptyState
          title="No goals yet"
          description="Set your first financial goal. What are you working toward?"
        />
        {createModal}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      {header}
      <GoalSummaryBar goals={goalList} />
      <GoalList
        goals={goalList}
        onGoalClick={(goal) => setFormMode({ kind: "edit", goal })}
        onContribute={(goal) => setContributeTarget(goal)}
        onReorder={(data) => reorderGoals.mutate(data)}
      />
      {createModal}
      <Modal
        isOpen={formMode.kind === "edit"}
        onClose={() => setFormMode({ kind: "closed" })}
        title="Edit Goal"
      >
        {formMode.kind === "edit" && (
          <div className="flex flex-col gap-4">
            <GoalForm
              onSubmit={handleUpdate}
              onCancel={() => setFormMode({ kind: "closed" })}
              isSubmitting={updateGoal.isPending}
              initialValues={{
                name: formMode.goal.name,
                goal_type: formMode.goal.goal_type,
                target_amount: formMode.goal.target_amount,
                current_amount: formMode.goal.current_amount,
                deadline: formMode.goal.deadline,
              }}
            />
            <button
              type="button"
              onClick={() => {
                handleDelete(formMode.goal.id);
                setFormMode({ kind: "closed" });
              }}
              className="rounded-md px-4 py-2 text-sm font-medium transition-colors"
              style={{ color: "var(--color-negative)" }}
            >
              Delete Goal
            </button>
          </div>
        )}
      </Modal>
      <Modal
        isOpen={contributeTarget !== null}
        onClose={() => setContributeTarget(null)}
        title={contributeTarget?.goal_type === "debt_payoff" ? "Make Payment" : "Contribute"}
      >
        {contributeTarget && (
          <ContributeForm
            onSubmit={handleContribute}
            onCancel={() => setContributeTarget(null)}
            isSubmitting={contributeGoal.isPending}
            transactions={recentTransactions.data?.data}
          />
        )}
      </Modal>
    </div>
  );
}
