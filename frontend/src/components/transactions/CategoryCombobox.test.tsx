import { describe, it, expect, vi } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "@/test-utils";
import CategoryCombobox from "./CategoryCombobox";

vi.mock("@/hooks/useCategories", () => ({
  useCategories: () => ({
    data: ["food", "rent", "salary", "utilities"],
    isLoading: false,
  }),
}));

describe("CategoryCombobox", () => {
  const onChange = vi.fn<(value: string) => void>();

  it("renders with the given value", () => {
    renderWithProviders(<CategoryCombobox value="food" onChange={onChange} />);
    expect(screen.getByDisplayValue("food")).toBeInTheDocument();
  });

  it("shows dropdown on focus with matching categories", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CategoryCombobox value="" onChange={onChange} />);

    await user.click(screen.getByRole("combobox"));

    await waitFor(() => {
      expect(screen.getByText("food")).toBeInTheDocument();
      expect(screen.getByText("rent")).toBeInTheDocument();
    });
  });

  it("filters categories based on current value", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CategoryCombobox value="re" onChange={onChange} />);

    await user.click(screen.getByRole("combobox"));

    await waitFor(() => {
      expect(screen.getByText("rent")).toBeInTheDocument();
      expect(screen.queryByText("salary")).not.toBeInTheDocument();
    });
  });

  it("calls onChange when selecting a category", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CategoryCombobox value="" onChange={onChange} />);

    await user.click(screen.getByRole("combobox"));

    await waitFor(() => {
      expect(screen.getByText("rent")).toBeInTheDocument();
    });

    await user.click(screen.getByText("rent"));
    expect(onChange).toHaveBeenCalledWith("rent");
  });

  it("allows free text input for new categories", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CategoryCombobox value="" onChange={onChange} />);

    await user.type(screen.getByRole("combobox"), "new_cat");
    expect(onChange).toHaveBeenCalled();
  });

  it("shows add option when typed text has no exact match", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CategoryCombobox value="groc" onChange={onChange} />);

    await user.click(screen.getByRole("combobox"));

    await waitFor(() => {
      expect(screen.getByText('Add "groc"')).toBeInTheDocument();
    });
  });
});
