import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "@/test-utils";
import Transactions from "./Transactions";

const mockTransactionsResponse = {
  data: [
    {
      id: "txn-1",
      account_id: "acc-1",
      type: "expense" as const,
      amount: 42.5,
      category: "food",
      date: "2026-03-08",
      notes: "",
      tags: [],
      created_at: "2026-03-08T10:00:00Z",
      updated_at: "2026-03-08T10:00:00Z",
    },
  ],
  pagination: { page: 1, per_page: 25, total: 1, total_pages: 1 },
};

const mockAccountsResponse = {
  data: [
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
  ],
  pagination: { page: 1, per_page: 100, total: 1, total_pages: 1 },
};

beforeEach(() => {
  vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
    const url =
      typeof input === "string"
        ? input
        : input instanceof URL
          ? input.toString()
          : (input as Request).url;

    if (url.includes("/api/v1/transactions")) {
      return Promise.resolve(
        new Response(JSON.stringify(mockTransactionsResponse), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
      );
    }
    if (url.includes("/api/v1/accounts")) {
      return Promise.resolve(
        new Response(JSON.stringify(mockAccountsResponse), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
      );
    }
    return Promise.resolve(new Response("Not found", { status: 404 }));
  });
});

describe("Transactions", () => {
  it("renders page title and action buttons", async () => {
    renderWithProviders(<Transactions />);
    expect(screen.getByText("Transactions")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Add Transaction" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Import CSV" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Split Transaction" })).toBeInTheDocument();
  });

  it("loads and displays transactions", async () => {
    renderWithProviders(<Transactions />);
    await waitFor(() => {
      expect(screen.getByText("food")).toBeInTheDocument();
    });
    expect(screen.getByText("$42.50")).toBeInTheDocument();
  });

  it("opens add transaction modal", async () => {
    const user = userEvent.setup();
    renderWithProviders(<Transactions />);
    await user.click(screen.getByRole("button", { name: "Add Transaction" }));
    await waitFor(() => {
      expect(screen.getByText("Add Transaction", { selector: "h2" })).toBeInTheDocument();
    });
  });

  it("opens import csv modal", async () => {
    const user = userEvent.setup();
    renderWithProviders(<Transactions />);
    await user.click(screen.getByRole("button", { name: "Import CSV" }));
    await waitFor(() => {
      expect(screen.getByText("Import Transactions", { selector: "h2" })).toBeInTheDocument();
    });
  });

  it("opens split transaction modal", async () => {
    const user = userEvent.setup();
    renderWithProviders(<Transactions />);
    await user.click(screen.getByRole("button", { name: "Split Transaction" }));
    await waitFor(() => {
      expect(screen.getByText("Split Transaction", { selector: "h2" })).toBeInTheDocument();
    });
  });

  it("shows loading skeleton initially", () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(() => new Promise(() => {}));
    renderWithProviders(<Transactions />);
    expect(screen.getByTestId("transactions-skeleton")).toBeInTheDocument();
  });
});
