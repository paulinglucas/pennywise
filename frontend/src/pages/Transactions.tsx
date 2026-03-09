import { useState } from "react";
import { Plus, Upload, Split } from "lucide-react";
import {
  useTransactions,
  useCreateTransaction,
  useUpdateTransaction,
  useDeleteTransaction,
  type TransactionFilters as Filters,
} from "@/hooks/useTransactions";
import { useAccounts } from "@/hooks/useAccounts";
import { useCreateTransactionGroup } from "@/hooks/useTransactionGroups";
import type {
  TransactionResponse,
  CreateTransactionRequest,
  CreateTransactionGroupRequest,
} from "@/api/client";
import TransactionFilters from "@/components/transactions/TransactionFilters";
import TransactionList from "@/components/transactions/TransactionList";
import TransactionForm from "@/components/transactions/TransactionForm";
import TransactionGroupForm from "@/components/transactions/TransactionGroupForm";
import CsvImport from "@/components/transactions/CsvImport";
import Modal from "@/components/shared/Modal";
import { Skeleton } from "@/components/shared/Skeleton";

export default function Transactions() {
  const [filters, setFilters] = useState<Filters>({ page: 1, per_page: 25 });
  const [showAddModal, setShowAddModal] = useState(false);
  const [editingTransaction, setEditingTransaction] = useState<TransactionResponse | null>(null);
  const [showImportModal, setShowImportModal] = useState(false);
  const [showGroupModal, setShowGroupModal] = useState(false);

  const transactions = useTransactions(filters);
  const accounts = useAccounts();
  const createMutation = useCreateTransaction();
  const updateMutation = useUpdateTransaction();
  const deleteMutation = useDeleteTransaction();
  const createGroupMutation = useCreateTransactionGroup();

  const accountList = accounts.data?.data ?? [];

  function handleCreate(data: CreateTransactionRequest) {
    createMutation.mutate(data, {
      onSuccess: () => setShowAddModal(false),
    });
  }

  function handleUpdate(data: CreateTransactionRequest) {
    if (!editingTransaction) return;
    updateMutation.mutate(
      { id: editingTransaction.id, body: data },
      { onSuccess: () => setEditingTransaction(null) },
    );
  }

  function handleDelete(id: string) {
    deleteMutation.mutate(id);
  }

  function handleCreateGroup(data: CreateTransactionGroupRequest) {
    createGroupMutation.mutate(data, {
      onSuccess: () => setShowGroupModal(false),
    });
  }

  function handlePageChange(page: number) {
    setFilters((prev) => ({ ...prev, page }));
  }

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        onAdd={() => setShowAddModal(true)}
        onImport={() => setShowImportModal(true)}
        onSplit={() => setShowGroupModal(true)}
      />

      <TransactionFilters filters={filters} accounts={accountList} onFiltersChange={setFilters} />

      <TransactionContent
        transactions={transactions}
        onEdit={setEditingTransaction}
        onDelete={handleDelete}
        onPageChange={handlePageChange}
      />

      <Modal isOpen={showAddModal} onClose={() => setShowAddModal(false)} title="Add Transaction">
        <TransactionForm
          accounts={accountList}
          onSubmit={handleCreate}
          onCancel={() => setShowAddModal(false)}
          isSubmitting={createMutation.isPending}
        />
      </Modal>

      <Modal
        isOpen={editingTransaction !== null}
        onClose={() => setEditingTransaction(null)}
        title="Edit Transaction"
      >
        {editingTransaction && (
          <TransactionForm
            accounts={accountList}
            transaction={editingTransaction}
            onSubmit={handleUpdate}
            onCancel={() => setEditingTransaction(null)}
            isSubmitting={updateMutation.isPending}
          />
        )}
      </Modal>

      <Modal
        isOpen={showImportModal}
        onClose={() => setShowImportModal(false)}
        title="Import Transactions"
      >
        <CsvImport
          accounts={accountList}
          onImported={() => setShowImportModal(false)}
          onClose={() => setShowImportModal(false)}
        />
      </Modal>

      <Modal
        isOpen={showGroupModal}
        onClose={() => setShowGroupModal(false)}
        title="Split Transaction"
      >
        <TransactionGroupForm
          accounts={accountList}
          onSubmit={handleCreateGroup}
          onCancel={() => setShowGroupModal(false)}
          isSubmitting={createGroupMutation.isPending}
        />
      </Modal>
    </div>
  );
}

function PageHeader({
  onAdd,
  onImport,
  onSplit,
}: {
  onAdd: () => void;
  onImport: () => void;
  onSplit: () => void;
}) {
  return (
    <div className="flex items-center justify-between">
      <h1 className="text-xl font-semibold" style={{ color: "var(--color-text-primary)" }}>
        Transactions
      </h1>
      <div className="flex gap-2">
        <button
          onClick={onImport}
          className="btn-icon flex items-center gap-1.5 rounded-md px-3 py-2 text-sm transition-all"
          style={{
            color: "var(--color-text-secondary)",
            border: "1px solid var(--color-border)",
          }}
        >
          <Upload size={14} />
          Import CSV
        </button>
        <button
          onClick={onSplit}
          className="btn-icon flex items-center gap-1.5 rounded-md px-3 py-2 text-sm transition-all"
          style={{
            color: "var(--color-text-secondary)",
            border: "1px solid var(--color-border)",
          }}
        >
          <Split size={14} />
          Split Transaction
        </button>
        <button
          onClick={onAdd}
          className="btn-primary flex items-center gap-1.5 rounded-md px-3 py-2 text-sm font-medium transition-all"
          style={{
            backgroundColor: "var(--color-accent)",
            color: "var(--color-background)",
          }}
        >
          <Plus size={14} />
          Add Transaction
        </button>
      </div>
    </div>
  );
}

function TransactionContent({
  transactions,
  onEdit,
  onDelete,
  onPageChange,
}: {
  transactions: ReturnType<typeof useTransactions>;
  onEdit: (txn: TransactionResponse) => void;
  onDelete: (id: string) => void;
  onPageChange: (page: number) => void;
}) {
  if (transactions.isError) {
    return (
      <div className="py-12 text-center">
        <p className="text-sm" style={{ color: "var(--color-negative)" }}>
          Failed to load transactions.
        </p>
      </div>
    );
  }

  if (!transactions.data) {
    return <TransactionsSkeleton />;
  }

  if (transactions.data.data.length === 0) {
    return (
      <div className="py-12 text-center">
        <p className="text-sm" style={{ color: "var(--color-text-secondary)" }}>
          No transactions found.
        </p>
      </div>
    );
  }

  return (
    <TransactionList
      transactions={transactions.data.data}
      pagination={transactions.data.pagination}
      onEdit={onEdit}
      onDelete={onDelete}
      onPageChange={onPageChange}
    />
  );
}

function TransactionsSkeleton() {
  return (
    <div className="flex flex-col gap-3" data-testid="transactions-skeleton">
      {Array.from({ length: 5 }, (_, idx) => (
        <Skeleton key={idx} className="h-16 w-full" />
      ))}
    </div>
  );
}
