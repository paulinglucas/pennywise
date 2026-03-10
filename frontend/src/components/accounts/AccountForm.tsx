import { useState, type FormEvent } from "react";
import type { AccountType, CreateAccountRequest } from "@/api/client";

interface AccountFormProps {
  onSubmit: (data: CreateAccountRequest) => void;
  onCancel: () => void;
  isSubmitting?: boolean;
  initialValues?: Partial<CreateAccountRequest & { is_active?: boolean }>;
}

const accountTypes: { key: AccountType; label: string }[] = [
  { key: "checking", label: "Checking" },
  { key: "savings", label: "Savings" },
  { key: "hysa", label: "High-Yield Savings" },
  { key: "credit_card", label: "Credit Card" },
  { key: "mortgage", label: "Mortgage" },
  { key: "brokerage", label: "Brokerage" },
  { key: "retirement_401k", label: "401(k)" },
  { key: "retirement_ira", label: "Traditional IRA" },
  { key: "retirement_roth_ira", label: "Roth IRA" },
  { key: "rollover_ira", label: "Rollover IRA" },
  { key: "hsa", label: "HSA" },
  { key: "crypto_wallet", label: "Crypto Wallet" },
  { key: "safe_agreement", label: "SAFE Agreement" },
  { key: "credit_line", label: "Credit Line" },
  { key: "venmo", label: "Venmo" },
  { key: "other", label: "Other" },
];

const inputStyle = {
  backgroundColor: "var(--color-background)",
  borderColor: "var(--color-border)",
  color: "var(--color-text-primary)",
};

export default function AccountForm({
  onSubmit,
  onCancel,
  isSubmitting = false,
  initialValues,
}: AccountFormProps) {
  const [name, setName] = useState(initialValues?.name ?? "");
  const [institution, setInstitution] = useState(initialValues?.institution ?? "");
  const [accountType, setAccountType] = useState<AccountType>(
    initialValues?.account_type ?? "checking",
  );

  const isEditing = !!initialValues;

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    onSubmit({ name, institution, account_type: accountType });
  }

  return (
    <form role="form" onSubmit={handleSubmit} className="flex flex-col gap-4">
      <div>
        <label
          htmlFor="account-name"
          className="mb-1 block text-sm font-medium"
          style={{ color: "var(--color-text-secondary)" }}
        >
          Name
        </label>
        <input
          id="account-name"
          type="text"
          required
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="form-input"
          style={inputStyle}
          placeholder="e.g. Chase Checking"
        />
      </div>

      <div>
        <label
          htmlFor="account-institution"
          className="mb-1 block text-sm font-medium"
          style={{ color: "var(--color-text-secondary)" }}
        >
          Institution
        </label>
        <input
          id="account-institution"
          type="text"
          required
          value={institution}
          onChange={(e) => setInstitution(e.target.value)}
          className="form-input"
          style={inputStyle}
          placeholder="e.g. Chase, Fidelity"
        />
      </div>

      <div>
        <label
          htmlFor="account-type"
          className="mb-1 block text-sm font-medium"
          style={{ color: "var(--color-text-secondary)" }}
        >
          Account Type
        </label>
        <select
          id="account-type"
          value={accountType}
          onChange={(e) => setAccountType(e.target.value as AccountType)}
          className="form-input"
          style={inputStyle}
        >
          {accountTypes.map((t) => (
            <option key={t.key} value={t.key}>
              {t.label}
            </option>
          ))}
        </select>
      </div>

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
          {isEditing ? "Save Changes" : "Add Account"}
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
