import { describe, it, expect, vi } from "vitest";
import { render } from "@testing-library/react";
import AssetSparkline, { computeChange } from "./AssetSparkline";
import type { AssetHistoryEntry } from "@/api/client";

vi.mock("recharts", async () => {
  const actual = await vi.importActual<typeof import("recharts")>("recharts");
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
      <div style={{ width: 300, height: 60 }}>{children}</div>
    ),
  };
});

const mockEntries: AssetHistoryEntry[] = [
  { id: "h1", asset_id: "a1", value: 10000, recorded_at: "2025-10-01T00:00:00Z" },
  { id: "h2", asset_id: "a1", value: 11000, recorded_at: "2025-12-01T00:00:00Z" },
  { id: "h3", asset_id: "a1", value: 12500, recorded_at: "2026-02-01T00:00:00Z" },
];

describe("AssetSparkline", () => {
  it("renders nothing when entries are empty", () => {
    const { container } = render(
      <AssetSparkline entries={[]} currentValue={10000} color="#22c55e" gradientId="test" period="6m" />,
    );
    expect(container.firstChild).toBeNull();
  });

  it("renders chart with single entry by synthesizing now point", () => {
    const { container } = render(
      <AssetSparkline
        entries={mockEntries.slice(0, 1)}
        currentValue={10500}
        color="#22c55e"
        gradientId="test"
        period="6m"
      />,
    );
    expect(container.querySelector(".recharts-wrapper")).toBeInTheDocument();
  });

  it("renders chart container when 2+ entries provided", () => {
    const { container } = render(
      <AssetSparkline
        entries={mockEntries}
        currentValue={12500}
        color="#22c55e"
        gradientId="test"
        period="6m"
      />,
    );
    expect(container.querySelector(".recharts-wrapper")).toBeInTheDocument();
  });
});

describe("computeChange", () => {
  it("returns null for fewer than 2 entries", () => {
    expect(computeChange([])).toBeNull();
    expect(computeChange(mockEntries.slice(0, 1))).toBeNull();
  });

  it("computes positive change", () => {
    expect(computeChange(mockEntries)).toBeCloseTo(25.0);
  });

  it("computes negative change", () => {
    const declining: AssetHistoryEntry[] = [
      { id: "h1", asset_id: "a1", value: 10000, recorded_at: "2025-10-01T00:00:00Z" },
      { id: "h2", asset_id: "a1", value: 8000, recorded_at: "2026-02-01T00:00:00Z" },
    ];
    expect(computeChange(declining)).toBeCloseTo(-20.0);
  });

  it("returns null when first value is zero", () => {
    const zeroStart: AssetHistoryEntry[] = [
      { id: "h1", asset_id: "a1", value: 0, recorded_at: "2025-10-01T00:00:00Z" },
      { id: "h2", asset_id: "a1", value: 1000, recorded_at: "2026-02-01T00:00:00Z" },
    ];
    expect(computeChange(zeroStart)).toBeNull();
  });
});
