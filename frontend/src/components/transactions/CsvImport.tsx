import { useState, useRef, type ChangeEvent, type DragEvent } from "react";
import { parseCsv, type CsvResult } from "@/utils/csv";
import type { AccountResponse } from "@/api/client";
import { useImportTransactions } from "@/hooks/useTransactions";
import {
  guessMapping,
  UploadArea,
  PreviewSection,
  MappingSection,
  AccountSelect,
  type MappableField,
} from "./CsvImportParts";

interface CsvImportProps {
  accounts: AccountResponse[];
  onImported: () => void;
  onClose: () => void;
}

export default function CsvImport({ accounts, onImported, onClose }: CsvImportProps) {
  const [csvData, setCsvData] = useState<CsvResult | null>(null);
  const [file, setFile] = useState<File | null>(null);
  const [columnMapping, setColumnMapping] = useState<MappableField[]>([]);
  const [accountId, setAccountId] = useState("");
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const importMutation = useImportTransactions();

  function handleFile(selectedFile: File) {
    setFile(selectedFile);
    const reader = new FileReader();
    reader.onload = (event) => {
      const text = event.target?.result as string;
      const parsed = parseCsv(text);
      setCsvData(parsed);
      setColumnMapping(parsed.headers.map(guessMapping));
    };
    reader.readAsText(selectedFile);
  }

  function handleFileChange(event: ChangeEvent<HTMLInputElement>) {
    const selectedFile = event.target.files?.[0];
    if (selectedFile) handleFile(selectedFile);
  }

  function handleDrop(event: DragEvent) {
    event.preventDefault();
    setIsDragging(false);
    const droppedFile = event.dataTransfer.files[0];
    if (droppedFile) handleFile(droppedFile);
  }

  function handleDragOver(event: DragEvent) {
    event.preventDefault();
    setIsDragging(true);
  }

  function handleDragLeave() {
    setIsDragging(false);
  }

  function updateMapping(index: number, value: MappableField) {
    setColumnMapping((prev) => {
      const next = [...prev];
      next[index] = value;
      return next;
    });
  }

  function handleImport() {
    if (!file || !accountId) return;
    importMutation.mutate(
      { file, accountId },
      {
        onSuccess: () => {
          onImported();
          onClose();
        },
      },
    );
  }

  const canImport = file !== null && accountId !== "" && !importMutation.isPending;

  if (!csvData) {
    return (
      <UploadArea
        fileInputRef={fileInputRef}
        isDragging={isDragging}
        onFileChange={handleFileChange}
        onDrop={handleDrop}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
      />
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <PreviewSection csvData={csvData} fileName={file?.name ?? ""} />
      <MappingSection
        headers={csvData.headers}
        mapping={columnMapping}
        onUpdateMapping={updateMapping}
      />
      <AccountSelect accounts={accounts} value={accountId} onChange={setAccountId} />
      {importMutation.isError && (
        <p className="text-sm" style={{ color: "var(--color-negative)" }}>
          Import failed. Please check your file and try again.
        </p>
      )}
      {importMutation.isSuccess && importMutation.data && (
        <p className="text-sm" style={{ color: "var(--color-positive)" }}>
          Imported {importMutation.data.imported} transactions.
        </p>
      )}
      <div className="flex justify-end gap-2">
        <button
          onClick={onClose}
          className="btn-icon rounded-md px-4 py-2 text-sm transition-all"
          style={{ color: "var(--color-text-secondary)" }}
        >
          Cancel
        </button>
        <button
          onClick={handleImport}
          disabled={!canImport}
          className="btn-primary rounded-md px-4 py-2 text-sm font-medium transition-all disabled:opacity-40"
        >
          {importMutation.isPending ? "Importing..." : "Import"}
        </button>
      </div>
    </div>
  );
}
