import { describe, it, expect, vi } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "@/test-utils";
import ErrorState from "./ErrorState";

describe("ErrorState", () => {
  it("renders message and retry button", () => {
    const onRetry = vi.fn();
    renderWithProviders(<ErrorState message="Could not load data." onRetry={onRetry} />);

    expect(screen.getByText("Something went wrong")).toBeInTheDocument();
    expect(screen.getByText("Could not load data.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /retry/i })).toBeInTheDocument();
  });

  it("calls onRetry when retry button is clicked", async () => {
    const user = userEvent.setup();
    const onRetry = vi.fn();
    renderWithProviders(<ErrorState message="Error occurred." onRetry={onRetry} />);

    await user.click(screen.getByRole("button", { name: /retry/i }));
    expect(onRetry).toHaveBeenCalledOnce();
  });

  it("displays request ID when provided", () => {
    renderWithProviders(
      <ErrorState message="Server error." onRetry={() => {}} requestId="req-abc-123" />,
    );

    expect(screen.getByText("Request ID: req-abc-123")).toBeInTheDocument();
  });

  it("does not display request ID when not provided", () => {
    renderWithProviders(<ErrorState message="Server error." onRetry={() => {}} />);

    expect(screen.queryByText(/request id/i)).not.toBeInTheDocument();
  });

  it("renders custom title when provided", () => {
    renderWithProviders(
      <ErrorState title="Connection lost" message="Check your network." onRetry={() => {}} />,
    );

    expect(screen.getByText("Connection lost")).toBeInTheDocument();
    expect(screen.queryByText("Something went wrong")).not.toBeInTheDocument();
  });
});
