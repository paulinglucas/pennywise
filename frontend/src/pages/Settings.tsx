import { useState } from "react";
import { RefreshCw, Link2, Unlink, Plug, Trash2 } from "lucide-react";
import { useAccounts } from "@/hooks/useAccounts";
import {
  useSimplefinStatus,
  useSimplefinAccounts,
  useSetupSimplefin,
  useDisconnectSimplefin,
  useLinkSimplefinAccount,
  useUnlinkSimplefinAccount,
  useSyncSimplefin,
} from "@/hooks/useSimplefin";
import type { SimplefinAccount } from "@/api/client";
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

function AccountMapper({ simplefinAccounts }: { simplefinAccounts: SimplefinAccount[] }) {
  const accountsQuery = useAccounts();
  const statusQuery = useSimplefinStatus();
  const linkMutation = useLinkSimplefinAccount();
  const unlinkMutation = useUnlinkSimplefinAccount();
  const [mapping, setMapping] = useState<Record<string, string>>({});

  const accounts = accountsQuery.data?.data ?? [];
  const linkedAccounts = statusQuery.data?.linked_accounts ?? [];

  const linkedSimplefinIds = new Set(linkedAccounts.map((la) => la.simplefin_id));
  const linkedAccountIds = new Set(linkedAccounts.map((la) => la.account_id));

  function handleLink(simplefinId: string) {
    const accountId = mapping[simplefinId];
    if (!accountId) return;
    linkMutation.mutate({ accountId, simplefinId });
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
        {simplefinAccounts.map((sfinAcct) => {
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
                      {sfinAcct.institution} — linked to {linked.account_name}
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

          if (linkedSimplefinIds.has(sfinAcct.id)) return null;

          const availableAccounts = accounts.filter((a) => !linkedAccountIds.has(a.id));

          return (
            <div
              key={sfinAcct.id}
              className="flex items-center justify-between gap-3 rounded-md px-3 py-2"
              style={{ backgroundColor: "var(--color-background)" }}
            >
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
                  value={mapping[sfinAcct.id] ?? ""}
                  onChange={(e) =>
                    setMapping((prev) => ({ ...prev, [sfinAcct.id]: e.target.value }))
                  }
                  className="rounded-md px-2 py-1 text-sm"
                  style={{
                    backgroundColor: "var(--color-surface)",
                    border: "1px solid var(--color-border)",
                    color: "var(--color-text-primary)",
                  }}
                >
                  <option value="">Select account</option>
                  {availableAccounts.map((a) => (
                    <option key={a.id} value={a.id}>
                      {a.name}
                    </option>
                  ))}
                </select>
                <button
                  onClick={() => handleLink(sfinAcct.id)}
                  disabled={!mapping[sfinAcct.id] || linkMutation.isPending}
                  className="rounded-md px-3 py-1 text-sm font-medium transition-all disabled:opacity-50"
                  style={{
                    backgroundColor: "var(--color-accent-muted)",
                    color: "var(--color-accent)",
                  }}
                >
                  Link
                </button>
              </div>
            </div>
          );
        })}
      </div>
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
            </p>
          )}

          <p className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
            {status.linked_accounts.length} account
            {status.linked_accounts.length !== 1 ? "s" : ""} linked
          </p>
        </div>
      </div>

      {sfinAccountsQuery.data && (
        <AccountMapper simplefinAccounts={sfinAccountsQuery.data.accounts} />
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
