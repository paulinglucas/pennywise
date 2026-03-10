import { describe, it, expect, vi } from "vitest";
import { screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "@/test-utils";
import Modal from "./Modal";

describe("Modal", () => {
  it("renders children when open", () => {
    renderWithProviders(
      <Modal isOpen onClose={vi.fn()} title="Test Modal">
        <p>Modal content</p>
      </Modal>,
    );
    expect(screen.getByText("Test Modal")).toBeInTheDocument();
    expect(screen.getByText("Modal content")).toBeInTheDocument();
  });

  it("does not render when closed", () => {
    renderWithProviders(
      <Modal isOpen={false} onClose={vi.fn()} title="Hidden">
        <p>Hidden content</p>
      </Modal>,
    );
    expect(screen.queryByText("Hidden")).not.toBeInTheDocument();
  });

  it("calls onClose when backdrop is clicked", () => {
    const onClose = vi.fn();
    renderWithProviders(
      <Modal isOpen onClose={onClose} title="Closeable">
        <p>Content</p>
      </Modal>,
    );
    fireEvent.click(screen.getByTestId("modal-backdrop"));
    expect(onClose).toHaveBeenCalledOnce();
  });

  it("calls onClose when close button is clicked", () => {
    const onClose = vi.fn();
    renderWithProviders(
      <Modal isOpen onClose={onClose} title="Closeable">
        <p>Content</p>
      </Modal>,
    );
    fireEvent.click(screen.getByLabelText("Close"));
    expect(onClose).toHaveBeenCalledOnce();
  });

  it("does not close when modal content is clicked", () => {
    const onClose = vi.fn();
    renderWithProviders(
      <Modal isOpen onClose={onClose} title="Closeable">
        <p>Content</p>
      </Modal>,
    );
    fireEvent.click(screen.getByText("Content"));
    expect(onClose).not.toHaveBeenCalled();
  });

  it("has role=dialog and aria-modal", () => {
    renderWithProviders(
      <Modal isOpen onClose={vi.fn()} title="Accessible">
        <p>Content</p>
      </Modal>,
    );
    const dialog = screen.getByRole("dialog");
    expect(dialog).toHaveAttribute("aria-modal", "true");
  });

  it("has aria-labelledby pointing to title", () => {
    renderWithProviders(
      <Modal isOpen onClose={vi.fn()} title="My Title">
        <p>Content</p>
      </Modal>,
    );
    const dialog = screen.getByRole("dialog");
    const titleId = dialog.getAttribute("aria-labelledby");
    expect(titleId).toBeTruthy();
    const titleElement = document.getElementById(titleId as string);
    expect(titleElement).toHaveTextContent("My Title");
  });

  it("traps focus within the modal", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <Modal isOpen onClose={vi.fn()} title="Focus Trap">
        <input data-testid="first-input" />
        <button data-testid="last-button">Submit</button>
      </Modal>,
    );

    const closeButton = screen.getByLabelText("Close");
    const firstInput = screen.getByTestId("first-input");
    const lastButton = screen.getByTestId("last-button");

    closeButton.focus();
    expect(document.activeElement).toBe(closeButton);

    await user.tab();
    expect(document.activeElement).toBe(firstInput);

    await user.tab();
    expect(document.activeElement).toBe(lastButton);

    await user.tab();
    await waitFor(() => {
      expect(document.activeElement).toBe(closeButton);
    });
  });

  it("focuses first focusable element on open", async () => {
    renderWithProviders(
      <Modal isOpen onClose={vi.fn()} title="Auto Focus">
        <input data-testid="auto-focus-input" />
      </Modal>,
    );

    await waitFor(() => {
      const dialog = screen.getByRole("dialog");
      expect(dialog.contains(document.activeElement)).toBe(true);
    });
  });
});
