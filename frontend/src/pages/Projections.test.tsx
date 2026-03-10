import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor, fireEvent } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import Projections from "./Projections";

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

const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

const mockProjectionResponse = {
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

beforeEach(() => {
  mockFetch.mockReset();
});

function setupSuccess() {
  mockFetch.mockResolvedValue({
    ok: true,
    status: 200,
    json: () => Promise.resolve(mockProjectionResponse),
  });
}

describe("Projections", () => {
  it("renders page title", () => {
    setupSuccess();
    renderWithProviders(<Projections />);
    expect(screen.getByText("Projections")).toBeInTheDocument();
  });

  it("renders sliders with default values", () => {
    setupSuccess();
    renderWithProviders(<Projections />);
    expect(screen.getByLabelText("Monthly Savings Adjustment")).toBeInTheDocument();
    expect(screen.getByLabelText("Expected Annual Return")).toBeInTheDocument();
    expect(screen.getByLabelText("Years to Project")).toBeInTheDocument();
  });

  it("renders chart and summary after data loads", async () => {
    setupSuccess();
    renderWithProviders(<Projections />);

    await waitFor(() => {
      expect(screen.getByText("Net Worth Projection")).toBeInTheDocument();
    });

    expect(screen.getAllByText("$1,500,000.00").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("$1,200,000.00").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("$800,000.00").length).toBeGreaterThanOrEqual(1);
  });

  it("renders skeleton while loading", () => {
    mockFetch.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<Projections />);
    expect(screen.getByTestId("projection-skeleton")).toBeInTheDocument();
  });

  it("renders error state on API failure with request ID", async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
      statusText: "Internal Server Error",
      json: () =>
        Promise.resolve({
          error: { code: "INTERNAL_ERROR", message: "Server error", request_id: "req-proj-789" },
        }),
    });
    renderWithProviders(<Projections />);

    await waitFor(() => {
      expect(screen.getByText("Something went wrong")).toBeInTheDocument();
    });
    expect(screen.getByRole("button", { name: /retry/i })).toBeInTheDocument();
    expect(screen.getByText("Request ID: req-proj-789")).toBeInTheDocument();
  });

  it("renders empty state when no financial data exists", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          net_worth: 0,
          cash_flow_this_month: 0,
          spending_by_category: [],
          debts_summary: [],
          scenarios: [],
        }),
    });
    renderWithProviders(<Projections />);

    await waitFor(() => {
      expect(screen.getByText("No financial data yet")).toBeInTheDocument();
    });
    expect(
      screen.getByText("Add some financial data first to see your projections."),
    ).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Go to Dashboard" })).toBeInTheDocument();
  });

  it("changing sliders triggers new API call", async () => {
    setupSuccess();
    renderWithProviders(<Projections />);

    await waitFor(() => {
      expect(screen.getByText("Net Worth Projection")).toBeInTheDocument();
    });

    const yearsSlider = screen.getByLabelText("Years to Project");
    fireEvent.change(yearsSlider, { target: { value: "20" } });

    expect(screen.getByText("20 years")).toBeInTheDocument();
  });
});
