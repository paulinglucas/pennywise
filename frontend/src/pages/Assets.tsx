import { useState } from "react";
import { Plus, Settings, ArrowLeftRight } from "lucide-react";
import {
  useAssets,
  useAssetAllocation,
  useAssetHistory,
  useCreateAsset,
  useUpdateAsset,
  useDeleteAsset,
} from "@/hooks/useAssets";
import { useAccounts } from "@/hooks/useAccounts";
import type { AssetResponse, CreateAssetRequest, UpdateAssetRequest } from "@/api/client";
import AssetOverview from "@/components/assets/AssetOverview";
import AllocationChart from "@/components/assets/AllocationChart";
import AllocationTimelapse from "@/components/assets/AllocationTimelapse";
import AssetCard from "@/components/assets/AssetCard";
import AssetForm from "@/components/assets/AssetForm";
import AssetTransactions from "@/components/assets/AssetTransactions";
import Modal from "@/components/shared/Modal";
import EmptyState from "@/components/shared/EmptyState";
import ErrorState, { extractRequestId } from "@/components/shared/ErrorState";
import { SkeletonCard, SkeletonChart } from "@/components/shared/Skeleton";

type FormMode = { kind: "closed" } | { kind: "create" } | { kind: "edit"; asset: AssetResponse };

const historyPeriods = [
  { key: "1m", label: "1M" },
  { key: "3m", label: "3M" },
  { key: "6m", label: "6M" },
  { key: "1y", label: "1Y" },
  { key: "all", label: "All" },
] as const;

function AssetsSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <SkeletonCard />
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <SkeletonChart />
        <SkeletonChart />
      </div>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <SkeletonCard />
        <SkeletonCard />
        <SkeletonCard />
      </div>
    </div>
  );
}

type DetailTab = "details" | "transactions";

function AssetDetailTabs({
  hasLinkedAccount,
  detailsContent,
  transactionsContent,
}: {
  hasLinkedAccount: boolean;
  detailsContent: React.ReactNode;
  transactionsContent: React.ReactNode;
}) {
  const [tab, setTab] = useState<DetailTab>("details");

  if (!hasLinkedAccount) {
    return <>{detailsContent}</>;
  }

  return (
    <div className="flex flex-col gap-4">
      <div
        className="flex gap-0 border-b"
        style={{ borderColor: "var(--color-border)" }}
        role="tablist"
      >
        <button
          role="tab"
          aria-selected={tab === "details"}
          onClick={() => setTab("details")}
          className="flex items-center gap-1.5 px-4 py-2 text-sm font-medium transition-colors"
          style={{
            color: tab === "details" ? "var(--color-accent)" : "var(--color-text-secondary)",
            borderBottom: tab === "details" ? "2px solid var(--color-accent)" : "2px solid transparent",
            marginBottom: -1,
          }}
        >
          <Settings size={14} />
          Details
        </button>
        <button
          role="tab"
          aria-selected={tab === "transactions"}
          onClick={() => setTab("transactions")}
          className="flex items-center gap-1.5 px-4 py-2 text-sm font-medium transition-colors"
          style={{
            color: tab === "transactions" ? "var(--color-accent)" : "var(--color-text-secondary)",
            borderBottom: tab === "transactions" ? "2px solid var(--color-accent)" : "2px solid transparent",
            marginBottom: -1,
          }}
        >
          <ArrowLeftRight size={14} />
          Transactions
        </button>
      </div>
      <div role="tabpanel">
        {tab === "details" ? detailsContent : transactionsContent}
      </div>
    </div>
  );
}

function AssetCardWithHistory({
  asset,
  portfolioTotal,
  period,
  onClick,
}: {
  asset: AssetResponse;
  portfolioTotal: number;
  period: string;
  onClick: () => void;
}) {
  const history = useAssetHistory(asset.id, period);
  const entries = history.data?.entries ?? [];

  return (
    <AssetCard
      asset={asset}
      portfolioTotal={portfolioTotal}
      onClick={onClick}
      historyEntries={entries}
      historyPeriod={period}
    />
  );
}

