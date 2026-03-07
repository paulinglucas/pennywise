import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { useCurrentUser } from "@/hooks/useAuth";
import AppShell from "@/components/layout/AppShell";
import Login from "@/pages/Login";
import Dashboard from "@/pages/Dashboard";
import Transactions from "@/pages/Transactions";
import Assets from "@/pages/Assets";
import Goals from "@/pages/Goals";
import Projections from "@/pages/Projections";

function RequireAuth({ children }: { children: React.ReactNode }) {
  const { data: user, isLoading, isError } = useCurrentUser();

  if (isLoading) {
    return (
      <div
        className="flex h-screen items-center justify-center"
        style={{ backgroundColor: "var(--color-background)" }}
      >
        <div className="text-sm" style={{ color: "var(--color-text-secondary)" }}>
          Loading...
        </div>
      </div>
    );
  }

  if (isError || !user) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

export default function App() {
  return (
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
          <Route path="goals" element={<Goals />} />
          <Route path="projections" element={<Projections />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
