import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "@/test-utils";
import AssetTransactions from "./AssetTransactions";

const mockTransactions = [
  {
    id: "t1",
    user_id: "u1",
    account_id: "acc1",
    type: "withdrawal" as const,
    amount: 50,
    category: "Groceries",
    date: "2026-03-01",
    notes: "",
    tags: [],
    created_at: "2026-03-01T00:00:00Z",
    updated_at: "2026-03-01T00:00:00Z",
  },
  {
    id: "t2",
    user_id: "u1",
    account_id: "acc1",
    type: "deposit" as const,
    amount: 3000,
    category: "Salary",
    date: "2026-03-01",
    notes: "",
    tags: [],
    created_at: "2026-03-01T00:00:00Z",
    updated_at: "2026-03-01T00:00:00Z",
  },
];

const mockFetch = vi.fn();

beforeEach(() => {
  mockFetch.mockReset();
  globalThis.fetch = mockFetch;
});

function mockSuccessResponse(data: unknown[], total: number, page = 1, totalPages = 1) {
  mockFetch.mockResolvedValueOnce({
    ok: true,
    status: 200,
    headers: new Headers({ "content-type": "application/json" }),
    json: () =>
      Promise.resolve({
        data,
        pagination: { page, per_page: 10, total, total_pages: totalPages },
      }),
  });
}

describe("AssetTransactions", () => {
  it("renders loading state initially", () => {
    mockFetch.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<AssetTransactions accountId="acc1" accountName="Fidelity" />);
    expect(screen.getByText("Loading transactions...")).toBeInTheDocument();
  });

  it("renders empty state when no transactions", async () => {
    mockSuccessResponse([], 0);
    renderWithProviders(<AssetTransactions accountId="acc1" accountName="Fidelity" />);
    await waitFor(() => {
      expect(screen.getByText("No transactions for Fidelity")).toBeInTheDocument();
    });
  });

  it("renders transaction rows", async () => {
    mockSuccessResponse(mockTransactions, 2);
    renderWithProviders(<AssetTransactions accountId="acc1" accountName="Fidelity" />);
    await waitFor(() => {
      expect(screen.getByText("Groceries")).toBeInTheDocument();
    });
    expect(screen.getByText("Salary")).toBeInTheDocument();
  });

  it("shows deposits with positive prefix", async () => {
    mockSuccessResponse(mockTransactions, 2);
    renderWithProviders(<AssetTransactions accountId="acc1" accountName="Fidelity" />);
    await waitFor(() => {
      expect(screen.getByText("Salary")).toBeInTheDocument();
    });
    const depositAmount = screen.getByText("+$3,000.00");
    expect(depositAmount).toBeInTheDocument();
  });

  it("renders pagination controls when multiple pages", async () => {
    mockSuccessResponse(mockTransactions, 25, 1, 3);
    renderWithProviders(<AssetTransactions accountId="acc1" accountName="Fidelity" />);
    await waitFor(() => {
      expect(screen.getByText("25 transactions")).toBeInTheDocument();
    });
    expect(screen.getByText("1/3")).toBeInTheDocument();
    expect(screen.getByLabelText("Previous page")).toBeDisabled();
    expect(screen.getByLabelText("Next page")).not.toBeDisabled();
  });

  it("navigates to next page on click", async () => {
    const user = userEvent.setup();
    mockSuccessResponse(mockTransactions, 25, 1, 3);
    renderWithProviders(<AssetTransactions accountId="acc1" accountName="Fidelity" />);
    await waitFor(() => {
      expect(screen.getByText("1/3")).toBeInTheDocument();
    });
    mockSuccessResponse(mockTransactions, 25, 2, 3);
    await user.click(screen.getByLabelText("Next page"));
    await waitFor(() => {
      expect(screen.getByText("2/3")).toBeInTheDocument();
    });
  });
});
