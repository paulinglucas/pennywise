import { useState } from "react";
import { Plus, X } from "lucide-react";
import type { ProjectionParams } from "@/hooks/useProjections";
import { formatCurrency } from "@/utils/formatting";

interface ScenarioSlidersProps {
  params: ProjectionParams;
  onChange: (params: ProjectionParams) => void;
  baseMonthlySavings?: number;
}

function SliderField({
  label,
  value,
  min,
  max,
  step,
  displayValue,
  hint,
  onChange,
  children,
}: {
  label: string;
  value: number;
  min: number;
  max: number;
  step: number;
  displayValue: string;
  hint?: string;
  onChange: (value: number) => void;
  children?: React.ReactNode;
}) {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between">
        <label
          htmlFor={`slider-${label}`}
          className="text-sm font-medium"
          style={{ color: "var(--color-text-secondary)" }}
        >
          {label}
        </label>
        <span
          className="tabular-nums text-sm font-semibold"
          style={{ color: "var(--color-text-primary)" }}
        >
          {displayValue}
        </span>
      </div>
      <input
        id={`slider-${label}`}
        type="range"
        min={min}
        max={max}
        step={step}
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        className="accent-[var(--color-accent)] w-full"
        aria-label={label}
      />
      {hint && (
        <span className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
          {hint}
        </span>
      )}
      {children}
    </div>
  );
}

function EventRow({
  event,
  onRemove,
}: {
  event: { amount: number; date: string; type: "windfall" | "expense" };
  onRemove: () => void;
}) {
  const typeLabel = event.type === "windfall" ? "Windfall" : "Expense";
  const color = event.type === "windfall" ? "var(--color-positive)" : "var(--color-negative)";

  return (
    <div
      className="flex items-center justify-between rounded-md px-3 py-2 text-sm"
      style={{ backgroundColor: "var(--color-surface-hover)" }}
    >
      <div className="flex items-center gap-3">
        <span className="font-medium" style={{ color }}>
          {typeLabel}
        </span>
        <span className="tabular-nums" style={{ color: "var(--color-text-primary)" }}>
          {formatCurrency(event.amount)}
        </span>
        <span style={{ color: "var(--color-text-secondary)" }}>{event.date}</span>
      </div>
      <button
        type="button"
        onClick={onRemove}
        aria-label="Remove event"
        className="rounded p-1 transition-colors hover:bg-[var(--color-surface)]"
        style={{ color: "var(--color-text-secondary)" }}
      >
        <X size={14} />
      </button>
    </div>
  );
}

const inputStyle = {
  backgroundColor: "var(--color-background)",
  borderColor: "var(--color-border)",
  color: "var(--color-text-primary)",
};

function EventInputRow({
  type,
  amount,
  date,
  onTypeChange,
  onAmountChange,
  onDateChange,
}: {
  type: "windfall" | "expense";
  amount: string;
  date: string;
  onTypeChange: (v: "windfall" | "expense") => void;
  onAmountChange: (v: string) => void;
  onDateChange: (v: string) => void;
}) {
  return (
    <div className="flex gap-2">
      <select
        value={type}
        onChange={(e) => onTypeChange(e.target.value as "windfall" | "expense")}
        className="rounded-md border px-2 py-1.5 text-sm"
        style={inputStyle}
      >
        <option value="windfall">Windfall</option>
        <option value="expense">Expense</option>
      </select>
      <input
        type="number"
        placeholder="Amount"
        value={amount}
        onChange={(e) => onAmountChange(e.target.value)}
        min="1"
        step="1"
        className="w-28 rounded-md border px-2 py-1.5 text-sm tabular-nums"
        style={inputStyle}
      />
      <input
        type="date"
        value={date}
        onChange={(e) => onDateChange(e.target.value)}
        className="rounded-md border px-2 py-1.5 text-sm"
        style={inputStyle}
      />
    </div>
  );
}

