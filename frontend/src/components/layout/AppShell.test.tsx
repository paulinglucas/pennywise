import { describe, it, expect } from "vitest";
import { screen } from "@testing-library/react";
import { Routes, Route } from "react-router-dom";
import { renderWithProviders } from "@/test-utils";
import AppShell from "./AppShell";

describe("AppShell", () => {
  it("renders sidebar and content area", () => {
    renderWithProviders(
      <Routes>
        <Route element={<AppShell />}>
          <Route index element={<div>Test Content</div>} />
        </Route>
      </Routes>,
    );

    expect(screen.getAllByRole("navigation", { name: /main navigation/i }).length).toBe(2);
    expect(screen.getByText("Test Content")).toBeInTheDocument();
  });

  it("renders Pennywise branding in sidebar", () => {
    renderWithProviders(
      <Routes>
        <Route element={<AppShell />}>
          <Route index element={<div>Page</div>} />
        </Route>
      </Routes>,
    );

    expect(screen.getByText("Pennywise")).toBeInTheDocument();
  });

  it("renders all navigation links", () => {
    renderWithProviders(
      <Routes>
        <Route element={<AppShell />}>
          <Route index element={<div>Page</div>} />
        </Route>
      </Routes>,
    );

    const navLinks = ["Dashboard", "Transactions", "Assets", "Goals", "Projections"];
    for (const label of navLinks) {
      expect(screen.getAllByText(label).length).toBeGreaterThan(0);
    }
  });

  it("renders logout button", () => {
    renderWithProviders(
      <Routes>
        <Route element={<AppShell />}>
          <Route index element={<div>Page</div>} />
        </Route>
      </Routes>,
    );

    expect(screen.getByRole("button", { name: /log out/i })).toBeInTheDocument();
  });
});
