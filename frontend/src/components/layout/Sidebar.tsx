import { NavLink } from "react-router-dom";
import {
  LayoutDashboard,
  ArrowLeftRight,
  PieChart,
  Landmark,
  Target,
  TrendingUp,
  Settings,
} from "lucide-react";
import BrandLogo from "@/components/shared/BrandLogo";

const navItems = [
  { to: "/", label: "Dashboard", icon: LayoutDashboard },
  { to: "/transactions", label: "Transactions", icon: ArrowLeftRight },
  { to: "/assets", label: "Assets", icon: PieChart },
  { to: "/accounts", label: "Accounts", icon: Landmark },
  { to: "/goals", label: "Goals", icon: Target },
  { to: "/projections", label: "Projections", icon: TrendingUp },
  { to: "/settings", label: "Settings", icon: Settings },
] as const;

function navLinkClass({ isActive }: { isActive: boolean }): string {
  const base =
    "nav-link flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-all";
  if (isActive) {
    return `${base} active text-white`;
  }
  return base;
}

function navLinkStyle({ isActive }: { isActive: boolean }): React.CSSProperties {
  if (isActive) {
    return {
      backgroundColor: "var(--color-accent-muted)",
      color: "var(--color-accent)",
      borderLeft: "2px solid var(--color-accent)",
      boxShadow: "var(--glow-sm)",
    };
  }
  return { color: "var(--color-text-secondary)", borderLeft: "2px solid transparent" };
}

export default function Sidebar() {
  return (
    <nav
      className="hidden md:flex md:w-56 md:flex-col md:gap-1 md:p-4"
      style={{
        backgroundColor: "var(--color-surface)",
        borderRight: "1px solid var(--color-border)",
      }}
      aria-label="Main navigation"
    >
      <div className="mb-6 px-3">
        <BrandLogo />
      </div>
      {navItems.map((item) => (
        <NavLink
          key={item.to}
          to={item.to}
          end={item.to === "/"}
          className={navLinkClass}
          style={navLinkStyle}
        >
          <item.icon size={18} aria-hidden="true" />
          <span>{item.label}</span>
        </NavLink>
      ))}
    </nav>
  );
}

export function MobileTabBar() {
  return (
    <nav
      className="fixed bottom-0 left-0 right-0 flex md:hidden"
      style={{
        backgroundColor: "var(--color-surface)",
        borderTop: "1px solid var(--color-border)",
        boxShadow: "0 -2px 12px #22c55e10",
      }}
      aria-label="Main navigation"
    >
      {navItems.map((item) => (
        <NavLink
          key={item.to}
          to={item.to}
          end={item.to === "/"}
          className={({ isActive }) =>
            `nav-tab flex flex-1 flex-col items-center gap-1 py-2 text-xs font-medium transition-colors${isActive ? " active" : ""}`
          }
          style={({ isActive }) => ({
            color: isActive ? "var(--color-accent)" : "var(--color-text-secondary)",
          })}
        >
          <item.icon size={20} aria-hidden="true" />
          <span>{item.label}</span>
        </NavLink>
      ))}
    </nav>
  );
}
