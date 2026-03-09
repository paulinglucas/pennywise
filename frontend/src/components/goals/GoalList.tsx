import { useState, useRef } from "react";
import type { GoalResponse, GoalReorderRequest } from "@/api/client";
import DebtGoalCard from "./DebtGoalCard";
import SavingsGoalCard from "./SavingsGoalCard";

interface GoalListProps {
  goals: GoalResponse[];
  onGoalClick: (goal: GoalResponse) => void;
  onContribute: (goal: GoalResponse) => void;
  onReorder: (data: GoalReorderRequest) => void;
}

function sortByPriority(goals: GoalResponse[]): GoalResponse[] {
  return [...goals].sort((a, b) => a.priority_rank - b.priority_rank);
}

export default function GoalList({ goals, onGoalClick, onContribute, onReorder }: GoalListProps) {
  const debtGoals = sortByPriority(goals.filter((g) => g.goal_type === "debt_payoff"));
  const savingsGoals = sortByPriority(goals.filter((g) => g.goal_type === "savings"));

  return (
    <div className="flex flex-col gap-8">
      {debtGoals.length > 0 && (
        <GoalSection
          title="Debt Payoff"
          goals={debtGoals}
          onGoalClick={onGoalClick}
          onReorder={onReorder}
          renderCard={(goal, onClick) => (
            <DebtGoalCard goal={goal} onClick={onClick} onContribute={() => onContribute(goal)} />
          )}
        />
      )}
      {savingsGoals.length > 0 && (
        <GoalSection
          title="Savings Goals"
          goals={savingsGoals}
          onGoalClick={onGoalClick}
          onReorder={onReorder}
          renderCard={(goal, onClick) => (
            <SavingsGoalCard
              goal={goal}
              onClick={onClick}
              onContribute={() => onContribute(goal)}
            />
          )}
        />
      )}
    </div>
  );
}

interface GoalSectionProps {
  title: string;
  goals: GoalResponse[];
  onGoalClick: (goal: GoalResponse) => void;
  onReorder: (data: GoalReorderRequest) => void;
  renderCard: (goal: GoalResponse, onClick: () => void) => React.ReactNode;
}

function GoalSection({ title, goals, onGoalClick, onReorder, renderCard }: GoalSectionProps) {
  const [dragIndex, setDragIndex] = useState<number | null>(null);
  const [overIndex, setOverIndex] = useState<number | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const dragCounter = useRef(0);

  function handleDragStart(index: number, event: React.DragEvent) {
    setDragIndex(index);
    setIsDragging(true);
    event.dataTransfer.effectAllowed = "move";
  }

  function handleDragEnter(index: number) {
    dragCounter.current++;
    setOverIndex(index);
  }

  function handleDragOver(event: React.DragEvent) {
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
  }

  function resetDrag() {
    setDragIndex(null);
    setOverIndex(null);
    setIsDragging(false);
    dragCounter.current = 0;
  }

  function handleDragLeave() {
    dragCounter.current--;
    if (dragCounter.current === 0) {
      setOverIndex(null);
    }
  }

  function handleDrop(targetIndex: number) {
    if (dragIndex === null || dragIndex === targetIndex) {
      resetDrag();
      return;
    }

    const reordered = [...goals];
    const moved = reordered.splice(dragIndex, 1)[0];
    if (!moved) {
      resetDrag();
      return;
    }
    reordered.splice(targetIndex, 0, moved);

    const rankings = reordered.map((goal, idx) => ({
      id: goal.id,
      priority_rank: idx + 1,
    }));

    onReorder({ rankings });
    resetDrag();
  }

  function handleCardClick(goal: GoalResponse) {
    if (isDragging) return;
    onGoalClick(goal);
  }

  return (
    <div>
      <h2 className="mb-3 text-lg font-semibold" style={{ color: "var(--color-text-primary)" }}>
        {title}
      </h2>
      <div className="flex flex-col gap-3">
        {goals.map((goal, index) => (
          <div
            key={goal.id}
            draggable
            onDragStart={(e) => handleDragStart(index, e)}
            onDragEnter={() => handleDragEnter(index)}
            onDragLeave={handleDragLeave}
            onDragOver={handleDragOver}
            onDrop={() => handleDrop(index)}
            onDragEnd={resetDrag}
            className="transition-opacity"
            style={{
              opacity: dragIndex === index ? 0.5 : 1,
              borderTop:
                overIndex === index && dragIndex !== null && dragIndex !== index
                  ? "2px solid var(--color-accent)"
                  : "2px solid transparent",
            }}
          >
            {renderCard(goal, () => handleCardClick(goal))}
          </div>
        ))}
      </div>
    </div>
  );
}
