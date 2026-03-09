import { useState, type FormEvent } from "react";
import type { GoalType, CreateGoalRequest } from "@/api/client";

interface GoalFormProps {
  onSubmit: (data: CreateGoalRequest) => void;
  onCancel: () => void;
  isSubmitting?: boolean;
  initialValues?: Partial<CreateGoalRequest>;
}

const goalTypes: { key: GoalType; label: string }[] = [
  { key: "savings", label: "Savings" },
  { key: "debt_payoff", label: "Debt Payoff" },
];

const inputStyle = {
  backgroundColor: "var(--color-background)",
  borderColor: "var(--color-border)",
  color: "var(--color-text-primary)",
};

export default function GoalForm({
  onSubmit,
  onCancel,
  isSubmitting = false,
  initialValues,
}: GoalFormProps) {
  const [name, setName] = useState(initialValues?.name ?? "");
  const [goalType, setGoalType] = useState<GoalType>(initialValues?.goal_type ?? "savings");
  const [targetAmount, setTargetAmount] = useState(initialValues?.target_amount?.toString() ?? "");
  const [currentAmount, setCurrentAmount] = useState(
    initialValues?.current_amount?.toString() ?? "",
  );
  const [deadline, setDeadline] = useState(initialValues?.deadline ?? "");

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    const data: CreateGoalRequest = {
      name,
      goal_type: goalType,
      target_amount: parseFloat(targetAmount),
    };
    if (currentAmount) {
      data.current_amount = parseFloat(currentAmount);
    }
    if (deadline) {
      data.deadline = deadline;
    }
    onSubmit(data);
  }

  const isEditing = !!initialValues;

  return (
    <form role="form" onSubmit={handleSubmit} className="flex flex-col gap-4">
      <FormField label="Goal Type" htmlFor="goal-type">
        <select
          id="goal-type"
          value={goalType}
          onChange={(e) => setGoalType(e.target.value as GoalType)}
          className="form-input"
          style={inputStyle}
        >
          {goalTypes.map((t) => (
            <option key={t.key} value={t.key}>
              {t.label}
            </option>
          ))}
        </select>
      </FormField>

      <FormField label="Name" htmlFor="goal-name">
        <input
          id="goal-name"
          type="text"
          required
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="form-input"
          style={inputStyle}
          placeholder={goalType === "debt_payoff" ? "e.g. Credit Card" : "e.g. Emergency Fund"}
        />
      </FormField>

      <FormField label="Target Amount" htmlFor="goal-target">
        <input
          id="goal-target"
          type="number"
          required
          min="0.01"
          step="0.01"
          value={targetAmount}
          onChange={(e) => setTargetAmount(e.target.value)}
          className="form-input tabular-nums"
          style={inputStyle}
          placeholder="0.00"
        />
      </FormField>

      <FormField label="Current Amount" htmlFor="goal-current">
        <input
          id="goal-current"
          type="number"
          min="0"
          step="0.01"
          value={currentAmount}
          onChange={(e) => setCurrentAmount(e.target.value)}
          className="form-input tabular-nums"
          style={inputStyle}
          placeholder="0.00"
        />
      </FormField>

      <FormField label="Deadline" htmlFor="goal-deadline">
        <input
          id="goal-deadline"
          type="date"
          value={deadline}
          onChange={(e) => setDeadline(e.target.value)}
          className="form-input"
          style={inputStyle}
        />
      </FormField>

      <div className="mt-2 flex gap-3">
        <button
          type="submit"
          disabled={isSubmitting}
          className="btn-primary flex-1 rounded-md px-4 py-2 text-sm font-medium transition-all disabled:opacity-50"
          style={{
            backgroundColor: "var(--color-accent)",
            color: "var(--color-background)",
            boxShadow: "var(--glow-accent)",
          }}
        >
          {isEditing ? "Save Changes" : "Add Goal"}
        </button>
        <button
          type="button"
          onClick={onCancel}
          className="rounded-md px-4 py-2 text-sm font-medium transition-colors"
          style={{ color: "var(--color-text-secondary)" }}
        >
          Cancel
        </button>
      </div>
    </form>
  );
}

function FormField({
  label,
  htmlFor,
  children,
}: {
  label: string;
  htmlFor: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <label
        htmlFor={htmlFor}
        className="mb-1 block text-sm font-medium"
        style={{ color: "var(--color-text-secondary)" }}
      >
        {label}
      </label>
      {children}
    </div>
  );
}
