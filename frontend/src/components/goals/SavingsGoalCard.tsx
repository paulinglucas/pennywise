import { Plus } from "lucide-react";
import type { GoalResponse } from "@/api/client";
import { formatCurrency, formatDate } from "@/utils/formatting";

interface SavingsGoalCardProps {
  goal: GoalResponse;
  onClick: () => void;
  onContribute: () => void;
}

function OnTrackBadge({ onTrack }: { onTrack: boolean | undefined }) {
  if (onTrack === undefined) return null;

  const label = onTrack ? "On Track" : "Behind";
  const color = onTrack ? "var(--color-positive)" : "var(--color-negative)";

  return (
    <span
      className="inline-block rounded-full px-2 py-0.5 text-xs font-medium"
      style={{ backgroundColor: color + "22", color }}
    >
      {label}
    </span>
  );
}

export default function SavingsGoalCard({ goal, onClick, onContribute }: SavingsGoalCardProps) {
  const progressPercent =
    goal.target_amount > 0 ? (goal.current_amount / goal.target_amount) * 100 : 0;

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") onClick();
      }}
      aria-label={`${goal.name}, ${formatCurrency(goal.current_amount)} of ${formatCurrency(goal.target_amount)}`}
      className="cursor-pointer rounded-lg p-5 transition-all"
      style={{
        backgroundColor: "var(--color-surface)",
        border: "1px solid var(--color-border)",
      }}
    >
      <div className="flex items-start justify-between">
        <div>
          <p className="font-medium" style={{ color: "var(--color-text-primary)" }}>
            {goal.name}
          </p>
          <OnTrackBadge onTrack={goal.on_track} />
        </div>
        <p className="tabular-nums text-lg font-semibold" style={{ color: "var(--color-accent)" }}>
          {formatCurrency(goal.current_amount)}
        </p>
      </div>

      <div className="mt-3">
        <div
          role="progressbar"
          aria-valuenow={progressPercent}
          aria-valuemin={0}
          aria-valuemax={100}
          className="h-2 overflow-hidden rounded-full"
          style={{ backgroundColor: "var(--color-background)" }}
        >
          <div
            className="h-full rounded-full transition-all"
            style={{
              width: `${Math.min(progressPercent, 100)}%`,
              backgroundColor: "var(--color-accent)",
            }}
          />
        </div>
        <div className="mt-1 flex justify-between text-xs">
          <span className="tabular-nums" style={{ color: "var(--color-text-secondary)" }}>
            {progressPercent.toFixed(0)}%
          </span>
          <span className="tabular-nums" style={{ color: "var(--color-text-secondary)" }}>
            of {formatCurrency(goal.target_amount)}
          </span>
        </div>
      </div>

      <div className="mt-3 flex flex-col gap-1 text-xs">
        {goal.required_monthly_contribution !== undefined && (
          <div className="flex justify-between">
            <span style={{ color: "var(--color-text-secondary)" }}>Monthly Needed</span>
            <span className="tabular-nums" style={{ color: "var(--color-text-primary)" }}>
              {formatCurrency(goal.required_monthly_contribution)}
            </span>
          </div>
        )}
        {goal.deadline && (
          <div className="flex justify-between">
            <span style={{ color: "var(--color-text-secondary)" }}>Deadline</span>
            <span style={{ color: "var(--color-text-primary)" }}>{formatDate(goal.deadline)}</span>
          </div>
        )}
      </div>
      <button
        type="button"
        onClick={(event) => {
          event.stopPropagation();
          onContribute();
        }}
        onKeyDown={(event) => event.stopPropagation()}
        className="mt-3 flex w-full items-center justify-center gap-1 rounded-md py-1.5 text-xs font-medium transition-all"
        style={{
          backgroundColor: "var(--color-accent)" + "18",
          color: "var(--color-accent)",
        }}
      >
        <Plus size={14} />
        Contribute
      </button>
    </div>
  );
}
