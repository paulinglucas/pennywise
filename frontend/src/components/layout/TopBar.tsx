import { LogOut } from "lucide-react";
import { useCurrentUser, useLogout } from "@/hooks/useAuth";

export default function TopBar() {
  const { data: user } = useCurrentUser();
  const logoutMutation = useLogout();

  return (
    <header
      className="flex h-14 items-center justify-end gap-4 px-6"
      style={{
        backgroundColor: "var(--color-surface)",
        borderBottom: "1px solid var(--color-border)",
        boxShadow: "0 2px 8px #22c55e08",
      }}
    >
      {user && (
        <span className="text-sm" style={{ color: "var(--color-text-secondary)" }}>
          {user.name}
        </span>
      )}
      <button
        onClick={() => logoutMutation.mutate()}
        disabled={logoutMutation.isPending}
        className="rounded-md p-2 transition-colors"
        style={{ color: "var(--color-text-secondary)" }}
        aria-label="Log out"
      >
        <LogOut size={18} />
      </button>
    </header>
  );
}
