import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import NetWorthChart from "./NetWorthChart";

const mockDataPoints = [
  { date: "2025-01-01", value: 140000 },
  { date: "2025-02-01", value: 145000 },
  { date: "2025-03-01", value: 150000 },
];

describe("NetWorthChart", () => {
  it("renders the chart title", () => {
    render(<NetWorthChart dataPoints={mockDataPoints} period="1y" onPeriodChange={() => {}} />);

    expect(screen.getByText("Net Worth Over Time")).toBeInTheDocument();
  });

  it("renders period toggle buttons", () => {
    render(<NetWorthChart dataPoints={mockDataPoints} period="1y" onPeriodChange={() => {}} />);

    expect(screen.getByRole("button", { name: "1M" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "1Y" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "5Y" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "All" })).toBeInTheDocument();
  });

  it("highlights the active period button", () => {
    render(<NetWorthChart dataPoints={mockDataPoints} period="1y" onPeriodChange={() => {}} />);

    const activeButton = screen.getByRole("button", { name: "1Y" });
    expect(activeButton.style.backgroundColor).toBe("var(--color-accent-muted)");
  });

  it("calls onPeriodChange when a toggle is clicked", async () => {
    const user = userEvent.setup();
    let selectedPeriod = "";
    const handleChange = (period: string) => {
      selectedPeriod = period;
    };

    render(<NetWorthChart dataPoints={mockDataPoints} period="1y" onPeriodChange={handleChange} />);

    await user.click(screen.getByRole("button", { name: "5Y" }));
    expect(selectedPeriod).toBe("5y");
  });

  it("renders empty state when no data points", () => {
    render(<NetWorthChart dataPoints={[]} period="1y" onPeriodChange={() => {}} />);

    expect(screen.getByText("No data available for this period")).toBeInTheDocument();
  });
});
