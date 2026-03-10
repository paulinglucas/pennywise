import { useState } from "react";
import { Plus } from "lucide-react";
import {
  useAssets,
  useAssetAllocation,
  useCreateAsset,
  useUpdateAsset,
  useDeleteAsset,
} from "@/hooks/useAssets";
import type { AssetResponse, CreateAssetRequest, UpdateAssetRequest } from "@/api/client";
import AssetOverview from "@/components/assets/AssetOverview";
import AllocationChart from "@/components/assets/AllocationChart";
import AllocationTimelapse from "@/components/assets/AllocationTimelapse";
import AssetCard from "@/components/assets/AssetCard";
import AssetForm from "@/components/assets/AssetForm";
import Modal from "@/components/shared/Modal";
import EmptyState from "@/components/shared/EmptyState";
import ErrorState, { extractRequestId } from "@/components/shared/ErrorState";
import { SkeletonCard, SkeletonChart } from "@/components/shared/Skeleton";

type FormMode = { kind: "closed" } | { kind: "create" } | { kind: "edit"; asset: AssetResponse };

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

export default function Assets() {
  const [formMode, setFormMode] = useState<FormMode>({ kind: "closed" });
  const [allocationPeriod, setAllocationPeriod] = useState("1y");
  const assets = useAssets();
  const allocation = useAssetAllocation(allocationPeriod);
  const createAsset = useCreateAsset();
  const updateAsset = useUpdateAsset();
  const deleteAsset = useDeleteAsset();

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

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {assetList.map((asset) => (
          <AssetCard
            key={asset.id}
            asset={asset}
            portfolioTotal={summary.total_value}
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
        />
      </Modal>

      <Modal
        isOpen={formMode.kind === "edit"}
        onClose={() => setFormMode({ kind: "closed" })}
        title="Edit Asset"
      >
        {formMode.kind === "edit" && (
          <div className="flex flex-col gap-4">
            <AssetForm
              onSubmit={handleUpdate}
              onCancel={() => setFormMode({ kind: "closed" })}
              isSubmitting={updateAsset.isPending}
              initialValues={{
                name: formMode.asset.name,
                asset_type: formMode.asset.asset_type,
                current_value: formMode.asset.current_value,
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
        )}
      </Modal>
    </div>
  );
}
