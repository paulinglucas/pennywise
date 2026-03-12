import { describe, it, expect, vi } from "vitest";
import { screen, fireEvent } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import ScenarioSliders from "./ScenarioSliders";
import type { ProjectionParams } from "@/hooks/useProjections";

const defaultParams: ProjectionParams = {
  monthlySavingsAdjustment: 0,
  returnRate: 7,
  yearsToProject: 10,
  oneTimeEvents: [],
};

describe("ScenarioSliders", () => {
  it("renders all three sliders", () => {
    renderWithProviders(<ScenarioSliders params={defaultParams} onChange={vi.fn()} />);
    expect(screen.getByLabelText("Monthly Savings Adjustment")).toBeInTheDocument();
    expect(screen.getByLabelText("Expected Annual Return")).toBeInTheDocument();
    expect(screen.getByLabelText("Years to Project")).toBeInTheDocument();
  });

  it("displays current slider values", () => {
    renderWithProviders(<ScenarioSliders params={defaultParams} onChange={vi.fn()} />);
    expect(screen.getByText("+0%")).toBeInTheDocument();
    expect(screen.getByText("7%")).toBeInTheDocument();
    expect(screen.getByText("10 years")).toBeInTheDocument();
  });

  it("calls onChange when savings adjustment slider changes", () => {
    const onChange = vi.fn();
    renderWithProviders(<ScenarioSliders params={defaultParams} onChange={onChange} />);
    const slider = screen.getByLabelText("Monthly Savings Adjustment");
    fireEvent.change(slider, { target: { value: "25" } });
    expect(onChange).toHaveBeenCalledWith({
      ...defaultParams,
      monthlySavingsAdjustment: 25,
    });
  });

  it("calls onChange when return rate slider changes", () => {
    const onChange = vi.fn();
    renderWithProviders(<ScenarioSliders params={defaultParams} onChange={onChange} />);
    const slider = screen.getByLabelText("Expected Annual Return");
    fireEvent.change(slider, { target: { value: "10" } });
    expect(onChange).toHaveBeenCalledWith({
      ...defaultParams,
      returnRate: 10,
    });
  });

  it("calls onChange when years slider changes", () => {
    const onChange = vi.fn();
    renderWithProviders(<ScenarioSliders params={defaultParams} onChange={onChange} />);
    const slider = screen.getByLabelText("Years to Project");
    fireEvent.change(slider, { target: { value: "20" } });
    expect(onChange).toHaveBeenCalledWith({
      ...defaultParams,
      yearsToProject: 20,
    });
  });

  it("renders hint text for sliders", () => {
    renderWithProviders(<ScenarioSliders params={defaultParams} onChange={vi.fn()} />);
    expect(screen.getByText(/Adjusts your current monthly cash flow/)).toBeInTheDocument();
    expect(screen.getByText(/Sets the average scenario/)).toBeInTheDocument();
  });

  it("renders scenario breakdown showing best, avg, worst rates", () => {
    renderWithProviders(<ScenarioSliders params={defaultParams} onChange={vi.fn()} />);
    expect(screen.getByText("Best: 10%")).toBeInTheDocument();
    expect(screen.getByText("Avg: 7%")).toBeInTheDocument();
    expect(screen.getByText("Worst: 4%")).toBeInTheDocument();
  });

  it("updates scenario breakdown when return rate changes", () => {
    const params: ProjectionParams = { ...defaultParams, returnRate: 12 };
    renderWithProviders(<ScenarioSliders params={params} onChange={vi.fn()} />);
    expect(screen.getByText("Best: 15%")).toBeInTheDocument();
    expect(screen.getByText("Avg: 12%")).toBeInTheDocument();
    expect(screen.getByText("Worst: 9%")).toBeInTheDocument();
  });

  it("floors worst case at 1%", () => {
    const params: ProjectionParams = { ...defaultParams, returnRate: 2 };
    renderWithProviders(<ScenarioSliders params={params} onChange={vi.fn()} />);
    expect(screen.getByText("Worst: 1%")).toBeInTheDocument();
  });

  it("renders savings breakdown when baseMonthlySavings is provided", () => {
    renderWithProviders(
      <ScenarioSliders params={defaultParams} onChange={vi.fn()} baseMonthlySavings={2000} />,
    );
    expect(screen.getByText("Current monthly savings")).toBeInTheDocument();
    expect(screen.getByText("Projected monthly savings")).toBeInTheDocument();
    expect(screen.getAllByText("$2,000.00")).toHaveLength(2);
  });

  it("shows adjusted savings amount based on slider percentage", () => {
    const params: ProjectionParams = { ...defaultParams, monthlySavingsAdjustment: 50 };
    renderWithProviders(
      <ScenarioSliders params={params} onChange={vi.fn()} baseMonthlySavings={2000} />,
    );
    expect(screen.getByText("$3,000.00")).toBeInTheDocument();
  });

  it("does not render savings breakdown when baseMonthlySavings is not provided", () => {
    renderWithProviders(<ScenarioSliders params={defaultParams} onChange={vi.fn()} />);
    expect(screen.queryByText("Current monthly savings")).not.toBeInTheDocument();
  });

  it("renders add event button", () => {
    renderWithProviders(<ScenarioSliders params={defaultParams} onChange={vi.fn()} />);
    expect(screen.getByText("Add Event")).toBeInTheDocument();
  });

  it("renders existing one-time events", () => {
    const params: ProjectionParams = {
      ...defaultParams,
      oneTimeEvents: [
        { amount: 50000, date: "2027-06-01", type: "windfall" },
        { amount: 10000, date: "2028-01-15", type: "expense" },
      ],
    };
    renderWithProviders(<ScenarioSliders params={params} onChange={vi.fn()} />);
    expect(screen.getByText("Windfall")).toBeInTheDocument();
    expect(screen.getByText("Expense")).toBeInTheDocument();
  });

  it("removes a one-time event when remove button is clicked", () => {
    const onChange = vi.fn();
    const params: ProjectionParams = {
      ...defaultParams,
      oneTimeEvents: [{ amount: 50000, date: "2027-06-01", type: "windfall" }],
    };
    renderWithProviders(<ScenarioSliders params={params} onChange={onChange} />);
    fireEvent.click(screen.getByLabelText("Remove event"));
    expect(onChange).toHaveBeenCalledWith({
      ...defaultParams,
      oneTimeEvents: [],
    });
  });

  it("opens add event form when Add Event is clicked", () => {
    renderWithProviders(<ScenarioSliders params={defaultParams} onChange={vi.fn()} />);
    fireEvent.click(screen.getByText("Add Event"));
    expect(screen.getByText("Add")).toBeInTheDocument();
    expect(screen.getByText("Cancel")).toBeInTheDocument();
  });

  it("closes add event form when Cancel is clicked", () => {
    renderWithProviders(<ScenarioSliders params={defaultParams} onChange={vi.fn()} />);
    fireEvent.click(screen.getByText("Add Event"));
    fireEvent.click(screen.getByText("Cancel"));
    expect(screen.getByText("Add Event")).toBeInTheDocument();
  });

  it("adds a one-time event via the form", () => {
    const onChange = vi.fn();
    renderWithProviders(<ScenarioSliders params={defaultParams} onChange={onChange} />);
    fireEvent.click(screen.getByText("Add Event"));

    const amountInput = screen.getByPlaceholderText("Amount");
    const dateInput = document.querySelector('input[type="date"]') as HTMLInputElement;

    fireEvent.change(amountInput, { target: { value: "25000" } });
    fireEvent.change(dateInput, { target: { value: "2027-06-01" } });

    fireEvent.click(screen.getByText("Add"));

    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({
        oneTimeEvents: expect.arrayContaining([
          expect.objectContaining({
            amount: 25000,
            date: "2027-06-01",
          }),
        ]),
      }),
    );
  });

  it("shows negative savings adjustment formatting", () => {
    const params: ProjectionParams = { ...defaultParams, monthlySavingsAdjustment: -20 };
    renderWithProviders(<ScenarioSliders params={params} onChange={vi.fn()} />);
    expect(screen.getByText("-20%")).toBeInTheDocument();
  });
});
