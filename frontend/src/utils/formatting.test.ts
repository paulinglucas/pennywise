import { describe, it, expect } from "vitest";
import { formatCurrency, formatPercentage, formatDate, formatRelativeTime } from "./formatting";

describe("formatCurrency", () => {
  it("formats positive amounts", () => {
    expect(formatCurrency(1234.56)).toBe("$1,234.56");
  });

  it("formats negative amounts", () => {
    expect(formatCurrency(-500)).toBe("-$500.00");
  });

  it("formats zero", () => {
    expect(formatCurrency(0)).toBe("$0.00");
  });

  it("formats large amounts with commas", () => {
    expect(formatCurrency(1234567.89)).toBe("$1,234,567.89");
  });

  it("rounds to two decimal places", () => {
    expect(formatCurrency(99.999)).toBe("$100.00");
  });
});

describe("formatPercentage", () => {
  it("formats with one decimal place by default", () => {
    expect(formatPercentage(45.678)).toBe("45.7%");
  });

  it("formats with custom decimal places", () => {
    expect(formatPercentage(33.3333, 2)).toBe("33.33%");
  });

  it("formats zero", () => {
    expect(formatPercentage(0)).toBe("0.0%");
  });

  it("formats negative percentages", () => {
    expect(formatPercentage(-12.5)).toBe("-12.5%");
  });
});

describe("formatDate", () => {
  it("formats a date string", () => {
    const result = formatDate("2025-01-15");
    expect(result).toBe("Jan 15, 2025");
  });

  it("formats another date", () => {
    const result = formatDate("2024-12-31");
    expect(result).toBe("Dec 31, 2024");
  });
});

describe("formatRelativeTime", () => {
  it("formats months less than 12", () => {
    expect(formatRelativeTime(3)).toBe("3 months");
  });

  it("formats exactly 1 month", () => {
    expect(formatRelativeTime(1)).toBe("1 month");
  });

  it("formats exactly 12 months as 1 year", () => {
    expect(formatRelativeTime(12)).toBe("1 year");
  });

  it("formats years and months", () => {
    expect(formatRelativeTime(26)).toBe("2 years, 2 months");
  });

  it("formats exact years without months", () => {
    expect(formatRelativeTime(24)).toBe("2 years");
  });

  it("formats zero months", () => {
    expect(formatRelativeTime(0)).toBe("0 months");
  });
});
