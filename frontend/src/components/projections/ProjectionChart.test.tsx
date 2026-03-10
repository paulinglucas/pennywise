import { describe, it, expect, vi } from "vitest";
import { screen } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import ProjectionChart from "./ProjectionChart";
import type { ProjectionResponse } from "@/api/client";

vi.mock("recharts", async () => {
  const actual = await vi.importActual<typeof import("recharts")>("recharts");
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
      <div data-testid="responsive-container" style={{ width: 800, height: 400 }}>
        {children}
      </div>
    ),
  };
});

const mockData: ProjectionResponse = {
  scenarios: [
    {
      scenario: "best",
      data_points: [
        { date: "2026-03-01", value: 100000 },
        { date: "2026-06-01", value: 115000 },
        { date: "2026-09-01", value: 130000 },
        { date: "2026-12-01", value: 148000 },
      ],
    },
    {
      scenario: "average",
      data_points: [
        { date: "2026-03-01", value: 100000 },
        { date: "2026-06-01", value: 110000 },
        { date: "2026-09-01", value: 120000 },
        { date: "2026-12-01", value: 132000 },
      ],
    },
    {
      scenario: "worst",
      data_points: [
        { date: "2026-03-01", value: 100000 },
        { date: "2026-06-01", value: 105000 },
        { date: "2026-09-01", value: 110000 },
        { date: "2026-12-01", value: 116000 },
      ],
    },
  ],
};

describe("ProjectionChart", () => {
  it("renders chart container", () => {
    renderWithProviders(<ProjectionChart data={mockData} />);
    expect(screen.getByTestId("responsive-container")).toBeInTheDocument();
  });

  it("renders chart title", () => {
    renderWithProviders(<ProjectionChart data={mockData} />);
    expect(screen.getByText("Net Worth Projection")).toBeInTheDocument();
  });

  it("renders scenario legend labels", () => {
    renderWithProviders(<ProjectionChart data={mockData} />);
    expect(screen.getAllByText("Best Case").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("Average").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("Worst Case").length).toBeGreaterThanOrEqual(1);
  });

  it("renders chart with large values without errors", () => {
    const bigData: ProjectionResponse = {
      scenarios: [
        {
          scenario: "best",
          data_points: [
            { date: "2026-03-01", value: 100000 },
            { date: "2036-03-01", value: 1500000 },
          ],
        },
        {
          scenario: "average",
          data_points: [
            { date: "2026-03-01", value: 100000 },
            { date: "2036-03-01", value: 1200000 },
          ],
        },
        {
          scenario: "worst",
          data_points: [
            { date: "2026-03-01", value: 100000 },
            { date: "2036-03-01", value: 800000 },
          ],
        },
      ],
    };
    renderWithProviders(<ProjectionChart data={bigData} />);
    expect(screen.getByTestId("responsive-container")).toBeInTheDocument();
    expect(screen.getByText("Net Worth Projection")).toBeInTheDocument();
  });

  it("renders empty state when no data", () => {
    const emptyData: ProjectionResponse = { scenarios: [] };
    renderWithProviders(<ProjectionChart data={emptyData} />);
    expect(screen.getByText("No projection data available")).toBeInTheDocument();
  });

  it("renders a screen-reader data table with projection scenarios", () => {
    renderWithProviders(<ProjectionChart data={mockData} />);

    const table = screen.getByRole("table", { name: "Net worth projection scenarios" });
    expect(table).toBeInTheDocument();
    expect(table.className).toContain("sr-only");

    const rows = screen.getAllByRole("row");
    expect(rows.length).toBe(5);
  });
});
