import { useState, useCallback, useMemo } from "react";
import { ChevronRight, Check } from "lucide-react";
import { formatCurrency, formatDate } from "@/utils/formatting";
import { useTransactions, useBulkCategorize } from "@/hooks/useTransactions";
import { useCategories } from "@/hooks/useCategories";
import type { TransactionResponse } from "@/api/client";
import Modal from "@/components/shared/Modal";

interface BulkCategorizeModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function BulkCategorizeModal({ isOpen, onClose }: BulkCategorizeModalProps) {
  const transactions = useTransactions({
    category: "uncategorized",
    per_page: 200,
    page: 1,
  });

  const uncategorized = useMemo(
    () => transactions.data?.data ?? [],
    [transactions.data],
  );

  if (!isOpen) return null;

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Categorize Transactions">
      <CategorizerContent transactions={uncategorized} onDone={onClose} />
    </Modal>
  );
}

interface CategorizerContentProps {
  transactions: TransactionResponse[];
  onDone: () => void;
}

function CategorizerContent({ transactions, onDone }: CategorizerContentProps) {
  const { data: categories = [] } = useCategories();
  const bulkCategorize = useBulkCategorize();
  const [currentIndex, setCurrentIndex] = useState(0);
  const [pendingUpdates, setPendingUpdates] = useState<
    { transaction_id: string; category: string }[]
  >([]);
  const [filterText, setFilterText] = useState("");
  const [savedCount, setSavedCount] = useState(0);

  const filteredCategories = useMemo(
    () =>
      categories
        .filter((cat) => cat !== "uncategorized")
        .filter((cat) => cat.toLowerCase().includes(filterText.toLowerCase())),
    [categories, filterText],
  );

  const flushUpdates = useCallback(
    (updates: { transaction_id: string; category: string }[]) => {
      if (updates.length === 0) return;
      const batch = [...updates];
      bulkCategorize.mutate(
        { updates: batch },
        {
          onSuccess: (result) => {
            setSavedCount((prev) => prev + result.updated);
          },
        },
      );
    },
    [bulkCategorize],
  );

  if (transactions.length === 0) {
    return (
      <div className="flex flex-col items-center gap-3 py-8">
        <p style={{ color: "var(--color-text-primary)" }}>
          No uncategorized transactions remaining.
        </p>
        {savedCount > 0 && (
          <p className="text-sm" style={{ color: "var(--color-positive)" }}>
            {savedCount} transactions categorized this session
          </p>
        )}
        <button
          onClick={onDone}
          className="btn-primary mt-2 rounded-md px-4 py-2 text-sm font-medium"
          style={{
            backgroundColor: "var(--color-accent)",
            color: "var(--color-background)",
          }}
        >
          Done
        </button>
      </div>
    );
  }

  const currentTransaction = transactions[currentIndex];
  if (!currentTransaction) {
    flushUpdates(pendingUpdates);
    return (
      <div className="flex flex-col items-center gap-3 py-8">
        <p style={{ color: "var(--color-text-primary)" }}>
          All done! {pendingUpdates.length + savedCount} transactions categorized.
        </p>
        <button
          onClick={onDone}
          className="btn-primary mt-2 rounded-md px-4 py-2 text-sm font-medium"
          style={{
            backgroundColor: "var(--color-accent)",
            color: "var(--color-background)",
          }}
        >
          Done
        </button>
      </div>
    );
  }

  function selectCategory(category: string) {
    if (!currentTransaction) return;
    const newUpdates = [
      ...pendingUpdates,
      { transaction_id: currentTransaction.id, category },
    ];

    if (newUpdates.length >= 20) {
      flushUpdates(newUpdates);
      setPendingUpdates([]);
    } else {
      setPendingUpdates(newUpdates);
    }

    setFilterText("");
    setCurrentIndex((prev) => prev + 1);
  }

  function skipTransaction() {
    setFilterText("");
    setCurrentIndex((prev) => prev + 1);
  }

  const remaining = transactions.length - currentIndex;
  const isDeposit = currentTransaction.type === "deposit";

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <span className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
          {remaining} remaining
        </span>
        {(savedCount > 0 || pendingUpdates.length > 0) && (
          <span className="text-xs" style={{ color: "var(--color-positive)" }}>
            <Check size={12} className="mr-1 inline" />
            {savedCount + pendingUpdates.length} categorized
          </span>
        )}
      </div>

      <div
        className="rounded-lg p-4"
        style={{
          backgroundColor: "var(--color-background)",
          border: "1px solid var(--color-border)",
        }}
      >
        <div className="flex items-center justify-between">
          <div className="flex min-w-0 flex-1 flex-col gap-1">
            <span
              className="text-sm font-medium"
              style={{ color: "var(--color-text-primary)" }}
            >
              {currentTransaction.notes || "No description"}
            </span>
            <span className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
              {formatDate(currentTransaction.date)}
            </span>
          </div>
          <span
            className="tabular-nums text-sm font-semibold"
            style={{
              color: isDeposit ? "var(--color-positive)" : "var(--color-text-primary)",
            }}
          >
            {isDeposit ? "+" : ""}
            {formatCurrency(currentTransaction.amount)}
          </span>
        </div>
      </div>

      <input
        type="text"
        value={filterText}
        onChange={(e) => setFilterText(e.target.value)}
        placeholder="Filter categories..."
        autoFocus
        className="form-input w-full text-sm"
        style={{
          backgroundColor: "var(--color-background)",
          borderColor: "var(--color-border)",
          color: "var(--color-text-primary)",
        }}
      />

      <div
        className="max-h-48 overflow-y-auto rounded-md"
        style={{ border: "1px solid var(--color-border)" }}
      >
        {filteredCategories.map((category) => (
          <button
            key={category}
            onClick={() => selectCategory(category)}
            className="flex w-full items-center justify-between px-3 py-2 text-left text-sm transition-colors"
            style={{ color: "var(--color-text-primary)" }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.backgroundColor =
                "var(--color-border)";
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.backgroundColor = "transparent";
            }}
          >
            <span>{category}</span>
            <ChevronRight size={14} style={{ color: "var(--color-text-secondary)" }} />
          </button>
        ))}
        {filterText && !filteredCategories.some((c) => c === filterText) && (
          <button
            onClick={() => selectCategory(filterText)}
            className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm"
            style={{ color: "var(--color-accent)" }}
          >
            Add "{filterText}"
          </button>
        )}
      </div>

      <button
        onClick={skipTransaction}
        className="self-end text-xs"
        style={{ color: "var(--color-text-secondary)" }}
      >
        Skip
      </button>
    </div>
  );
}
