import { ApiError } from "@/api/client";

export function extractRequestId(error: Error | null): string | undefined {
  if (error instanceof ApiError && error.requestId) {
    return error.requestId;
  }
  return undefined;
}

interface ErrorStateProps {
  title?: string;
  message: string;
  onRetry: () => void;
  requestId?: string;
}

export default function ErrorState({
  title = "Something went wrong",
  message,
  onRetry,
  requestId,
}: ErrorStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <h2 className="mb-2 text-lg font-semibold" style={{ color: "var(--color-text-primary)" }}>
        {title}
      </h2>
      <p className="mb-6 max-w-sm text-sm" style={{ color: "var(--color-text-secondary)" }}>
        {message}
      </p>
      <button
        onClick={onRetry}
        className="rounded-md px-4 py-2 text-sm font-medium transition-all"
        style={{
          backgroundColor: "var(--color-accent)",
          color: "var(--color-background)",
          boxShadow: "var(--glow-accent)",
        }}
      >
        Retry
      </button>
      {requestId && (
        <p className="mt-4 text-xs" style={{ color: "var(--color-text-muted)" }}>
          Request ID: {requestId}
        </p>
      )}
    </div>
  );
}