export default function Assets() {
  const [formMode, setFormMode] = useState<FormMode>({ kind: "closed" });
  const [allocationPeriod, setAllocationPeriod] = useState("1y");
  const [historyPeriod, setHistoryPeriod] = useState("6m");
  const assets = useAssets();
  const allocation = useAssetAllocation(allocationPeriod);
  const createAsset = useCreateAsset();
  const updateAsset = useUpdateAsset();
  const deleteAsset = useDeleteAsset();
  const accountsQuery = useAccounts();

  if (assets.isError) {
    return (
      <ErrorState
        message="Could not load your assets. Please try again."
        onRetry={() => assets.refetch()}
        requestId={extractRequestId(assets.error)}
      />
    );
  }

  if (!assets.data) {
    return <AssetsSkeleton />;
  }

  const { data: assetList, summary } = assets.data;

  function handleCreate(data: CreateAssetRequest) {
    createAsset.mutate(data, {
      onSuccess: () => setFormMode({ kind: "closed" }),
    });
  }

  function handleUpdate(data: CreateAssetRequest) {
    if (formMode.kind !== "edit") return;
    const updateData: UpdateAssetRequest = {
      name: data.name,
      asset_type: data.asset_type,
      current_value: data.current_value,
      metadata: data.metadata,
      account_id: data.account_id,
    };
    updateAsset.mutate(
      { id: formMode.asset.id, data: updateData },
      { onSuccess: () => setFormMode({ kind: "closed" }) },
    );
  }

  function handleDelete(id: string) {
    deleteAsset.mutate(id);
  }

  if (assetList.length === 0) {
    return (
      <div className="flex flex-col gap-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-semibold" style={{ color: "var(--color-text-primary)" }}>
            Assets
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
            Add Asset
          </button>
        </div>
        <EmptyState
          title="No assets yet"
          description="Add your accounts and assets to see your portfolio."
        />
        <Modal
          isOpen={formMode.kind === "create"}
          onClose={() => setFormMode({ kind: "closed" })}
          title="Add Asset"
        >
          <AssetForm
            onSubmit={handleCreate}
            onCancel={() => setFormMode({ kind: "closed" })}
            isSubmitting={createAsset.isPending}
            accounts={accountsQuery.data?.data}
          />
        </Modal>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold" style={{ color: "var(--color-text-primary)" }}>
          Assets
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
          Add Asset
        </button>
      </div>

      <AssetOverview summary={summary} />

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <AllocationChart allocation={summary.allocation} />
        {allocation.data && (
          <AllocationTimelapse
            data={allocation.data}
            period={allocationPeriod}
            onPeriodChange={setAllocationPeriod}
          />
        )}
      </div>

      <div className="flex items-center justify-between">
        <h2 className="text-sm font-medium" style={{ color: "var(--color-text-secondary)" }}>
          Holdings
        </h2>
        <div className="flex gap-1">
          {historyPeriods.map((p) => (
            <button
              key={p.key}
              onClick={() => setHistoryPeriod(p.key)}
              className="rounded-md px-3 py-1 text-xs font-medium transition-all"
              style={
                historyPeriod === p.key
                  ? {
                      backgroundColor: "var(--color-accent-muted)",
                      color: "var(--color-accent)",
                    }
                  : { color: "var(--color-text-secondary)" }
              }
            >
              {p.label}
            </button>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {assetList.map((asset) => (
          <AssetCardWithHistory
            key={asset.id}
            asset={asset}
            portfolioTotal={summary.total_value}
            period={historyPeriod}
            onClick={() => setFormMode({ kind: "edit", asset })}
          />
        ))}
      </div>

      <Modal
        isOpen={formMode.kind === "create"}
        onClose={() => setFormMode({ kind: "closed" })}
        title="Add Asset"
      >
        <AssetForm
          onSubmit={handleCreate}
          onCancel={() => setFormMode({ kind: "closed" })}
          isSubmitting={createAsset.isPending}
          accounts={accountsQuery.data?.data}
        />
      </Modal>

      <Modal
        isOpen={formMode.kind === "edit"}
        onClose={() => setFormMode({ kind: "closed" })}
        title={formMode.kind === "edit" ? formMode.asset.name : "Edit Asset"}
      >
        {formMode.kind === "edit" && (
          <AssetDetailTabs
            key={formMode.asset.id}
            hasLinkedAccount={!!formMode.asset.account_id && !!formMode.asset.linked_account}
            detailsContent={
              <div className="flex flex-col gap-4">
                <AssetForm
                  onSubmit={handleUpdate}
                  onCancel={() => setFormMode({ kind: "closed" })}
                  isSubmitting={updateAsset.isPending}
                  accounts={accountsQuery.data?.data}
                  initialValues={{
                    name: formMode.asset.name,
                    asset_type: formMode.asset.asset_type,
                    current_value: formMode.asset.current_value,
                    account_id: formMode.asset.account_id,
                    metadata: formMode.asset.metadata as Record<string, unknown>,
                  }}
                />
                <button
                  type="button"
                  onClick={() => {
                    handleDelete(formMode.asset.id);
                    setFormMode({ kind: "closed" });
                  }}
                  className="rounded-md px-4 py-2 text-sm font-medium transition-colors"
                  style={{ color: "var(--color-negative)" }}
                >
                  Delete Asset
                </button>
              </div>
            }
            transactionsContent={
              formMode.asset.account_id && formMode.asset.linked_account ? (
                <AssetTransactions
                  accountId={formMode.asset.account_id}
                  accountName={formMode.asset.linked_account.name}
                />
              ) : null
            }
          />
        )}
      </Modal>
    </div>
  );
}
