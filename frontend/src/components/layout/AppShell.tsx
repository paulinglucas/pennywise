import { Outlet } from "react-router-dom";
import Sidebar, { MobileTabBar } from "./Sidebar";
import TopBar from "./TopBar";

export default function AppShell() {
  return (
    <div className="flex h-screen" style={{ backgroundColor: "var(--color-background)" }}>
      <Sidebar />
      <div className="flex flex-1 flex-col overflow-hidden">
        <TopBar />
        <main className="flex-1 overflow-y-auto p-6 pb-20 md:pb-6">
          <Outlet />
        </main>
      </div>
      <MobileTabBar />
    </div>
  );
}
