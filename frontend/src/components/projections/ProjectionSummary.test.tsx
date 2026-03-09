import { describe, it, expect } from "vitest";
import { screen } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import ProjectionSummary from "./ProjectionSummary";
import type { ProjectionResponse } from "@/api/client";

const dataWithMillionaire: ProjectionResponse = {
  scenarios: [
    {
      scenario: "best",
      data_points: [
        { date: "2026-03-01", value: 100000 },
        { date: "2036-03-01", value: 1500000 },
      ],
      millionaire_date: "2033-06-01",
    },
    {
      scenario: "average",
      data_points: [
        { date: "2026-03-01", value: 100000 },
        { date: "2036-03-01", value: 1200000 },
      ],
      millionaire_date: "2035-01-01",
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

const dataWithoutMillionaire: ProjectionResponse = {
  scenarios: [
    {
      scenario: "best",
      data_points: [
        { date: "2026-03-01", value: 100000 },
        { date: "2036-03-01", value: 400000 },
      ],
    },
    {
      scenario: "average",
      data_points: [
        { date: "2026-03-01", value: 100000 },
        { date: "2036-03-01", value: 300000 },
      ],
    },
    {
      scenario: "worst",
      data_points: [
        { date: "2026-03-01", value: 100000 },
        { date: "2036-03-01", value: 200000 },
      ],
    },
  ],
};

describe("ProjectionSummary", () => {
  it("renders projected net worth for each scenario", () => {
    renderWithProviders(<ProjectionSummary data={dataWithMillionaire} />);
    expect(screen.getByText("Best Case")).toBeInTheDocument();
    expect(screen.getByText("Average")).toBeInTheDocument();
    expect(screen.getByText("Worst Case")).toBeInTheDocument();
  });

  it("renders final projected values", () => {
    renderWithProviders(<ProjectionSummary data={dataWithMillionaire} />);
    expect(screen.getByText("$1,500,000.00")).toBeInTheDocument();
    expect(screen.getByText("$1,200,000.00")).toBeInTheDocument();
    expect(screen.getByText("$800,000.00")).toBeInTheDocument();
  });

  it("renders millionaire date when available", () => {
    renderWithProviders(<ProjectionSummary data={dataWithMillionaire} />);
    expect(screen.getByText(/Jun 2033/)).toBeInTheDocument();
    expect(screen.getByText(/Jan 2035/)).toBeInTheDocument();
  });

  it("renders dash when millionaire date not reached", () => {
    renderWithProviders(<ProjectionSummary data={dataWithoutMillionaire} />);
    const cards = screen.getAllByText("--");
    expect(cards.length).toBeGreaterThanOrEqual(3);
  });

  it("renders empty state when no scenarios", () => {
    renderWithProviders(<ProjectionSummary data={{ scenarios: [] }} />);
    expect(screen.queryByText("Best Case")).not.toBeInTheDocument();
  });
});
