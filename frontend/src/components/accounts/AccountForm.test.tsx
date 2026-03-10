import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import AccountForm from "./AccountForm";

describe("AccountForm", () => {
  it("renders all fields", () => {
    render(<AccountForm onSubmit={vi.fn()} onCancel={vi.fn()} />);

    expect(screen.getByLabelText("Name")).toBeInTheDocument();
    expect(screen.getByLabelText("Institution")).toBeInTheDocument();
    expect(screen.getByLabelText("Account Type")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Add Account" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();
  });

  it("shows Save Changes when editing", () => {
    render(
      <AccountForm
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        initialValues={{ name: "Test", institution: "Bank", account_type: "checking" }}
      />,
    );

    expect(screen.getByRole("button", { name: "Save Changes" })).toBeInTheDocument();
    expect(screen.getByDisplayValue("Test")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Bank")).toBeInTheDocument();
  });

  it("calls onSubmit with form data", async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();
    render(<AccountForm onSubmit={onSubmit} onCancel={vi.fn()} />);

    await user.type(screen.getByLabelText("Name"), "My Savings");
    await user.type(screen.getByLabelText("Institution"), "Ally");

    await user.click(screen.getByRole("button", { name: "Add Account" }));

    expect(onSubmit).toHaveBeenCalledWith({
      name: "My Savings",
      institution: "Ally",
      account_type: "checking",
    });
  });

  it("calls onCancel when cancel clicked", async () => {
    const onCancel = vi.fn();
    const user = userEvent.setup();
    render(<AccountForm onSubmit={vi.fn()} onCancel={onCancel} />);

    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onCancel).toHaveBeenCalled();
  });

  it("disables submit when isSubmitting", () => {
    render(<AccountForm onSubmit={vi.fn()} onCancel={vi.fn()} isSubmitting />);

    expect(screen.getByRole("button", { name: "Add Account" })).toBeDisabled();
  });
});
