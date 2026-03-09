import { describe, it, expect } from "vitest";
import { amortizationSchedule, currentEquity, safeScenarios, compoundGrowth } from "./calculations";

describe("amortizationSchedule", () => {
  it("computes monthly payments for a 30-year mortgage", () => {
    const schedule = amortizationSchedule({
      principal: 285950,
      annualRate: 6.875,
      termYears: 30,
      extraMonthly: 0,
    });

    expect(schedule.monthlyPayment).toBeCloseTo(1878.1, 0);
    expect(schedule.entries.length).toBe(360);
    expect(schedule.entries[359]!.balance).toBeCloseTo(0, 0);
  });

  it("handles extra monthly payments", () => {
    const withExtra = amortizationSchedule({
      principal: 285950,
      annualRate: 6.875,
      termYears: 30,
      extraMonthly: 200,
    });

    const without = amortizationSchedule({
      principal: 285950,
      annualRate: 6.875,
      termYears: 30,
      extraMonthly: 0,
    });

    const paidOffMonth = withExtra.entries.findIndex((e) => e.balance <= 0);
    expect(paidOffMonth).toBeLessThan(without.entries.length);
    expect(withExtra.totalInterest).toBeLessThan(without.totalInterest);
  });

  it("returns empty schedule for zero principal", () => {
    const schedule = amortizationSchedule({
      principal: 0,
      annualRate: 5,
      termYears: 30,
      extraMonthly: 0,
    });

    expect(schedule.entries.length).toBe(0);
    expect(schedule.monthlyPayment).toBe(0);
  });
});

describe("currentEquity", () => {
  it("computes equity after N months of payments", () => {
    const equity = currentEquity({
      purchasePrice: 301000,
      downPaymentPercent: 5,
      annualRate: 6.875,
      termYears: 30,
      extraMonthly: 0,
      monthsElapsed: 24,
      currentValuation: 308000,
    });

    expect(equity.downPayment).toBeCloseTo(15050, 0);
    expect(equity.currentEquity).toBeGreaterThan(equity.downPayment);
    expect(equity.loanBalance).toBeLessThan(285950);
    expect(equity.currentEquity).toBeCloseTo(308000 - equity.loanBalance, 0);
  });

  it("returns full equity when loan is paid off", () => {
    const equity = currentEquity({
      purchasePrice: 100000,
      downPaymentPercent: 20,
      annualRate: 5,
      termYears: 30,
      extraMonthly: 0,
      monthsElapsed: 360,
      currentValuation: 150000,
    });

    expect(equity.loanBalance).toBeCloseTo(0, 0);
    expect(equity.currentEquity).toBeCloseTo(150000, 0);
  });
});

describe("safeScenarios", () => {
  it("computes ownership value at different valuations", () => {
    const scenarios = safeScenarios({
      ownershipPercentage: 1.0,
      valuationCap: 3000000,
    });

    expect(scenarios).toHaveLength(5);
    expect(scenarios[0]!.valuation).toBe(5000000);
    expect(scenarios[0]!.ownershipValue).toBeCloseTo(50000, 0);

    expect(scenarios[4]!.valuation).toBe(100000000);
    expect(scenarios[4]!.ownershipValue).toBeCloseTo(1000000, 0);
  });

  it("handles fractional ownership", () => {
    const scenarios = safeScenarios({
      ownershipPercentage: 0.5,
      valuationCap: 5000000,
    });

    expect(scenarios[0]!.ownershipValue).toBeCloseTo(25000, 0);
  });

  it("applies valuation cap correctly", () => {
    const scenarios = safeScenarios({
      ownershipPercentage: 1.0,
      valuationCap: 10000000,
    });

    const belowCap = scenarios.find((s) => s.valuation === 5000000);
    expect(belowCap?.ownershipValue).toBeCloseTo(50000, 0);
    expect(belowCap?.effectiveOwnership).toBeCloseTo(1.0, 2);

    const aboveCap = scenarios.find((s) => s.valuation === 25000000);
    expect(aboveCap?.ownershipValue).toBeCloseTo(250000, 0);
    expect(aboveCap?.effectiveOwnership).toBeCloseTo(1.0, 2);
  });
});

describe("compoundGrowth", () => {
  it("computes future value with monthly contributions", () => {
    const result = compoundGrowth({
      principal: 10000,
      monthlyContribution: 500,
      annualRate: 7,
      years: 10,
    });

    expect(result.finalValue).toBeGreaterThan(10000 + 500 * 120);
    expect(result.totalContributed).toBeCloseTo(10000 + 500 * 120, 0);
    expect(result.totalGrowth).toBeCloseTo(result.finalValue - result.totalContributed, 0);
  });

  it("handles zero contributions", () => {
    const result = compoundGrowth({
      principal: 10000,
      monthlyContribution: 0,
      annualRate: 7,
      years: 10,
    });

    expect(result.finalValue).toBeGreaterThan(10000);
    expect(result.totalContributed).toBe(10000);
  });

  it("handles zero rate", () => {
    const result = compoundGrowth({
      principal: 10000,
      monthlyContribution: 100,
      annualRate: 0,
      years: 5,
    });

    expect(result.finalValue).toBeCloseTo(10000 + 100 * 60, 0);
    expect(result.totalGrowth).toBeCloseTo(0, 0);
  });
});