function AddEventForm({
  onAdd,
}: {
  onAdd: (event: { amount: number; date: string; type: "windfall" | "expense" }) => void;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [amount, setAmount] = useState("");
  const [date, setDate] = useState("");
  const [type, setType] = useState<"windfall" | "expense">("windfall");

  function handleSubmit() {
    const parsed = parseFloat(amount);
    if (!parsed || parsed <= 0 || !date) return;
    onAdd({ amount: parsed, date, type });
    setAmount("");
    setDate("");
    setType("windfall");
    setIsOpen(false);
  }

  if (!isOpen) {
    return (
      <button
        type="button"
        onClick={() => setIsOpen(true)}
        className="flex items-center gap-1 text-sm font-medium transition-colors"
        style={{ color: "var(--color-accent)" }}
      >
        <Plus size={14} />
        Add Event
      </button>
    );
  }

  return (
    <div className="flex flex-col gap-2">
      <EventInputRow
        type={type}
        amount={amount}
        date={date}
        onTypeChange={setType}
        onAmountChange={setAmount}
        onDateChange={setDate}
      />
      <div className="flex gap-2">
        <button
          type="button"
          onClick={handleSubmit}
          className="rounded-md px-3 py-1 text-sm font-medium transition-all"
          style={{
            backgroundColor: "var(--color-accent)",
            color: "var(--color-background)",
          }}
        >
          Add
        </button>
        <button
          type="button"
          onClick={() => setIsOpen(false)}
          className="rounded-md px-3 py-1 text-sm font-medium transition-colors"
          style={{ color: "var(--color-text-secondary)" }}
        >
          Cancel
        </button>
      </div>
    </div>
  );
}

function ScenarioBreakdown({ returnRate }: { returnRate: number }) {
  const best = returnRate + 3;
  const worst = Math.max(returnRate - 3, 1);

  return (
    <div
      className="flex gap-3 rounded-md px-3 py-2 text-xs tabular-nums"
      style={{ backgroundColor: "var(--color-surface-hover)" }}
    >
      <span style={{ color: "#22c55e" }}>Best: {best}%</span>
      <span style={{ color: "#3b82f6" }}>Avg: {returnRate}%</span>
      <span style={{ color: "#f59e0b" }}>Worst: {worst}%</span>
    </div>
  );
}

function formatYears(years: number): string {
  return years === 1 ? "1 year" : `${years} years`;
}

function SavingsBreakdown({ base, adjustmentPct }: { base: number; adjustmentPct: number }) {
  const adjusted = base * (1 + adjustmentPct / 100);

  return (
    <div
      className="flex flex-col gap-1 rounded-md px-3 py-2 text-xs tabular-nums"
      style={{ backgroundColor: "var(--color-surface-hover)" }}
    >
      <div className="flex justify-between">
        <span style={{ color: "var(--color-text-secondary)" }}>Current monthly savings</span>
        <span style={{ color: "var(--color-text-primary)" }}>{formatCurrency(base)}</span>
      </div>
      <div className="flex justify-between">
        <span style={{ color: "var(--color-text-secondary)" }}>Projected monthly savings</span>
        <span
          style={{
            color: adjusted >= base ? "var(--color-positive)" : "var(--color-negative)",
          }}
        >
          {formatCurrency(adjusted)}
        </span>
      </div>
    </div>
  );
}

export default function ScenarioSliders({
  params,
  onChange,
  baseMonthlySavings,
}: ScenarioSlidersProps) {
  function updateParam<K extends keyof ProjectionParams>(key: K, value: ProjectionParams[K]) {
    onChange({ ...params, [key]: value });
  }

  function removeEvent(index: number) {
    const updated = params.oneTimeEvents.filter((_, i) => i !== index);
    onChange({ ...params, oneTimeEvents: updated });
  }

  function addEvent(event: { amount: number; date: string; type: "windfall" | "expense" }) {
    onChange({ ...params, oneTimeEvents: [...params.oneTimeEvents, event] });
  }

  return (
    <div
      className="flex flex-col gap-8 rounded-lg p-6"
      style={{
        backgroundColor: "var(--color-surface)",
        border: "1px solid var(--color-border)",
        boxShadow: "var(--glow-sm)",
      }}
    >
      <h3 className="text-sm font-medium" style={{ color: "var(--color-text-secondary)" }}>
        Scenario Parameters
      </h3>

      <SliderField
        label="Monthly Savings Adjustment"
        value={params.monthlySavingsAdjustment}
        min={-50}
        max={100}
        step={5}
        displayValue={`${params.monthlySavingsAdjustment >= 0 ? "+" : ""}${params.monthlySavingsAdjustment}%`}
        hint="Adjusts your current monthly cash flow. 0% uses this month's actual savings rate."
        onChange={(v) => updateParam("monthlySavingsAdjustment", v)}
      >
        {baseMonthlySavings !== undefined && (
          <SavingsBreakdown
            base={baseMonthlySavings}
            adjustmentPct={params.monthlySavingsAdjustment}
          />
        )}
      </SliderField>

      <SliderField
        label="Expected Annual Return"
        value={params.returnRate}
        min={1}
        max={15}
        step={0.5}
        displayValue={`${params.returnRate}%`}
        hint="Sets the average scenario. Best and worst cases are derived from this."
        onChange={(v) => updateParam("returnRate", v)}
      >
        <ScenarioBreakdown returnRate={params.returnRate} />
      </SliderField>

      <SliderField
        label="Years to Project"
        value={params.yearsToProject}
        min={1}
        max={40}
        step={1}
        displayValue={formatYears(params.yearsToProject)}
        onChange={(v) => updateParam("yearsToProject", v)}
      />

      <div className="flex flex-col gap-2">
        <span className="text-sm font-medium" style={{ color: "var(--color-text-secondary)" }}>
          One-Time Events
        </span>
        {params.oneTimeEvents.map((event, index) => (
          <EventRow
            key={`${event.date}-${event.type}-${index}`}
            event={event}
            onRemove={() => removeEvent(index)}
          />
        ))}
        <AddEventForm onAdd={addEvent} />
      </div>
    </div>
  );
}
