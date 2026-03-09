import { useState, type FormEvent } from "react";
import type { TransactionResponse } from "@/api/client";
import { formatCurrency, formatDate } from "@/utils/formatting";

interface ContributeFormData {
  amount: number;
  notes?: string;
  transaction_id?: string;
}

interface ContributeFormProps {
  onSubmit: (data: ContributeFormData) => void;
  onCancel: () => void;
  isSubmitting: boolean;
  transactions?: TransactionResponse[];
}

export default function ContributeForm({
  onSubmit,
  onCancel,
  isSubmitting,
  transactions,
}: ContributeFormProps) {
  const [amount, setAmount] = useState("");
  const [notes, setNotes] = useState("");
  const [transactionId, setTransactionId] = useState("");

  function handleTransactionSelect(txnId: string) {
    setTransactionId(txnId);
    if (!txnId) return;
    const txn = transactions?.find((t) => t.id === txnId);
    if (txn && !amount) {
      setAmount(String(txn.amount));
    }
  }

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    const parsed = parseFloat(amount);
    if (!parsed || parsed <= 0) return;

    const data: ContributeFormData = { amount: parsed };
    if (notes.trim()) {
      data.notes = notes.trim();
    }
    if (transactionId) {
      data.transaction_id = transactionId;
    }
    onSubmit(data);
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4">
      {transactions && transactions.length > 0 && (
        <div className="flex flex-col gap-1">
          <label
            htmlFor="contribute-transaction"
            className="text-sm font-medium"
            style={{ color: "var(--color-text-secondary)" }}
          >
            Link to Transaction
          </label>
          <select
            id="contribute-transaction"
            value={transactionId}
            onChange={(event) => handleTransactionSelect(event.target.value)}
            className="rounded-md border px-3 py-2 text-sm"
            style={{
              backgroundColor: "var(--color-background)",
              borderColor: "var(--color-border)",
              color: "var(--color-text-primary)",
            }}
          >
            <option value="">None</option>
            {transactions.map((txn) => (
              <option key={txn.id} value={txn.id}>
                {formatDate(txn.date)} - {txn.category} - {formatCurrency(txn.amount)}
              </option>
            ))}
          </select>
        </div>
      )}
      <div className="flex flex-col gap-1">
        <label
          htmlFor="contribute-amount"
          className="text-sm font-medium"
          style={{ color: "var(--color-text-secondary)" }}
        >
          Amount
        </label>
        <input
          id="contribute-amount"
          type="number"
          min="0.01"
          step="0.01"
          value={amount}
          onChange={(event) => setAmount(event.target.value)}
          className="rounded-md border px-3 py-2 tabular-nums"
          style={{
            backgroundColor: "var(--color-background)",
            borderColor: "var(--color-border)",
            color: "var(--color-text-primary)",
          }}
          autoFocus
        />
      </div>
      <div className="flex flex-col gap-1">
        <label
          htmlFor="contribute-notes"
          className="text-sm font-medium"
          style={{ color: "var(--color-text-secondary)" }}
        >
          Notes
        </label>
        <input
          id="contribute-notes"
          type="text"
          value={notes}
          onChange={(event) => setNotes(event.target.value)}
          maxLength={500}
          className="rounded-md border px-3 py-2"
          style={{
            backgroundColor: "var(--color-background)",
            borderColor: "var(--color-border)",
            color: "var(--color-text-primary)",
          }}
        />
      </div>
      <div className="flex justify-end gap-3">
        <button
          type="button"
          onClick={onCancel}
          className="rounded-md px-4 py-2 text-sm font-medium transition-colors"
          style={{ color: "var(--color-text-secondary)" }}
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isSubmitting}
          className="rounded-md px-4 py-2 text-sm font-medium transition-all"
          style={{
            backgroundColor: "var(--color-accent)",
            color: "var(--color-background)",
            boxShadow: "var(--glow-accent)",
            opacity: isSubmitting ? 0.7 : 1,
          }}
        >
          {isSubmitting ? "Adding..." : "Add Contribution"}
        </button>
      </div>
    </form>
  );
}
