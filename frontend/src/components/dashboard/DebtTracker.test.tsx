import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import DebtTracker from "./DebtTracker";

const mockDebts = [
  {
    account_id: "a1",
    name: "Mortgage",
    balance: 250000,
    monthly_payment: 1500,
    payoff_date: "2050-01-01",
    months_remaining: 300,
  },
  {
    account_id: "a2",
    name: "Car Loan",
    balance: 15000,
    monthly_payment: 400,
    payoff_date: "2028-06-01",
    months_remaining: 38,
  },
];

describe("DebtTracker", () => {
  it("renders the title", () => {
    render(<DebtTracker debts={mockDebts} />);

    expect(screen.getByText("Debt Tracker")).toBeInTheDocument();
  });

  it("renders debt names", () => {
    render(<DebtTracker debts={mockDebts} />);

    expect(screen.getByText("Mortgage")).toBeInTheDocument();
    expect(screen.getByText("Car Loan")).toBeInTheDocument();
  });

  it("renders debt balances", () => {
    render(<DebtTracker debts={mockDebts} />);

    expect(screen.getByText("$250,000.00")).toBeInTheDocument();
    expect(screen.getByText("$15,000.00")).toBeInTheDocument();
  });

  it("renders monthly payments", () => {
    render(<DebtTracker debts={mockDebts} />);

    expect(screen.getByText("$1,500.00/mo")).toBeInTheDocument();
    expect(screen.getByText("$400.00/mo")).toBeInTheDocument();
  });

  it("renders time remaining", () => {
    render(<DebtTracker debts={mockDebts} />);

    expect(screen.getByText("25 years")).toBeInTheDocument();
    expect(screen.getByText("3 years, 2 months")).toBeInTheDocument();
  });

  it("renders empty state when no debts", () => {
    render(<DebtTracker debts={[]} />);

    expect(screen.getByText("No debts tracked")).toBeInTheDocument();
  });

  it("renders progress bar when original_balance is present", () => {
    const debtsWithOriginal = [
      {
        account_id: "a1",
        name: "Car Loan",
        balance: 10000,
        monthly_payment: 400,
        original_balance: 20000,
        payoff_date: "2028-06-01",
        months_remaining: 25,
      },
    ];

    render(<DebtTracker debts={debtsWithOriginal} />);

    expect(screen.getByText("Paid off")).toBeInTheDocument();
    expect(screen.getByText("50.0%")).toBeInTheDocument();
  });

  it("does not render progress bar without original_balance", () => {
    render(<DebtTracker debts={mockDebts} />);

    expect(screen.queryByText("Paid off")).not.toBeInTheDocument();
  });
});
