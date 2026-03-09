import { useState, type FormEvent } from "react";
import { Plus, Trash2 } from "lucide-react";
import type {
  AccountResponse,
  TransactionType,
  CreateTransactionGroupRequest,
  TransactionGroupResponse,
} from "@/api/client";
import { formatCurrency } from "@/utils/formatting";
import CategoryCombobox from "./CategoryCombobox";

interface MemberRow {
  type: TransactionType;
  category: string;
  amount: string;
  date: string;
  accountId: string;
  notes: string;
}

interface TransactionGroupFormProps {
  accounts: AccountResponse[];
  group?: TransactionGroupResponse;
  onSubmit: (data: CreateTransactionGroupRequest) => void;
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

function emptyMember(accounts: AccountResponse[]): MemberRow {
  return {
    type: "deposit",
    category: "",
    amount: "",
    date: todayString(),
    accountId: accounts[0]?.id ?? "",
    notes: "",
  };
}

export default function TransactionGroupForm({
  accounts,
  group,
  onSubmit,
  onCancel,
  isSubmitting = false,
}: TransactionGroupFormProps) {
  const [name, setName] = useState(group?.name ?? "");
  const [members, setMembers] = useState<MemberRow[]>(() => {
    if (group?.members && group.members.length > 0) {
      return group.members.map((m) => ({
        type: m.type,
        category: m.category,
        amount: m.amount.toFixed(2),
        date: m.date,
        accountId: m.account_id,
        notes: m.notes ?? "",
      }));
    }
    return [emptyMember(accounts), emptyMember(accounts)];
  });

  function updateMember(index: number, field: keyof MemberRow, value: string) {
    setMembers((prev) => prev.map((row, i) => (i === index ? { ...row, [field]: value } : row)));
  }

  function addMember() {
    setMembers((prev) => [...prev, emptyMember(accounts)]);
  }

  function removeMember(index: number) {
    if (members.length <= 2) return;
    setMembers((prev) => prev.filter((_, i) => i !== index));
  }

  const total = members.reduce((sum, m) => sum + (parseFloat(m.amount) || 0), 0);
  const validMembers = members.filter(
    (m) => m.category !== "" && m.amount !== "" && parseFloat(m.amount) > 0,
  );
  const canSubmit = name.trim() !== "" && validMembers.length >= 2;

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    onSubmit({
      name: name.trim(),
      members: validMembers.map((m) => ({
        type: m.type,
        category: m.category,
        amount: parseFloat(m.amount),
        date: m.date,
        account_id: m.accountId,
        notes: m.notes || undefined,
      })),
    });
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4">
      <div>
        <label
          htmlFor="group-name"
          className="mb-1 block text-sm font-medium"
          style={{ color: "var(--color-text-secondary)" }}
        >
          Group Name
        </label>
        <input
          id="group-name"
          type="text"
          required
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="form-input w-full"
          style={inputStyle}
          placeholder="e.g. March Paycheck"
        />
      </div>

      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium" style={{ color: "var(--color-text-secondary)" }}>
            Splits
          </span>
          <span
            className="tabular-nums text-sm font-medium"
            style={{ color: "var(--color-accent)" }}
            data-testid="group-total"
          >
            {formatCurrency(total)}
          </span>
        </div>

        {members.map((member, index) => (
          <MemberRowInput
            key={index}
            member={member}
            accounts={accounts}
            canRemove={members.length > 2}
            onUpdate={(field, value) => updateMember(index, field, value)}
            onRemove={() => removeMember(index)}
          />
        ))}

        <button
          type="button"
          onClick={addMember}
          className="btn-icon flex items-center justify-center gap-1.5 rounded-md px-3 py-2 text-sm transition-all"
          style={{
            color: "var(--color-text-secondary)",
            border: "1px dashed var(--color-border)",
          }}
        >
          <Plus size={14} />
          Add Split
        </button>
      </div>

      <div className="mt-2 flex gap-3">
        <button
          type="submit"
          disabled={!canSubmit || isSubmitting}
          className="btn-primary flex-1 rounded-md px-4 py-2 text-sm font-medium transition-all disabled:opacity-50"
          style={{
            backgroundColor: "var(--color-accent)",
            color: "var(--color-background)",
          }}
        >
          {group ? "Save Changes" : "Create Group"}
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

function MemberRowInput({
  member,
  accounts,
  canRemove,
  onUpdate,
  onRemove,
}: {
  member: MemberRow;
  accounts: AccountResponse[];
  canRemove: boolean;
  onUpdate: (field: keyof MemberRow, value: string) => void;
  onRemove: () => void;
}) {
  return (
    <div
      data-testid="member-row"
      className="flex flex-col gap-2 rounded-md p-3"
      style={{
        backgroundColor: "var(--color-background)",
        border: "1px solid var(--color-border)",
      }}
    >
      <div className="flex items-center gap-2">
        <select
          value={member.type}
          onChange={(e) => onUpdate("type", e.target.value)}
          className="form-input min-w-0 flex-1"
          style={inputStyle}
        >
          <option value="expense">Expense</option>
          <option value="deposit">Deposit</option>
          <option value="transfer">Transfer</option>
        </select>
        <CategoryCombobox value={member.category} onChange={(val) => onUpdate("category", val)} />
        <input
          type="text"
          inputMode="decimal"
          value={member.amount}
          onChange={(e) => {
            const val = e.target.value;
            if (val === "" || /^\d*\.?\d{0,2}$/.test(val)) {
              onUpdate("amount", val);
            }
          }}
          className="form-input tabular-nums min-w-0 flex-1"
          style={inputStyle}
          placeholder="Amount"
        />
        {canRemove && (
          <button
            type="button"
            onClick={onRemove}
            className="btn-icon flex-shrink-0 rounded-md p-1.5 transition-all"
            style={{ color: "var(--color-negative)" }}
          >
            <Trash2 size={14} />
          </button>
        )}
      </div>
      <div className="flex gap-2">
        <select
          value={member.accountId}
          onChange={(e) => onUpdate("accountId", e.target.value)}
          className="form-input min-w-0 flex-1"
          style={inputStyle}
        >
          {accounts.map((a) => (
            <option key={a.id} value={a.id}>
              {a.name}
            </option>
          ))}
        </select>
        <input
          type="date"
          value={member.date}
          onChange={(e) => onUpdate("date", e.target.value)}
          className="form-input min-w-0 flex-shrink-0"
          style={{ ...inputStyle, width: "150px" }}
        />
      </div>
    </div>
  );
}

const inputStyle = {
  backgroundColor: "var(--color-background)",
  borderColor: "var(--color-border)",
  color: "var(--color-text-primary)",
};
