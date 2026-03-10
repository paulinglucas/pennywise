import { useState } from "react";
import { Plus } from "lucide-react";
import {
  useAccounts,
  useCreateAccount,
  useUpdateAccount,
  useDeleteAccount,
} from "@/hooks/useAccounts";
import type { AccountResponse, CreateAccountRequest, UpdateAccountRequest } from "@/api/client";
import AccountForm from "@/components/accounts/AccountForm";
import Modal from "@/components/shared/Modal";
import EmptyState from "@/components/shared/EmptyState";
import ErrorState, { extractRequestId } from "@/components/shared/ErrorState";
import { SkeletonCard } from "@/components/shared/Skeleton";

type FormMode =
  | { kind: "closed" }
  | { kind: "create" }
  | { kind: "edit"; account: AccountResponse };

const accountTypeLabels: Record<string, string> = {
  checking: "Checking",
  savings: "Savings",
  hysa: "High-Yield Savings",
  credit_card: "Credit Card",
  mortgage: "Mortgage",
  brokerage: "Brokerage",
  retirement_401k: "401(k)",
  retirement_ira: "Traditional IRA",
  retirement_roth_ira: "Roth IRA",
  rollover_ira: "Rollover IRA",
  hsa: "HSA",
  crypto_wallet: "Crypto Wallet",
  safe_agreement: "SAFE Agreement",
  credit_line: "Credit Line",
  venmo: "Venmo",
  other: "Other",
};

function AccountCard({ account, onClick }: { account: AccountResponse; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="flex flex-col gap-2 rounded-lg p-4 text-left transition-all"
      style={{
        backgroundColor: "var(--color-surface)",
        border: "1px solid var(--color-border)",
      }}
    >
      <div className="flex items-center justify-between">
        <span className="text-sm font-semibold" style={{ color: "var(--color-text-primary)" }}>
          {account.name}
        </span>
        <span
          className="rounded-full px-2 py-0.5 text-xs"
          style={{
            backgroundColor: "var(--color-accent-muted)",
            color: "var(--color-accent)",
          }}
        >
          {accountTypeLabels[account.account_type] ?? account.account_type}
        </span>
      </div>
      <span className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
        {account.institution}
      </span>
    </button>
  );
}

export default function Accounts() {
  const [formMode, setFormMode] = useState<FormMode>({ kind: "closed" });
  const accounts = useAccounts();
  const createAccount = useCreateAccount();
  const updateAccount = useUpdateAccount();
  const deleteAccount = useDeleteAccount();

  if (accounts.isError) {
    return (
      <ErrorState
        message="Could not load your accounts. Please try again."
        onRetry={() => accounts.refetch()}
        requestId={extractRequestId(accounts.error)}
      />
    );
  }

  if (!accounts.data) {
    return (
      <div className="flex flex-col gap-6">
        <SkeletonCard />
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <SkeletonCard />
          <SkeletonCard />
          <SkeletonCard />
        </div>
      </div>
    );
  }

  const accountList = accounts.data.data;

  function handleCreate(data: CreateAccountRequest) {
    createAccount.mutate(data, {
      onSuccess: () => setFormMode({ kind: "closed" }),
    });
  }

  function handleUpdate(data: CreateAccountRequest) {
    if (formMode.kind !== "edit") return;
    const updateData: UpdateAccountRequest = {
      name: data.name,
      institution: data.institution,
      account_type: data.account_type,
    };
    updateAccount.mutate(
      { id: formMode.account.id, data: updateData },
      { onSuccess: () => setFormMode({ kind: "closed" }) },
    );
  }

  function handleDelete(id: string) {
    deleteAccount.mutate(id, {
      onSuccess: () => setFormMode({ kind: "closed" }),
    });
  }

  const header = (
    <div className="flex items-center justify-between">
      <h1 className="text-2xl font-semibold" style={{ color: "var(--color-text-primary)" }}>
        Accounts
      </h1>
      <button
        onClick={() => setFormMode({ kind: "create" })}
        className="btn-primary flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-all"
        style={{
          backgroundColor: "var(--color-accent)",
          color: "var(--color-background)",
          boxShadow: "var(--glow-accent)",
        }}
      >
        <Plus size={16} />
        Add Account
      </button>
    </div>
  );

  return (
    <div className="flex flex-col gap-6">
      {header}

      {accountList.length === 0 ? (
        <EmptyState
          title="No accounts yet"
          description="Add your bank accounts, credit cards, and investment accounts to track them."
        />
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {accountList.map((account) => (
            <AccountCard
              key={account.id}
              account={account}
              onClick={() => setFormMode({ kind: "edit", account })}
            />
          ))}
        </div>
      )}

      <Modal
        isOpen={formMode.kind === "create"}
        onClose={() => setFormMode({ kind: "closed" })}
        title="Add Account"
      >
        <AccountForm
          onSubmit={handleCreate}
          onCancel={() => setFormMode({ kind: "closed" })}
          isSubmitting={createAccount.isPending}
        />
      </Modal>

      <Modal
        isOpen={formMode.kind === "edit"}
        onClose={() => setFormMode({ kind: "closed" })}
        title={formMode.kind === "edit" ? formMode.account.name : "Edit Account"}
      >
        {formMode.kind === "edit" && (
          <div className="flex flex-col gap-4">
            <AccountForm
              onSubmit={handleUpdate}
              onCancel={() => setFormMode({ kind: "closed" })}
              isSubmitting={updateAccount.isPending}
              initialValues={{
                name: formMode.account.name,
                institution: formMode.account.institution,
                account_type: formMode.account.account_type,
              }}
            />
            <button
              type="button"
              onClick={() => handleDelete(formMode.account.id)}
              className="rounded-md px-4 py-2 text-sm font-medium transition-colors"
              style={{ color: "var(--color-negative)" }}
            >
              Delete Account
            </button>
          </div>
        )}
      </Modal>
    </div>
  );
}
