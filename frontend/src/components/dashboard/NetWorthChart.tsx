import { ResponsiveContainer, AreaChart, Area, XAxis, YAxis, Tooltip } from "recharts";
import { formatCurrency, formatDate } from "@/utils/formatting";

interface DataPoint {
  date: string;
  value: number;
}

interface NetWorthChartProps {
  dataPoints: DataPoint[];
  period: string;
  onPeriodChange: (period: string) => void;
}

const periods = [
  { key: "1m", label: "1M" },
  { key: "1y", label: "1Y" },
  { key: "5y", label: "5Y" },
  { key: "all", label: "All" },
] as const;

function PeriodToggle({
  activePeriod,
  onPeriodChange,
}: {
  activePeriod: string;
  onPeriodChange: (period: string) => void;
}) {
  return (
    <div className="flex gap-1">
      {periods.map((period) => (
        <button
          key={period.key}
          onClick={() => onPeriodChange(period.key)}
          className={`btn-toggle rounded-md px-3 py-1 text-xs font-medium transition-all${activePeriod === period.key ? " active" : ""}`}
          style={
            activePeriod === period.key
              ? { backgroundColor: "var(--color-accent-muted)", color: "var(--color-accent)" }
              : { color: "var(--color-text-secondary)" }
          }
        >
          {period.label}
        </button>
      ))}
    </div>
  );
}

function ChartTooltip({
  active,
  payload,
}: {
  active?: boolean;
  payload?: Array<{ value: number; payload: TimeDataPoint }>;
  label?: number;
}) {
  const firstEntry = payload?.[0];
  if (!active || !firstEntry) return null;

  return (
    <div
      className="rounded-md px-3 py-2 text-sm"
      style={{
        backgroundColor: "var(--color-surface)",
        border: "1px solid var(--color-accent-muted)",
        boxShadow: "var(--glow-md)",
      }}
    >
      <p className="mb-1 text-xs" style={{ color: "var(--color-text-secondary)" }}>
        {formatDate(firstEntry.payload.date)}
      </p>
      <p className="tabular-nums font-semibold" style={{ color: "var(--color-accent)" }}>
        {formatCurrency(firstEntry.value)}
      </p>
    </div>
  );
}

function dateToTimestamp(dateStr: string): number {
  return new Date(dateStr + "T00:00:00").getTime();
}

interface TimeDataPoint {
  timestamp: number;
  date: string;
  value: number;
}

function toTimeData(dataPoints: DataPoint[]): TimeDataPoint[] {
  return dataPoints.map((dp) => ({
    timestamp: dateToTimestamp(dp.date),
    date: dp.date,
    value: dp.value,
  }));
}

function tickCountForPeriod(period: string): number {
  switch (period) {
    case "1m":
      return 4;
    case "1y":
      return 6;
    case "5y":
      return 5;
    default:
      return 6;
  }
}

function generateEvenTicks(minTs: number, maxTs: number, count: number): number[] {
  if (count <= 1 || minTs === maxTs) return [minTs];
  const step = (maxTs - minTs) / (count - 1);
  const ticks: number[] = [];
  for (let i = 0; i < count; i++) {
    ticks.push(Math.round(minTs + step * i));
  }
  return ticks;
}

function formatTickLabel(ts: number, period: string): string {
  const d = new Date(ts);
  switch (period) {
    case "1m":
      return d.toLocaleDateString("en-US", { month: "short", day: "numeric" });
    case "1y":
      return d.toLocaleDateString("en-US", { month: "short", year: "numeric" });
    case "5y":
      return d.toLocaleDateString("en-US", { month: "short", year: "numeric" });
    default:
      return d.toLocaleDateString("en-US", { year: "numeric" });
  }
}

function domainForPeriod(period: string, dataMinTs: number): [number, number] {
  const now = new Date();
  const endTs = now.getTime();
  switch (period) {
    case "1m": {
      const start = new Date(now);
      start.setDate(start.getDate() - 30);
      return [start.getTime(), endTs];
    }
    case "1y": {
      const start = new Date(now);
      start.setFullYear(start.getFullYear() - 1);
      return [start.getTime(), endTs];
    }
    case "5y": {
      const start = new Date(now);
      start.setFullYear(start.getFullYear() - 5);
      return [start.getTime(), endTs];
    }
    default:
      return [dataMinTs, endTs];
  }
}

export default function NetWorthChart({ dataPoints, period, onPeriodChange }: NetWorthChartProps) {
  const allTimeData = toTimeData(dataPoints);
  const timestamps = allTimeData.map((d) => d.timestamp);
  const dataMinTs = Math.min(...timestamps);
  const [periodMin, domainMax] = domainForPeriod(period, dataMinTs);
  const domainMin = Math.max(periodMin, dataMinTs);
  const timeData = allTimeData.filter((d) => d.timestamp >= periodMin);
  const ticks = generateEvenTicks(domainMin, domainMax, tickCountForPeriod(period));

  return (
    <div
      className="rounded-lg p-6"
      style={{
        backgroundColor: "var(--color-surface)",
        border: "1px solid var(--color-border)",
        boxShadow: "var(--glow-sm)",
      }}
    >
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-sm font-medium" style={{ color: "var(--color-text-secondary)" }}>
          Net Worth Over Time
        </h3>
        <PeriodToggle activePeriod={period} onPeriodChange={onPeriodChange} />
      </div>
      {dataPoints.length === 0 ? (
        <div
          className="flex h-48 items-center justify-center text-sm"
          style={{ color: "var(--color-text-secondary)" }}
        >
          No data available for this period
        </div>
      ) : (
        <>
          <div aria-hidden="true">
            <ResponsiveContainer width="100%" height={200}>
              <AreaChart data={timeData}>
                <defs>
                  <linearGradient id="netWorthGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#22c55e" stopOpacity={0.3} />
                    <stop offset="100%" stopColor="#22c55e" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <XAxis
                  dataKey="timestamp"
                  type="number"
                  scale="time"
                  domain={[domainMin, domainMax]}
                  ticks={ticks}
                  tickFormatter={(val: number) => formatTickLabel(val, period)}
                  tick={{ fill: "var(--color-text-secondary)", fontSize: 11 }}
                  axisLine={false}
                  tickLine={false}
                />
                <YAxis
                  tickFormatter={(val: number) => formatCurrency(val)}
                  tick={{ fill: "var(--color-text-secondary)", fontSize: 11 }}
                  axisLine={false}
                  tickLine={false}
                  width={90}
                  domain={[
                    (min: number) => Math.floor(min * 0.95),
                    (max: number) => Math.ceil(max * 1.05),
                  ]}
                />
                <Tooltip content={<ChartTooltip />} />
                <Area
                  type="monotone"
                  dataKey="value"
                  stroke="#22c55e"
                  strokeWidth={2}
                  fill="url(#netWorthGradient)"
                  animationDuration={600}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
          <table
            style={{
              position: "absolute",
              width: 1,
              height: 1,
              padding: 0,
              margin: -1,
              overflow: "hidden",
              clip: "rect(0,0,0,0)",
              whiteSpace: "nowrap",
              borderWidth: 0,
            }}
          >
            <caption>Net worth over time</caption>
            <thead>
              <tr>
                <th scope="col">Date</th>
                <th scope="col">Net Worth</th>
              </tr>
            </thead>
            <tbody>
              {dataPoints.map((dp) => (
                <tr key={dp.date}>
                  <td>{formatDate(dp.date)}</td>
                  <td>{formatCurrency(dp.value)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      )}
    </div>
  );
}
