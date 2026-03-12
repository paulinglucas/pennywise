import { useState } from "react";
import { RefreshCw, Link2, Unlink, Plug, Trash2, X, RotateCcw } from "lucide-react";
import {
  useSimplefinStatus,
  useSimplefinAccounts,
  useSetupSimplefin,
  useDisconnectSimplefin,
  useLinkSimplefinAccount,
  useUnlinkSimplefinAccount,
  useDismissSimplefinAccount,
  useUndismissSimplefinAccount,
  useSyncSimplefin,
} from "@/hooks/useSimplefin";
import type { SimplefinAccount, LinkSimplefinAccountRequest } from "@/api/client";
import { formatDate } from "@/utils/formatting";

function ConnectForm() {
  const [token, setToken] = useState("");
  const setup = useSetupSimplefin();

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (token.trim()) {
      setup.mutate(token.trim(), { onSuccess: () => setToken("") });
    }
  }

  return (
    <div
      className="rounded-lg p-6"
      style={{
        backgroundColor: "var(--color-surface)",
        border: "1px solid var(--color-border)",
        boxShadow: "var(--glow-sm)",
      }}
    >
      <h2 className="mb-2 text-lg font-semibold" style={{ color: "var(--color-text-primary)" }}>
        Connect SimpleFIN
      </h2>
      <p className="mb-4 text-sm" style={{ color: "var(--color-text-secondary)" }}>
        Paste your SimpleFIN setup token to connect your bank accounts for automatic daily balance
        sync.
      </p>
      <form onSubmit={handleSubmit} className="flex gap-3">
        <input
          type="text"
          value={token}
          onChange={(e) => setToken(e.target.value)}
          placeholder="Paste your SimpleFIN setup token"
          className="flex-1 rounded-md px-3 py-2 text-sm"
          style={{
            backgroundColor: "var(--color-background)",
            border: "1px solid var(--color-border)",
            color: "var(--color-text-primary)",
          }}
        />
        <button
          type="submit"
          disabled={!token.trim() || setup.isPending}
          className="flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-all disabled:opacity-50"
          style={{
            backgroundColor: "var(--color-accent)",
            color: "var(--color-background)",
          }}
        >
          <Plug size={14} />
          {setup.isPending ? "Connecting..." : "Connect"}
        </button>
      </form>
      {setup.isError && (
        <p className="mt-2 text-sm" style={{ color: "var(--color-negative)" }}>
          Failed to connect. The token may have already been used.
        </p>
      )}
    </div>
  );
}

const ACCOUNT_TYPE_OPTIONS = [
  { value: "checking", label: "Checking" },
  { value: "savings", label: "Savings" },
  { value: "hysa", label: "HYSA" },
  { value: "credit_card", label: "Credit Card" },
  { value: "brokerage", label: "Brokerage" },
  { value: "retirement_401k", label: "401(k)" },
  { value: "retirement_ira", label: "IRA" },
  { value: "retirement_roth_ira", label: "Roth IRA" },
  { value: "rollover_ira", label: "Rollover IRA" },
  { value: "hsa", label: "HSA" },
  { value: "mortgage", label: "Mortgage" },
  { value: "credit_line", label: "Credit Line" },
  { value: "venmo", label: "Venmo" },
  { value: "crypto_wallet", label: "Crypto Wallet" },
] as const;

function defaultAccountType(balance: string): string {
  const num = parseFloat(balance);
  if (!isNaN(num) && num < 0) return "credit_card";
  return "checking";
}

interface MortgageInputProps {
  label: string;
  type: string;
  placeholder?: string;
  step?: string;
  value: string;
  onChange: (value: string) => void;
}

function MortgageInput({ label, type, placeholder, step, value, onChange }: MortgageInputProps) {
  return (
    <label className="flex flex-col gap-1">
      <span className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
        {label}
      </span>
      <input
        type={type}
        placeholder={placeholder}
        step={step}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="rounded-md px-2 py-1 text-sm tabular-nums"
        style={{
          backgroundColor: "var(--color-surface)",
          border: "1px solid var(--color-border)",
          color: "var(--color-text-primary)",
        }}
      />
    </label>
  );
}

interface AccountMapperProps {
  simplefinAccounts: SimplefinAccount[];
  dismissedIds: string[];
}

