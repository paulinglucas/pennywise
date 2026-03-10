import { Outlet } from "react-router-dom";
import Sidebar, { MobileTabBar } from "./Sidebar";
import TopBar from "./TopBar";

export default function AppShell() {
  return (
    <div className="flex h-screen" style={{ backgroundColor: "var(--color-background)" }}>
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[100] focus:rounded-md focus:px-4 focus:py-2 focus:text-sm focus:font-medium"
        style={{
          backgroundColor: "var(--color-accent)",
          color: "var(--color-background)",
        }}
      >
        Skip to content
      </a>
      <Sidebar />
      <div className="flex flex-1 flex-col overflow-hidden">
        <TopBar />
        <main id="main-content" className="flex-1 overflow-y-auto p-6 pb-20 md:pb-6">
          <Outlet />
        </main>
      </div>
      <MobileTabBar />
    </div>
  );
}
