import { describe, it, expect, vi } from "vitest";
import { screen, fireEvent } from "@testing-library/react";
import { renderWithProviders } from "@/test-utils";
import AssetForm from "./AssetForm";

describe("AssetForm", () => {
  it("renders name and value fields", () => {
    renderWithProviders(<AssetForm onSubmit={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByLabelText("Name")).toBeInTheDocument();
    expect(screen.getByLabelText("Current Value")).toBeInTheDocument();
  });

  it("renders asset type selector", () => {
    renderWithProviders(<AssetForm onSubmit={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByLabelText("Asset Type")).toBeInTheDocument();
  });

  it("shows real estate fields when real_estate type selected", () => {
    renderWithProviders(<AssetForm onSubmit={vi.fn()} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByLabelText("Asset Type"), { target: { value: "real_estate" } });
    expect(screen.getByLabelText("Purchase Price")).toBeInTheDocument();
    expect(screen.getByLabelText("Mortgage Rate (%)")).toBeInTheDocument();
  });

  it("shows SAFE fields when speculative type selected", () => {
    renderWithProviders(<AssetForm onSubmit={vi.fn()} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByLabelText("Asset Type"), { target: { value: "speculative" } });
    expect(screen.getByLabelText("Company Name")).toBeInTheDocument();
    expect(screen.getByLabelText("Ownership (%)")).toBeInTheDocument();
  });

  it("calls onSubmit with form data", () => {
    const onSubmit = vi.fn();
    renderWithProviders(<AssetForm onSubmit={onSubmit} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Savings" } });
    fireEvent.change(screen.getByLabelText("Current Value"), { target: { value: "5000" } });
    fireEvent.submit(screen.getByRole("form"));
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "Savings",
        current_value: 5000,
        asset_type: "liquid",
      }),
    );
  });

  it("calls onCancel when cancel button clicked", () => {
    const onCancel = vi.fn();
    renderWithProviders(<AssetForm onSubmit={vi.fn()} onCancel={onCancel} />);
    fireEvent.click(screen.getByText("Cancel"));
    expect(onCancel).toHaveBeenCalled();
  });

  it("pre-fills fields when editing an existing asset", () => {
    renderWithProviders(
      <AssetForm
        onSubmit={vi.fn()}
        onCancel={vi.fn()}
        initialValues={{
          name: "Home",
          asset_type: "real_estate",
          current_value: 308000,
        }}
      />,
    );
    expect(screen.getByDisplayValue("Home")).toBeInTheDocument();
    expect(screen.getByDisplayValue("308000")).toBeInTheDocument();
  });
});
