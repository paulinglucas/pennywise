import { useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useTransactions } from "@/hooks/useTransactions";
import { formatCurrency, formatDate } from "@/utils/formatting";
import type { TransactionResponse } from "@/api/client";

interface AssetTransactionsProps {
  accountId: string;
  accountName: string;
}

function TransactionRow({ transaction }: { transaction: TransactionResponse }) {
  const isDeposit = transaction.type === "deposit";
  const amountColor = isDeposit ? "var(--color-positive)" : "var(--color-text-primary)";
  const amountPrefix = isDeposit ? "+" : "-";

  return (
    <div
      className="flex items-center justify-between border-b px-3 py-2 last:border-b-0"
      style={{ borderColor: "var(--color-border)" }}
    >
      <div className="flex min-w-0 flex-1 flex-col">
        <span className="truncate text-sm" style={{ color: "var(--color-text-primary)" }}>
          {transaction.category}
        </span>
        <span className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
          {formatDate(transaction.date)}
        </span>
      </div>
      <span className="tabular-nums text-sm font-medium" style={{ color: amountColor }}>
        {amountPrefix}
        {formatCurrency(transaction.amount)}
      </span>
    </div>
  );
}

export default function AssetTransactions({ accountId, accountName }: AssetTransactionsProps) {
  const [page, setPage] = useState(1);
  const query = useTransactions({ account_id: accountId, page, per_page: 10 });

  if (query.isLoading) {
    return (
      <div className="py-4 text-center text-sm" style={{ color: "var(--color-text-secondary)" }}>
        Loading transactions...
      </div>
    );
  }

  if (query.isError || !query.data) {
    return (
      <div className="py-4 text-center text-sm" style={{ color: "var(--color-text-secondary)" }}>
        Could not load transactions
      </div>
    );
  }

  const { data: transactions, pagination } = query.data;

  if (transactions.length === 0) {
    return (
      <div className="py-4 text-center text-sm" style={{ color: "var(--color-text-secondary)" }}>
        No transactions for {accountName}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-2">
      <h3 className="text-sm font-medium" style={{ color: "var(--color-text-secondary)" }}>
        Recent Transactions
      </h3>
      <div
        className="max-h-64 overflow-y-auto rounded-md"
        style={{
          backgroundColor: "var(--color-background)",
          border: "1px solid var(--color-border)",
        }}
      >
        {transactions.map((txn) => (
          <TransactionRow key={txn.id} transaction={txn} />
        ))}
      </div>
      {pagination.total_pages > 1 && (
        <div className="flex items-center justify-between">
          <span className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
            {pagination.total} transactions
          </span>
          <div className="flex gap-1">
            <button
              onClick={() => setPage((p) => p - 1)}
              disabled={page <= 1}
              className="btn-icon rounded p-1 transition-all disabled:opacity-30"
              style={{ color: "var(--color-text-secondary)" }}
              aria-label="Previous page"
            >
              <ChevronLeft size={14} />
            </button>
            <span className="text-xs leading-6" style={{ color: "var(--color-text-secondary)" }}>
              {page}/{pagination.total_pages}
            </span>
            <button
              onClick={() => setPage((p) => p + 1)}
              disabled={page >= pagination.total_pages}
              className="btn-icon rounded p-1 transition-all disabled:opacity-30"
              style={{ color: "var(--color-text-secondary)" }}
              aria-label="Next page"
            >
              <ChevronRight size={14} />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
