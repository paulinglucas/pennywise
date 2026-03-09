import { useState, type FormEvent } from "react";
import type {
  AccountResponse,
  TransactionResponse,
  CreateTransactionRequest,
  TransactionType,
} from "@/api/client";
import CategoryCombobox from "./CategoryCombobox";

interface TransactionFormProps {
  accounts: AccountResponse[];
  transaction?: TransactionResponse;
  onSubmit: (data: CreateTransactionRequest) => void;
  onCancel: () => void;
  isSubmitting?: boolean;
}

function todayString(): string {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export default function TransactionForm({
  accounts,
  transaction,
  onSubmit,
  onCancel,
  isSubmitting = false,
}: TransactionFormProps) {
  const [type, setType] = useState<TransactionType>(transaction?.type ?? "expense");
  const [category, setCategory] = useState(transaction?.category ?? "");
  const [amount, setAmount] = useState(transaction ? transaction.amount.toFixed(2) : "");
  const [date, setDate] = useState(transaction?.date ?? todayString());
  const [accountId, setAccountId] = useState(transaction?.account_id ?? accounts[0]?.id ?? "");
  const [notes, setNotes] = useState(transaction?.notes ?? "");
  const [tags, setTags] = useState(transaction?.tags.join(", ") ?? "");

  function handleSubmit(event: FormEvent) {
    event.preventDefault();

    const parsedTags = tags
      .split(",")
      .map((tag) => tag.trim())
      .filter((tag) => tag !== "");

    onSubmit({
      type,
      category,
      amount: parseFloat(amount),
      date,
      account_id: accountId,
      notes: notes || undefined,
      tags: parsedTags.length > 0 ? parsedTags : undefined,
    });
  }

  const isEditing = !!transaction;

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4">
      <TypeToggle activeType={type} onTypeChange={setType} />

      <FormField label="Category" htmlFor="category">
        <CategoryCombobox
          id="category"
          value={category}
          onChange={setCategory}
          placeholder="e.g. food, rent, salary"
        />
      </FormField>

      <FormField label="Amount" htmlFor="amount">
        <input
          id="amount"
          type="number"
          required
          min="0.01"
          step="0.01"
          value={amount}
          onChange={(event) => setAmount(event.target.value)}
          className="form-input tabular-nums"
          style={inputStyle}
          placeholder="0.00"
        />
      </FormField>

      <FormField label="Date" htmlFor="date">
        <input
          id="date"
          type="date"
          required
          value={date}
          onChange={(event) => setDate(event.target.value)}
          className="form-input"
          style={inputStyle}
        />
      </FormField>

      <FormField label="Account" htmlFor="account">
        <select
          id="account"
          value={accountId}
          onChange={(event) => setAccountId(event.target.value)}
          className="form-input"
          style={inputStyle}
        >
          {accounts.map((account) => (
            <option key={account.id} value={account.id}>
              {account.name} ({account.institution})
            </option>
          ))}
        </select>
      </FormField>

      <FormField label="Notes" htmlFor="notes">
        <input
          id="notes"
          type="text"
          value={notes}
          onChange={(event) => setNotes(event.target.value)}
          className="form-input"
          style={inputStyle}
          placeholder="Optional"
        />
      </FormField>

      <FormField label="Tags" htmlFor="tags">
        <input
          id="tags"
          type="text"
          value={tags}
          onChange={(event) => setTags(event.target.value)}
          className="form-input"
          style={inputStyle}
          placeholder="Comma-separated, e.g. dining, work"
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
          {isEditing ? "Save Changes" : "Add Transaction"}
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

const inputStyle = {
  backgroundColor: "var(--color-background)",
  borderColor: "var(--color-border)",
  color: "var(--color-text-primary)",
};

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

function TypeToggle({
  activeType,
  onTypeChange,
}: {
  activeType: TransactionType;
  onTypeChange: (type: TransactionType) => void;
}) {
  const types: { key: TransactionType; label: string }[] = [
    { key: "expense", label: "Expense" },
    { key: "deposit", label: "Deposit" },
    { key: "transfer", label: "Transfer" },
  ];

  return (
    <div
      className="flex gap-1 rounded-md p-1"
      style={{ backgroundColor: "var(--color-background)" }}
    >
      {types.map((item) => (
        <button
          key={item.key}
          type="button"
          onClick={() => onTypeChange(item.key)}
          className="btn-toggle flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-all"
          style={
            activeType === item.key
              ? { backgroundColor: "var(--color-accent-muted)", color: "var(--color-accent)" }
              : { color: "var(--color-text-secondary)" }
          }
        >
          {item.label}
        </button>
      ))}
    </div>
  );
}
