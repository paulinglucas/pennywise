import { Pencil, Trash2, ChevronLeft, ChevronRight } from "lucide-react";
import { formatCurrency, formatDate } from "@/utils/formatting";
import type { TransactionResponse, PaginationMeta } from "@/api/client";

interface TransactionListProps {
  transactions: TransactionResponse[];
  pagination: PaginationMeta;
  onEdit: (transaction: TransactionResponse) => void;
  onDelete: (id: string) => void;
  onPageChange: (page: number) => void;
}

interface DateGroup {
  date: string;
  transactions: TransactionResponse[];
  dailyNet: number;
}

function groupByDate(transactions: TransactionResponse[]): DateGroup[] {
  const groups = new Map<string, TransactionResponse[]>();

  for (const txn of transactions) {
    const existing = groups.get(txn.date);
    if (existing) {
      existing.push(txn);
    } else {
      groups.set(txn.date, [txn]);
    }
  }

  return Array.from(groups.entries()).map(([date, txns]) => ({
    date,
    transactions: txns,
    dailyNet: computeDailyNet(txns),
  }));
}

function computeDailyNet(transactions: TransactionResponse[]): number {
  return transactions.reduce((sum, txn) => {
    if (txn.type === "deposit") return sum + txn.amount;
    return sum - txn.amount;
  }, 0);
}

function formatDailyNet(net: number): string {
  if (net >= 0) return `+${formatCurrency(net)}`;
  return `-${formatCurrency(Math.abs(net))}`;
}

export default function TransactionList({
  transactions,
  pagination,
  onEdit,
  onDelete,
  onPageChange,
}: TransactionListProps) {
  const groups = groupByDate(transactions);

  return (
    <div className="flex flex-col gap-4">
      {groups.map((group) => (
        <DateGroupSection key={group.date} group={group} onEdit={onEdit} onDelete={onDelete} />
      ))}

      {pagination.total_pages > 1 && (
        <Pagination pagination={pagination} onPageChange={onPageChange} />
      )}
    </div>
  );
}

function DateGroupSection({
  group,
  onEdit,
  onDelete,
}: {
  group: DateGroup;
  onEdit: (txn: TransactionResponse) => void;
  onDelete: (id: string) => void;
}) {
  const netColor = group.dailyNet >= 0 ? "var(--color-positive)" : "var(--color-negative)";

  return (
    <div>
      <div className="mb-2 flex items-center justify-between px-1">
        <span className="text-sm font-medium" style={{ color: "var(--color-text-secondary)" }}>
          {formatDate(group.date)}
        </span>
        <span className="tabular-nums text-sm font-medium" style={{ color: netColor }}>
          {formatDailyNet(group.dailyNet)}
        </span>
      </div>
      <div
        className="rounded-lg"
        style={{
          backgroundColor: "var(--color-surface)",
          border: "1px solid var(--color-border)",
        }}
      >
        {group.transactions.map((txn, idx) => (
          <TransactionRow
            key={txn.id}
            transaction={txn}
            isLast={idx === group.transactions.length - 1}
            onEdit={() => onEdit(txn)}
            onDelete={() => onDelete(txn.id)}
          />
        ))}
      </div>
    </div>
  );
}

function TransactionRow({
  transaction,
  isLast,
  onEdit,
  onDelete,
}: {
  transaction: TransactionResponse;
  isLast: boolean;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const isDeposit = transaction.type === "deposit";
  const amountColor = isDeposit ? "var(--color-positive)" : "var(--color-text-primary)";
  const amountPrefix = isDeposit ? "+" : "";

  return (
    <div
      className={`flex items-center gap-3 px-4 py-3 transition-colors ${!isLast ? "border-b" : ""}`}
      style={{ borderColor: "var(--color-border)" }}
    >
      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium" style={{ color: "var(--color-text-primary)" }}>
            {transaction.category}
          </span>
          {transaction.notes && (
            <span className="truncate text-xs" style={{ color: "var(--color-text-secondary)" }}>
              {transaction.notes}
            </span>
          )}
        </div>
        {transaction.tags.length > 0 && (
          <div className="flex gap-1">
            {transaction.tags.map((tag) => (
              <span
                key={tag}
                className="rounded px-1.5 py-0.5 text-xs"
                style={{
                  backgroundColor: "var(--color-accent-muted)",
                  color: "var(--color-accent)",
                }}
              >
                {tag}
              </span>
            ))}
          </div>
        )}
      </div>

      <span className="tabular-nums text-sm font-medium" style={{ color: amountColor }}>
        {amountPrefix}
        {formatCurrency(transaction.amount)}
      </span>

      <div className="flex gap-1">
        <button
          onClick={onEdit}
          className="btn-icon rounded p-1 transition-all"
          style={{ color: "var(--color-text-secondary)" }}
          aria-label="Edit transaction"
        >
          <Pencil size={14} />
        </button>
        <button
          onClick={onDelete}
          className="btn-icon rounded p-1 transition-all"
          style={{ color: "var(--color-text-secondary)" }}
          aria-label="Delete transaction"
        >
          <Trash2 size={14} />
        </button>
      </div>
    </div>
  );
}

function Pagination({
  pagination,
  onPageChange,
}: {
  pagination: PaginationMeta;
  onPageChange: (page: number) => void;
}) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-sm" style={{ color: "var(--color-text-secondary)" }}>
        Page {pagination.page} of {pagination.total_pages} ({pagination.total} total)
      </span>
      <div className="flex gap-1">
        <button
          onClick={() => onPageChange(pagination.page - 1)}
          disabled={pagination.page <= 1}
          className="btn-icon rounded-md p-1.5 transition-all disabled:opacity-30"
          style={{ color: "var(--color-text-secondary)" }}
          aria-label="Previous page"
        >
          <ChevronLeft size={16} />
        </button>
        <button
          onClick={() => onPageChange(pagination.page + 1)}
          disabled={pagination.page >= pagination.total_pages}
          className="btn-icon rounded-md p-1.5 transition-all disabled:opacity-30"
          style={{ color: "var(--color-text-secondary)" }}
          aria-label="Next page"
        >
          <ChevronRight size={16} />
        </button>
      </div>
    </div>
  );
}
