import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { useCurrentUser } from "@/hooks/useAuth";
import AppShell from "@/components/layout/AppShell";
import ErrorBoundary from "@/components/shared/ErrorBoundary";
import { Skeleton } from "@/components/shared/Skeleton";
import Login from "@/pages/Login";
import Dashboard from "@/pages/Dashboard";
import Transactions from "@/pages/Transactions";
import Assets from "@/pages/Assets";
import Accounts from "@/pages/Accounts";
import Goals from "@/pages/Goals";
import Projections from "@/pages/Projections";
import Settings from "@/pages/Settings";

function AuthLoadingSkeleton() {
  return (
    <div
      className="flex h-screen items-center justify-center"
      style={{ backgroundColor: "var(--color-background)" }}
    >
      <div className="flex flex-col items-center gap-4">
        <Skeleton className="h-8 w-32" />
        <Skeleton className="h-4 w-48" />
      </div>
    </div>
  );
}

function RequireAuth({ children }: { children: React.ReactNode }) {
  const { data: user, isLoading, isError } = useCurrentUser();

  if (isLoading) {
    return <AuthLoadingSkeleton />;
  }

  if (isError || !user) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

export default function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route
            element={
              <RequireAuth>
                <AppShell />
              </RequireAuth>
            }
          >
            <Route index element={<Dashboard />} />
            <Route path="transactions" element={<Transactions />} />
            <Route path="assets" element={<Assets />} />
            <Route path="accounts" element={<Accounts />} />
            <Route path="goals" element={<Goals />} />
            <Route path="projections" element={<Projections />} />
            <Route path="settings" element={<Settings />} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </ErrorBoundary>
  );
}
