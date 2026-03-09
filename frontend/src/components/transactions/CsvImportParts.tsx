import type { ChangeEvent, DragEvent } from "react";
import { Upload } from "lucide-react";
import type { CsvResult } from "@/utils/csv";
import type { AccountResponse } from "@/api/client";

export type MappableField = "date" | "amount" | "category" | "type" | "notes" | "tags" | "skip";

export const FIELD_OPTIONS: { value: MappableField; label: string }[] = [
  { value: "skip", label: "Skip" },
  { value: "date", label: "Date" },
  { value: "amount", label: "Amount" },
  { value: "category", label: "Category" },
  { value: "type", label: "Type" },
  { value: "notes", label: "Notes" },
  { value: "tags", label: "Tags" },
];

export function guessMapping(header: string): MappableField {
  const lower = header.toLowerCase().trim();
  if (lower === "date" || lower === "transaction date") return "date";
  if (lower === "amount" || lower === "value") return "amount";
  if (lower === "category") return "category";
  if (lower === "type" || lower === "transaction type") return "type";
  if (lower === "notes" || lower === "description" || lower === "memo") return "notes";
  if (lower === "tags" || lower === "labels") return "tags";
  return "skip";
}

export function UploadArea({
  fileInputRef,
  isDragging,
  onFileChange,
  onDrop,
  onDragOver,
  onDragLeave,
}: {
  fileInputRef: React.RefObject<HTMLInputElement | null>;
  isDragging: boolean;
  onFileChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onDrop: (event: DragEvent) => void;
  onDragOver: (event: DragEvent) => void;
  onDragLeave: () => void;
}) {
  return (
    <div
      className="flex cursor-pointer flex-col items-center gap-3 rounded-lg border-2 border-dashed p-8 transition-colors"
      style={{
        borderColor: isDragging ? "var(--color-accent)" : "var(--color-border)",
        backgroundColor: isDragging ? "var(--color-accent-muted)" : "transparent",
      }}
      onClick={() => fileInputRef.current?.click()}
      onDrop={onDrop}
      onDragOver={onDragOver}
      onDragLeave={onDragLeave}
    >
      <Upload size={32} style={{ color: "var(--color-text-secondary)" }} />
      <p className="text-sm" style={{ color: "var(--color-text-secondary)" }}>
        Drop a CSV file or click to browse
      </p>
      <input
        ref={fileInputRef}
        type="file"
        accept=".csv"
        className="hidden"
        onChange={onFileChange}
        data-testid="csv-file-input"
      />
    </div>
  );
}

export function PreviewSection({ csvData, fileName }: { csvData: CsvResult; fileName: string }) {
  const previewRows = csvData.rows.slice(0, 5);
  const rowLabel = csvData.rows.length === 1 ? "1 row" : `${csvData.rows.length} rows`;

  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <span className="text-sm font-medium" style={{ color: "var(--color-text-primary)" }}>
          Preview
        </span>
        <span className="text-xs" style={{ color: "var(--color-text-secondary)" }}>
          {fileName} — {rowLabel}
        </span>
      </div>
      <div
        className="overflow-x-auto rounded-md"
        style={{
          backgroundColor: "var(--color-background)",
          border: "1px solid var(--color-border)",
        }}
      >
        <table className="w-full text-xs">
          <thead>
            <tr style={{ borderBottom: "1px solid var(--color-border)" }}>
              {csvData.headers.map((header, idx) => (
                <th
                  key={idx}
                  className="px-3 py-2 text-left font-medium"
                  style={{ color: "var(--color-text-secondary)" }}
                >
                  {header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {previewRows.map((row, rowIdx) => (
              <tr key={rowIdx} style={{ borderBottom: "1px solid var(--color-border)" }}>
                {row.map((cell, cellIdx) => (
                  <td
                    key={cellIdx}
                    className="px-3 py-1.5"
                    style={{ color: "var(--color-text-primary)" }}
                  >
                    {cell}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export function MappingSection({
  headers,
  mapping,
  onUpdateMapping,
}: {
  headers: string[];
  mapping: MappableField[];
  onUpdateMapping: (index: number, value: MappableField) => void;
}) {
  return (
    <div>
      <span
        className="mb-2 block text-sm font-medium"
        style={{ color: "var(--color-text-primary)" }}
      >
        Column Mapping
      </span>
      <div className="flex flex-col gap-2">
        {headers.map((header, idx) => (
          <div key={idx} className="flex items-center gap-3">
            <span
              className="w-32 truncate text-xs"
              style={{ color: "var(--color-text-secondary)" }}
            >
              {header}
            </span>
            <select
              value={mapping[idx] ?? "skip"}
              onChange={(event) => onUpdateMapping(idx, event.target.value as MappableField)}
              className="rounded-md px-2 py-1 text-xs"
              style={{
                backgroundColor: "var(--color-background)",
                color: "var(--color-text-primary)",
                border: "1px solid var(--color-border)",
              }}
            >
              {FIELD_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
        ))}
      </div>
    </div>
  );
}

export function AccountSelect({
  accounts,
  value,
  onChange,
}: {
  accounts: AccountResponse[];
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <div>
      <label
        htmlFor="import-account"
        className="mb-1 block text-sm font-medium"
        style={{ color: "var(--color-text-primary)" }}
      >
        Account
      </label>
      <select
        id="import-account"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className="w-full rounded-md px-3 py-2 text-sm"
        style={{
          backgroundColor: "var(--color-background)",
          color: "var(--color-text-primary)",
          border: "1px solid var(--color-border)",
        }}
      >
        <option value="">Select an account</option>
        {accounts.map((account) => (
          <option key={account.id} value={account.id}>
            {account.name}
          </option>
        ))}
      </select>
    </div>
  );
}