function AccountMapper({ simplefinAccounts, dismissedIds }: AccountMapperProps) {
  const statusQuery = useSimplefinStatus();
  const linkMutation = useLinkSimplefinAccount();
  const unlinkMutation = useUnlinkSimplefinAccount();
  const dismissMutation = useDismissSimplefinAccount();
  const undismissMutation = useUndismissSimplefinAccount();
  const [typeMapping, setTypeMapping] = useState<Record<string, string>>({});
  const [mortgageData, setMortgageData] = useState<
    Record<string, Partial<Pick<LinkSimplefinAccountRequest, "interest_rate" | "loan_term_months" | "purchase_price" | "purchase_date" | "down_payment_pct">>>
  >({});
  const [showDismissed, setShowDismissed] = useState(false);

  const linkedAccounts = statusQuery.data?.linked_accounts ?? [];
  const linkedSimplefinIds = new Set(linkedAccounts.map((la) => la.simplefin_id));
  const dismissedSet = new Set(dismissedIds);

  const visibleAccounts = simplefinAccounts.filter(
    (a) => !dismissedSet.has(a.id) || linkedSimplefinIds.has(a.id),
  );
  const dismissedAccounts = simplefinAccounts.filter(
    (a) => dismissedSet.has(a.id) && !linkedSimplefinIds.has(a.id),
  );

  function handleLink(sfinAcct: SimplefinAccount) {
    const accountType = typeMapping[sfinAcct.id] ?? defaultAccountType(sfinAcct.balance);
    const req: LinkSimplefinAccountRequest = {
      simplefin_id: sfinAcct.id,
      account_type: accountType,
      name: sfinAcct.name,
      institution: sfinAcct.institution,
      balance: sfinAcct.balance,
      currency: sfinAcct.currency,
    };

    if (accountType === "mortgage") {
      const md = mortgageData[sfinAcct.id];
      if (md) {
        req.interest_rate = md.interest_rate;
        req.loan_term_months = md.loan_term_months;
        req.purchase_price = md.purchase_price;
        req.purchase_date = md.purchase_date;
        req.down_payment_pct = md.down_payment_pct;
      }
    }

    linkMutation.mutate(req);
  }

  function updateMortgageField(accountId: string, field: string, value: string) {
    setMortgageData((prev) => ({
      ...prev,
      [accountId]: {
        ...prev[accountId],
        [field]: value === "" ? undefined : field === "purchase_date" ? value : parseFloat(value),
      },
    }));
  }

  return (
    <div
      className="rounded-lg p-6"
      style={{
        backgroundColor: "var(--color-surface)",
        border: "1px solid var(--color-border)",
        boxShadow: "var(--glow-sm)",
      }}
    >
      <h3 className="mb-4 text-sm font-medium" style={{ color: "var(--color-text-secondary)" }}>
        Link Accounts
      </h3>
      <div className="flex flex-col gap-3">
        {visibleAccounts.map((sfinAcct) => {
          const linked = linkedAccounts.find((la) => la.simplefin_id === sfinAcct.id);

          if (linked) {
            return (
              <div
                key={sfinAcct.id}
                className="flex items-center justify-between rounded-md px-3 py-2"
                style={{ backgroundColor: "var(--color-background)" }}
              >
                <div className="flex items-center gap-3">
                  <Link2 size={14} style={{ color: "var(--color-accent)" }} aria-hidden="true" />
                  <div>
                    <p className="text-sm" style={{ color: "var(--color-text-primary)" }}>
                      {sfinAcct.name}
                    </p>
                    <p className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
                      {sfinAcct.institution} — {linked.account_type}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <span
                    className="tabular-nums text-sm"
                    style={{ color: "var(--color-text-primary)" }}
                  >
                    ${sfinAcct.balance}
                  </span>
                  <button
                    onClick={() => unlinkMutation.mutate(linked.account_id)}
                    disabled={unlinkMutation.isPending}
                    className="rounded p-1 transition-colors"
                    style={{ color: "var(--color-text-secondary)" }}
                    title="Unlink"
                  >
                    <Unlink size={14} />
                  </button>
                </div>
              </div>
            );
          }

          const selectedType = typeMapping[sfinAcct.id] ?? defaultAccountType(sfinAcct.balance);
          const isMortgage = selectedType === "mortgage";
          const md = mortgageData[sfinAcct.id] ?? {};

          return (
            <div
              key={sfinAcct.id}
              className="flex flex-col gap-2 rounded-md px-3 py-2"
              style={{ backgroundColor: "var(--color-background)" }}
            >
              <div className="flex items-center justify-between gap-3">
                <div className="min-w-0 flex-1">
                  <p className="text-sm" style={{ color: "var(--color-text-primary)" }}>
                    {sfinAcct.name}
                  </p>
                  <p className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
                    {sfinAcct.institution} — ${sfinAcct.balance}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <select
                    value={selectedType}
                    onChange={(e) =>
                      setTypeMapping((prev) => ({ ...prev, [sfinAcct.id]: e.target.value }))
                    }
                    className="rounded-md px-2 py-1 text-sm"
                    style={{
                      backgroundColor: "var(--color-surface)",
                      border: "1px solid var(--color-border)",
                      color: "var(--color-text-primary)",
                    }}
                  >
                    {ACCOUNT_TYPE_OPTIONS.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </select>
                  <button
                    onClick={() => handleLink(sfinAcct)}
                    disabled={linkMutation.isPending}
                    className="rounded-md px-3 py-1 text-sm font-medium transition-all disabled:opacity-50"
                    style={{
                      backgroundColor: "var(--color-accent-muted)",
                      color: "var(--color-accent)",
                    }}
                  >
                    Link
                  </button>
                  <button
                    onClick={() => dismissMutation.mutate(sfinAcct.id)}
                    disabled={dismissMutation.isPending}
                    className="rounded p-1 transition-colors"
                    style={{ color: "var(--color-text-secondary)" }}
                    title="Dismiss"
                  >
                    <X size={14} />
                  </button>
                </div>
              </div>
              {isMortgage && (
                <div className="grid grid-cols-2 gap-2 pt-1 sm:grid-cols-3">
                  <MortgageInput
                    label="Purchase Price"
                    type="number"
                    placeholder="350000"
                    value={md.purchase_price?.toString() ?? ""}
                    onChange={(v) => updateMortgageField(sfinAcct.id, "purchase_price", v)}
                  />
                  <MortgageInput
                    label="Purchase Date"
                    type="date"
                    value={md.purchase_date ?? ""}
                    onChange={(v) => updateMortgageField(sfinAcct.id, "purchase_date", v)}
                  />
                  <MortgageInput
                    label="Interest Rate (%)"
                    type="number"
                    placeholder="6.5"
                    step="0.01"
                    value={md.interest_rate?.toString() ?? ""}
                    onChange={(v) => updateMortgageField(sfinAcct.id, "interest_rate", v)}
                  />
                  <MortgageInput
                    label="Loan Term (months)"
                    type="number"
                    placeholder="360"
                    value={md.loan_term_months?.toString() ?? ""}
                    onChange={(v) => updateMortgageField(sfinAcct.id, "loan_term_months", v)}
                  />
                  <MortgageInput
                    label="Down Payment (%)"
                    type="number"
                    placeholder="20"
                    step="0.1"
                    value={md.down_payment_pct?.toString() ?? ""}
                    onChange={(v) => updateMortgageField(sfinAcct.id, "down_payment_pct", v)}
                  />
                </div>
              )}
            </div>
          );
        })}
      </div>

      {dismissedAccounts.length > 0 && (
        <div className="mt-4">
          <button
            onClick={() => setShowDismissed(!showDismissed)}
            className="text-xs"
            style={{ color: "var(--color-text-secondary)" }}
          >
            {showDismissed ? "Hide" : "Show"} {dismissedAccounts.length} dismissed account
            {dismissedAccounts.length !== 1 ? "s" : ""}
          </button>
          {showDismissed && (
            <div className="mt-2 flex flex-col gap-2">
              {dismissedAccounts.map((sfinAcct) => (
                <div
                  key={sfinAcct.id}
                  className="flex items-center justify-between rounded-md px-3 py-2 opacity-60"
                  style={{ backgroundColor: "var(--color-background)" }}
                >
                  <div>
                    <p className="text-sm" style={{ color: "var(--color-text-secondary)" }}>
                      {sfinAcct.name}
                    </p>
                    <p className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
                      {sfinAcct.institution} — ${sfinAcct.balance}
                    </p>
                  </div>
                  <button
                    onClick={() => undismissMutation.mutate(sfinAcct.id)}
                    disabled={undismissMutation.isPending}
                    className="flex items-center gap-1 rounded-md px-2 py-1 text-xs transition-colors"
                    style={{ color: "var(--color-text-secondary)" }}
                    title="Restore"
                  >
                    <RotateCcw size={12} />
                    Restore
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function ConnectedStatus() {
  const statusQuery = useSimplefinStatus();
  const sfinAccountsQuery = useSimplefinAccounts(statusQuery.data?.connected ?? false);
  const syncMutation = useSyncSimplefin();
  const disconnectMutation = useDisconnectSimplefin();

  const status = statusQuery.data;
  if (!status) return null;

  return (
    <div className="flex flex-col gap-6">
      <div
        className="rounded-lg p-6"
        style={{
          backgroundColor: "var(--color-surface)",
          border: "1px solid var(--color-border)",
          boxShadow: "var(--glow-sm)",
        }}
      >
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold" style={{ color: "var(--color-text-primary)" }}>
            SimpleFIN
          </h2>
          <div className="flex items-center gap-2">
            <button
              onClick={() => syncMutation.mutate()}
              disabled={syncMutation.isPending}
              className="flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-all disabled:opacity-50"
              style={{
                backgroundColor: "var(--color-accent)",
                color: "var(--color-background)",
              }}
            >
              <RefreshCw size={14} className={syncMutation.isPending ? "animate-spin" : ""} />
              {syncMutation.isPending ? "Syncing..." : "Sync Now"}
            </button>
            <button
              onClick={() => {
                if (window.confirm("Disconnect SimpleFIN? All account links will be removed.")) {
                  disconnectMutation.mutate();
                }
              }}
              disabled={disconnectMutation.isPending}
              className="flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
              style={{ color: "var(--color-negative)" }}
            >
              <Trash2 size={14} />
              Disconnect
            </button>
          </div>
        </div>

        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-2">
            <div
              className="h-2 w-2 rounded-full"
              style={{ backgroundColor: "var(--color-accent)" }}
            />
            <span className="text-sm" style={{ color: "var(--color-text-primary)" }}>
              Connected
            </span>
          </div>

          {status.last_sync_at && (
            <p className="text-sm" style={{ color: "var(--color-text-secondary)" }}>
              Last sync: {formatDate(status.last_sync_at.split("T")[0] ?? status.last_sync_at)}
            </p>
          )}

          {status.sync_error && (
            <p className="text-sm" style={{ color: "var(--color-negative)" }}>
              Last error: {status.sync_error}
            </p>
          )}

          {syncMutation.isSuccess && syncMutation.data && (
            <p className="text-sm" style={{ color: "var(--color-accent)" }}>
              Updated {syncMutation.data.updated} account
              {syncMutation.data.updated !== 1 ? "s" : ""}
              {syncMutation.data.transactions_imported > 0 &&
                `, imported ${syncMutation.data.transactions_imported} transaction${syncMutation.data.transactions_imported !== 1 ? "s" : ""}`}
            </p>
          )}

          <p className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
            {status.linked_accounts.length} account
            {status.linked_accounts.length !== 1 ? "s" : ""} linked
          </p>
        </div>
      </div>

      {sfinAccountsQuery.data && (
        <AccountMapper
          simplefinAccounts={sfinAccountsQuery.data.accounts}
          dismissedIds={sfinAccountsQuery.data.dismissed ?? []}
        />
      )}
    </div>
  );
}

export default function Settings() {
  const statusQuery = useSimplefinStatus();

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold" style={{ color: "var(--color-text-primary)" }}>
        Settings
      </h1>

      {statusQuery.isLoading && (
        <div className="text-sm" style={{ color: "var(--color-text-secondary)" }}>
          Loading...
        </div>
      )}

      {statusQuery.data && !statusQuery.data.connected && <ConnectForm />}
      {statusQuery.data?.connected && <ConnectedStatus />}
    </div>
  );
}
