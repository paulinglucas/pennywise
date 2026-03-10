import { describe, it, expect, vi, beforeEach, afterEach, type MockInstance } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "@/test-utils";
import ErrorBoundary from "./ErrorBoundary";

function ThrowingComponent({ shouldThrow }: { shouldThrow: boolean }) {
  if (shouldThrow) {
    throw new Error("Test crash");
  }
  return <div>Content rendered</div>;
}

describe("ErrorBoundary", () => {
  let consoleErrorSpy: MockInstance;

  beforeEach(() => {
    consoleErrorSpy = vi.spyOn(globalThis.console, "error").mockImplementation(() => {});
  });

  afterEach(() => {
    consoleErrorSpy.mockRestore();
  });

  it("renders children when no error occurs", () => {
    renderWithProviders(
      <ErrorBoundary>
        <ThrowingComponent shouldThrow={false} />
      </ErrorBoundary>,
    );

    expect(screen.getByText("Content rendered")).toBeInTheDocument();
  });

  it("renders error UI when a child throws", () => {
    renderWithProviders(
      <ErrorBoundary>
        <ThrowingComponent shouldThrow={true} />
      </ErrorBoundary>,
    );

    expect(screen.getByText("Something went wrong")).toBeInTheDocument();
    expect(screen.getByText(/unexpected error/i)).toBeInTheDocument();
    expect(screen.queryByText("Content rendered")).not.toBeInTheDocument();
  });

  it("provides a reload button", async () => {
    const reloadMock = vi.fn();
    Object.defineProperty(window, "location", {
      value: { reload: reloadMock },
      writable: true,
    });

    const user = userEvent.setup();
    renderWithProviders(
      <ErrorBoundary>
        <ThrowingComponent shouldThrow={true} />
      </ErrorBoundary>,
    );

    await user.click(screen.getByRole("button", { name: /reload/i }));
    expect(reloadMock).toHaveBeenCalledOnce();
  });
});
