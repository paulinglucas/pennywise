import { describe, it, expect, vi } from "vitest";
import { screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "@/test-utils";
import CsvImport from "./CsvImport";
import type { AccountResponse } from "@/api/client";

const mockAccounts: AccountResponse[] = [
  {
    id: "acc-1",
    user_id: "user-1",
    name: "Checking",
    account_type: "checking",
    currency: "USD",
    institution: "Bank",
    is_active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
  {
    id: "acc-2",
    user_id: "user-1",
    name: "Savings",
    account_type: "savings",
    currency: "USD",
    institution: "Bank",
    is_active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

describe("CsvImport", () => {
  it("renders file upload area", () => {
    renderWithProviders(
      <CsvImport accounts={mockAccounts} onImported={vi.fn()} onClose={vi.fn()} />,
    );
    expect(screen.getByText("Drop a CSV file or click to browse")).toBeInTheDocument();
  });

  it("shows preview after file selection", async () => {
    renderWithProviders(
      <CsvImport accounts={mockAccounts} onImported={vi.fn()} onClose={vi.fn()} />,
    );

    const csvContent = "date,amount,category,type\n2026-01-15,42.50,food,expense";
    const file = new File([csvContent], "test.csv", { type: "text/csv" });

    const input = screen.getByTestId("csv-file-input");
    await userEvent.upload(input, file);

    await waitFor(() => {
      expect(screen.getByText("Preview")).toBeInTheDocument();
    });
    expect(screen.getByText(/1 row/)).toBeInTheDocument();
  });

  it("shows column mapping dropdowns", async () => {
    renderWithProviders(
      <CsvImport accounts={mockAccounts} onImported={vi.fn()} onClose={vi.fn()} />,
    );

    const csvContent = "date,amount,category,type\n2026-01-15,42.50,food,expense";
    const file = new File([csvContent], "test.csv", { type: "text/csv" });

    const input = screen.getByTestId("csv-file-input");
    await userEvent.upload(input, file);

    await waitFor(() => {
      expect(screen.getByText("Column Mapping")).toBeInTheDocument();
    });
  });

  it("requires account selection before import", async () => {
    renderWithProviders(
      <CsvImport accounts={mockAccounts} onImported={vi.fn()} onClose={vi.fn()} />,
    );

    const csvContent = "date,amount,category,type\n2026-01-15,42.50,food,expense";
    const file = new File([csvContent], "test.csv", { type: "text/csv" });

    const input = screen.getByTestId("csv-file-input");
    await userEvent.upload(input, file);

    await waitFor(() => {
      expect(screen.getByText("Preview")).toBeInTheDocument();
    });

    const importButton = screen.getByRole("button", { name: "Import" });
    expect(importButton).toBeDisabled();
  });

  it("handles drag and drop file upload", async () => {
    renderWithProviders(
      <CsvImport accounts={mockAccounts} onImported={vi.fn()} onClose={vi.fn()} />,
    );

    const dropZone = screen.getByText("Drop a CSV file or click to browse").closest("div")!;

    const csvContent = "date,amount,category,type\n2026-01-15,42.50,food,expense";
    const file = new File([csvContent], "test.csv", { type: "text/csv" });

    const dataTransfer = {
      files: [file],
      types: ["Files"],
    };

    fireEvent.dragOver(dropZone, { dataTransfer });
    fireEvent.drop(dropZone, { dataTransfer });

    await waitFor(() => {
      expect(screen.getByText("Preview")).toBeInTheDocument();
    });
  });

  it("handles drag leave", () => {
    renderWithProviders(
      <CsvImport accounts={mockAccounts} onImported={vi.fn()} onClose={vi.fn()} />,
    );

    const dropZone = screen.getByText("Drop a CSV file or click to browse").closest("div")!;

    fireEvent.dragOver(dropZone, { dataTransfer: { files: [], types: ["Files"] } });
    fireEvent.dragLeave(dropZone);
  });

  it("submits import when account is selected and import clicked", async () => {
    const mockFetch = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ imported: 1, skipped: 0, errors: [] }), {
        status: 201,
        headers: { "Content-Type": "application/json" },
      }),
    );

    const onImported = vi.fn();
    const onClose = vi.fn();

    renderWithProviders(
      <CsvImport accounts={mockAccounts} onImported={onImported} onClose={onClose} />,
    );

    const csvContent = "date,amount,category,type\n2026-01-15,42.50,food,expense";
    const file = new File([csvContent], "test.csv", { type: "text/csv" });

    const input = screen.getByTestId("csv-file-input");
    await userEvent.upload(input, file);

    await waitFor(() => {
      expect(screen.getByText("Preview")).toBeInTheDocument();
    });

    const accountSelect = screen.getByLabelText("Account");
    fireEvent.change(accountSelect, { target: { value: "acc-1" } });

    const importButton = screen.getByRole("button", { name: "Import" });
    fireEvent.click(importButton);

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalled();
    });

    mockFetch.mockRestore();
  });

  it("updates column mapping when changed", async () => {
    renderWithProviders(
      <CsvImport accounts={mockAccounts} onImported={vi.fn()} onClose={vi.fn()} />,
    );

    const csvContent = "col1,col2,col3,col4\n2026-01-15,42.50,food,expense";
    const file = new File([csvContent], "test.csv", { type: "text/csv" });

    const input = screen.getByTestId("csv-file-input");
    await userEvent.upload(input, file);

    await waitFor(() => {
      expect(screen.getByText("Column Mapping")).toBeInTheDocument();
    });

    const selects = screen.getAllByRole("combobox");
    if (selects.length > 0) {
      fireEvent.change(selects[0] as HTMLElement, { target: { value: "date" } });
    }
  });

  it("enables import button when account is selected", async () => {
    renderWithProviders(
      <CsvImport accounts={mockAccounts} onImported={vi.fn()} onClose={vi.fn()} />,
    );

    const csvContent = "date,amount,category,type\n2026-01-15,42.50,food,expense";
    const file = new File([csvContent], "test.csv", { type: "text/csv" });

    const input = screen.getByTestId("csv-file-input");
    await userEvent.upload(input, file);

    await waitFor(() => {
      expect(screen.getByText("Preview")).toBeInTheDocument();
    });

    const accountSelect = screen.getByLabelText("Account");
    fireEvent.change(accountSelect, { target: { value: "acc-1" } });

    const importButton = screen.getByRole("button", { name: "Import" });
    expect(importButton).not.toBeDisabled();
  });
});
