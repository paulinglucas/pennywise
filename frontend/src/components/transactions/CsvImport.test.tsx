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
