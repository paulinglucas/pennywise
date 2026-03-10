import { ResponsiveContainer, AreaChart, Area, XAxis, YAxis, Tooltip } from "recharts";
import { formatCurrency, formatDate } from "@/utils/formatting";
import type { AssetHistoryEntry } from "@/api/client";

interface AssetSparklineProps {
  entries: AssetHistoryEntry[];
  currentValue: number;
  color: string;
  gradientId: string;
  period: string;
}

interface TimeDataPoint {
  timestamp: number;
  date: string;
  value: number;
}

function extractDate(isoString: string): string {
  const idx = isoString.indexOf("T");
  return idx >= 0 ? isoString.slice(0, idx) : isoString;
}

function toTimeData(entries: AssetHistoryEntry[]): TimeDataPoint[] {
  return entries.map((e) => ({
    timestamp: new Date(e.recorded_at).getTime(),
    date: extractDate(e.recorded_at),
    value: e.value,
  }));
}

export function computeChange(entries: AssetHistoryEntry[]): number | null {
  if (entries.length < 2) return null;
  const first = entries[0];
  const last = entries[entries.length - 1];
  if (!first || !last || first.value === 0) return null;
  return ((last.value - first.value) / first.value) * 100;
}

function ChartTooltip({
  active,
  payload,
}: {
  active?: boolean;
  payload?: Array<{ value: number; payload: TimeDataPoint }>;
  label?: number;
}) {
  const entry = payload?.[0];
  if (!active || !entry) return null;

  return (
    <div
      className="rounded-md px-3 py-2 text-sm"
      style={{
        backgroundColor: "var(--color-surface)",
        border: "1px solid var(--color-border)",
        boxShadow: "var(--glow-sm)",
      }}
    >
      <p className="mb-1 text-xs" style={{ color: "var(--color-text-secondary)" }}>
        {formatDate(entry.payload.date)}
      </p>
      <p className="tabular-nums font-semibold" style={{ color: "var(--color-text-primary)" }}>
        {formatCurrency(entry.value)}
      </p>
    </div>
  );
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
    case "3m": {
      const start = new Date(now);
      start.setMonth(start.getMonth() - 3);
      return [start.getTime(), endTs];
    }
    case "6m": {
      const start = new Date(now);
      start.setMonth(start.getMonth() - 6);
      return [start.getTime(), endTs];
    }
    case "1y": {
      const start = new Date(now);
      start.setFullYear(start.getFullYear() - 1);
      return [start.getTime(), endTs];
    }
    default:
      return [dataMinTs, endTs];
  }
}

function todayDateString(): string {
  const d = new Date();
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

export default function AssetSparkline({ entries, currentValue, color, gradientId, period }: AssetSparklineProps) {
  if (entries.length === 0) return null;

  const timeData = toTimeData(entries);
  const now = Date.now();
  const lastPoint = timeData[timeData.length - 1];
  if (lastPoint && now - lastPoint.timestamp > 24 * 60 * 60 * 1000) {
    timeData.push({ timestamp: now, date: todayDateString(), value: currentValue });
  }

  if (timeData.length < 2) return null;

  const timestamps = timeData.map((d) => d.timestamp);
  const dataMinTs = Math.min(...timestamps);
  const [domainMin, domainMax] = domainForPeriod(period, dataMinTs);

  return (
    <div className="mt-auto flex min-h-[60px] flex-1 items-end px-3 pb-3 pt-2" aria-hidden="true">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={timeData} margin={{ top: 4, right: 0, bottom: 0, left: 0 }}>
          <defs>
            <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={color} stopOpacity={0.25} />
              <stop offset="100%" stopColor={color} stopOpacity={0} />
            </linearGradient>
          </defs>
          <XAxis
            dataKey="timestamp"
            type="number"
            scale="time"
            domain={[domainMin, domainMax]}
            hide
          />
          <YAxis
            hide
            domain={[
              (min: number) => Math.floor(min * 0.95),
              (max: number) => Math.ceil(max * 1.05),
            ]}
          />
          <Tooltip content={<ChartTooltip />} />
          <Area
            type="monotone"
            dataKey="value"
            stroke={color}
            strokeWidth={1.5}
            fill={`url(#${gradientId})`}
            animationDuration={400}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
