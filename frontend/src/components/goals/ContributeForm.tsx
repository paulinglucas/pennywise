import { useState, type FormEvent } from "react";
import type { TransactionResponse, AccountResponse } from "@/api/client";
import { formatCurrency, formatDate } from "@/utils/formatting";

type TransactionMode = "none" | "link" | "create";

interface ContributeFormData {
  amount: number;
  notes?: string;
  transaction_id?: string;
  create_transaction?: {
    account_id: string;
    category: string;
    date: string;
  };
}

interface ContributeFormProps {
  onSubmit: (data: ContributeFormData) => void;
  onCancel: () => void;
  isSubmitting: boolean;
  transactions?: TransactionResponse[];
  accounts?: AccountResponse[];
  categories?: string[];
  goalType?: string;
}

export type { ContributeFormData };

export default function ContributeForm({
  onSubmit,
  onCancel,
  isSubmitting,
  transactions,
  accounts,
  categories,
  goalType,
}: ContributeFormProps) {
  const [amount, setAmount] = useState("");
  const [notes, setNotes] = useState("");
  const [transactionMode, setTransactionMode] = useState<TransactionMode>("none");
  const [transactionId, setTransactionId] = useState("");
  const [accountId, setAccountId] = useState("");
  const [category, setCategory] = useState(goalType === "debt_payoff" ? "Debt Payment" : "Savings");
  const [date, setDate] = useState(new Date().toISOString().slice(0, 10));

  function handleTransactionSelect(txnId: string) {
    setTransactionId(txnId);
    if (!txnId) return;
    const txn = transactions?.find((t) => t.id === txnId);
    if (txn && !amount) {
      setAmount(String(txn.amount));
    }
  }

  function handleModeChange(mode: TransactionMode) {
    setTransactionMode(mode);
    setTransactionId("");
  }

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    const parsed = parseFloat(amount);
    if (!parsed || parsed <= 0) return;

    const data: ContributeFormData = { amount: parsed };
    if (notes.trim()) {
      data.notes = notes.trim();
    }
    if (transactionMode === "link" && transactionId) {
      data.transaction_id = transactionId;
    }
    if (transactionMode === "create" && accountId) {
      data.create_transaction = {
        account_id: accountId,
        category,
        date,
      };
    }
    onSubmit(data);
  }

  const selectStyle = {
    backgroundColor: "var(--color-background)",
    borderColor: "var(--color-border)",
    color: "var(--color-text-primary)",
  };

  const hasTransactions = transactions && transactions.length > 0;
  const hasAccounts = accounts && accounts.length > 0;

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4">
      {(hasTransactions || hasAccounts) && (
        <div className="flex flex-col gap-1">
          <span className="text-sm font-medium" style={{ color: "var(--color-text-secondary)" }}>
            Transaction
          </span>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => handleModeChange("none")}
              className="rounded-md px-3 py-1.5 text-sm font-medium transition-all"
              style={{
                backgroundColor:
                  transactionMode === "none" ? "var(--color-accent)" : "var(--color-background)",
                color:
                  transactionMode === "none"
                    ? "var(--color-background)"
                    : "var(--color-text-secondary)",
                border: "1px solid var(--color-border)",
              }}
            >
              None
            </button>
            {hasTransactions && (
              <button
                type="button"
                onClick={() => handleModeChange("link")}
                className="rounded-md px-3 py-1.5 text-sm font-medium transition-all"
                style={{
                  backgroundColor:
                    transactionMode === "link" ? "var(--color-accent)" : "var(--color-background)",
                  color:
                    transactionMode === "link"
                      ? "var(--color-background)"
                      : "var(--color-text-secondary)",
                  border: "1px solid var(--color-border)",
                }}
              >
                Link Existing
              </button>
            )}
            {hasAccounts && (
              <button
                type="button"
                onClick={() => handleModeChange("create")}
                className="rounded-md px-3 py-1.5 text-sm font-medium transition-all"
                style={{
                  backgroundColor:
                    transactionMode === "create"
                      ? "var(--color-accent)"
                      : "var(--color-background)",
                  color:
                    transactionMode === "create"
                      ? "var(--color-background)"
                      : "var(--color-text-secondary)",
                  border: "1px solid var(--color-border)",
                }}
              >
                Create New
              </button>
            )}
          </div>
        </div>
      )}
      {transactionMode === "link" && hasTransactions && (
        <div className="flex flex-col gap-1">
          <label
            htmlFor="contribute-transaction"
            className="text-sm font-medium"
            style={{ color: "var(--color-text-secondary)" }}
          >
            Select Transaction
          </label>
          <select
            id="contribute-transaction"
            value={transactionId}
            onChange={(event) => handleTransactionSelect(event.target.value)}
            className="rounded-md border px-3 py-2 text-sm"
            style={selectStyle}
          >
            <option value="">Choose...</option>
            {transactions.map((txn) => (
              <option key={txn.id} value={txn.id}>
                {formatDate(txn.date)} - {txn.category} - {formatCurrency(txn.amount)}
              </option>
            ))}
          </select>
        </div>
      )}
      {transactionMode === "create" && hasAccounts && (
        <>
          <div className="flex flex-col gap-1">
            <label
              htmlFor="contribute-account"
              className="text-sm font-medium"
              style={{ color: "var(--color-text-secondary)" }}
            >
              Account
            </label>
            <select
              id="contribute-account"
              value={accountId}
              onChange={(event) => setAccountId(event.target.value)}
              required
              className="rounded-md border px-3 py-2 text-sm"
              style={selectStyle}
            >
              <option value="">Choose account...</option>
              {accounts.map((acct) => (
                <option key={acct.id} value={acct.id}>
                  {acct.name}
                </option>
              ))}
            </select>
          </div>
          <div className="flex flex-col gap-1">
            <label
              htmlFor="contribute-category"
              className="text-sm font-medium"
              style={{ color: "var(--color-text-secondary)" }}
            >
              Category
            </label>
            {categories && categories.length > 0 ? (
              <select
                id="contribute-category"
                value={category}
                onChange={(event) => setCategory(event.target.value)}
                className="rounded-md border px-3 py-2 text-sm"
                style={selectStyle}
              >
                {categories.map((cat) => (
                  <option key={cat} value={cat}>
                    {cat}
                  </option>
                ))}
                {!categories.includes(category) && <option value={category}>{category}</option>}
              </select>
            ) : (
              <input
                id="contribute-category"
                type="text"
                value={category}
                onChange={(event) => setCategory(event.target.value)}
                required
                className="rounded-md border px-3 py-2 text-sm"
                style={selectStyle}
              />
            )}
          </div>
          <div className="flex flex-col gap-1">
            <label
              htmlFor="contribute-date"
              className="text-sm font-medium"
              style={{ color: "var(--color-text-secondary)" }}
            >
              Date
            </label>
            <input
              id="contribute-date"
              type="date"
              value={date}
              onChange={(event) => setDate(event.target.value)}
              required
              className="rounded-md border px-3 py-2 text-sm"
              style={selectStyle}
            />
          </div>
        </>
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
          style={selectStyle}
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
          style={selectStyle}
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
