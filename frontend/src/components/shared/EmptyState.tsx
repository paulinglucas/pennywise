import { Link } from "react-router-dom";

interface EmptyStateProps {
  title: string;
  description: string;
  actionLabel?: string;
  actionTo?: string;
}

export default function EmptyState({ title, description, actionLabel, actionTo }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <h2 className="mb-2 text-lg font-semibold" style={{ color: "var(--color-text-primary)" }}>
        {title}
      </h2>
      <p className="mb-6 max-w-sm text-sm" style={{ color: "var(--color-text-secondary)" }}>
        {description}
      </p>
      {actionLabel && actionTo && (
        <Link
          to={actionTo}
          className="rounded-md px-4 py-2 text-sm font-medium transition-all"
          style={{
            backgroundColor: "var(--color-accent)",
            color: "var(--color-background)",
            boxShadow: "var(--glow-accent)",
          }}
        >
          {actionLabel}
        </Link>
      )}
    </div>
  );
}
