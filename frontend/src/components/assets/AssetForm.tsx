import { useState, type FormEvent } from "react";
import type { AssetType, AccountResponse, CreateAssetRequest } from "@/api/client";

interface AssetFormProps {
  onSubmit: (data: CreateAssetRequest) => void;
  onCancel: () => void;
  isSubmitting?: boolean;
  initialValues?: Partial<CreateAssetRequest & { metadata?: Record<string, unknown> }>;
  accounts?: AccountResponse[];
}

const assetTypes: { key: AssetType; label: string }[] = [
  { key: "liquid", label: "Liquid" },
  { key: "retirement", label: "Retirement" },
  { key: "real_estate", label: "Real Estate" },
  { key: "brokerage", label: "Brokerage" },
  { key: "speculative", label: "Speculative" },
  { key: "other", label: "Other" },
];

const inputStyle = {
  backgroundColor: "var(--color-background)",
  borderColor: "var(--color-border)",
  color: "var(--color-text-primary)",
};

export default function AssetForm({
  onSubmit,
  onCancel,
  isSubmitting = false,
  initialValues,
  accounts,
}: AssetFormProps) {
  const meta = initialValues?.metadata ?? {};
  const [name, setName] = useState(initialValues?.name ?? "");
  const [assetType, setAssetType] = useState<AssetType>(initialValues?.asset_type ?? "liquid");
  const [currentValue, setCurrentValue] = useState(initialValues?.current_value?.toString() ?? "");
  const [accountId, setAccountId] = useState(initialValues?.account_id ?? "");

  const [purchasePrice, setPurchasePrice] = useState(
    (meta.purchase_price as number)?.toString() ?? "",
  );
  const [downPaymentPercent, setDownPaymentPercent] = useState(
    (meta.down_payment_percent as number)?.toString() ?? "5",
  );
  const [mortgageRate, setMortgageRate] = useState(
    (meta.mortgage_rate as number)?.toString() ?? "",
  );
  const [mortgageTermYears, setMortgageTermYears] = useState(
    (meta.mortgage_term_years as number)?.toString() ?? "30",
  );
  const [monthlyPayment, setMonthlyPayment] = useState(
    (meta.monthly_payment as number)?.toString() ?? "",
  );
  const [purchaseDate, setPurchaseDate] = useState((meta.purchase_date as string) ?? "");

  const [companyName, setCompanyName] = useState((meta.company_name as string) ?? "");
  const [ownershipPercentage, setOwnershipPercentage] = useState(
    (meta.ownership_percentage as number)?.toString() ?? "",
  );
  const [valuationCap, setValuationCap] = useState(
    (meta.valuation_cap as number)?.toString() ?? "",
  );

  function handleSubmit(event: FormEvent) {
    event.preventDefault();

    const data: CreateAssetRequest = {
      name,
      asset_type: assetType,
      current_value: parseFloat(currentValue),
      ...(accountId ? { account_id: accountId } : {}),
    };

    if (assetType === "real_estate") {
      data.metadata = {
        purchase_price: parseFloat(purchasePrice),
        down_payment_percent: parseFloat(downPaymentPercent),
        mortgage_rate: parseFloat(mortgageRate),
        mortgage_term_years: parseInt(mortgageTermYears, 10),
        monthly_payment: parseFloat(monthlyPayment),
        purchase_date: purchaseDate,
        current_valuation: parseFloat(currentValue),
        extra_principal_monthly: 0,
      };
    } else if (assetType === "speculative") {
      data.metadata = {
        company_name: companyName,
        ownership_percentage: parseFloat(ownershipPercentage),
        valuation_cap: parseFloat(valuationCap),
      };
    }

    onSubmit(data);
  }

  const isEditing = !!initialValues;

  return (
    <form role="form" onSubmit={handleSubmit} className="flex flex-col gap-4">
      <FormField label="Name" htmlFor="asset-name">
        <input
          id="asset-name"
          type="text"
          required
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="form-input"
          style={inputStyle}
          placeholder="e.g. Home, Roth IRA"
        />
      </FormField>

      <FormField label="Asset Type" htmlFor="asset-type">
        <select
          id="asset-type"
          value={assetType}
          onChange={(e) => setAssetType(e.target.value as AssetType)}
          className="form-input"
          style={inputStyle}
        >
          {assetTypes.map((t) => (
            <option key={t.key} value={t.key}>
              {t.label}
            </option>
          ))}
        </select>
      </FormField>

      <FormField label="Current Value" htmlFor="asset-value">
        <input
          id="asset-value"
          type="number"
          required
          min="0"
          step="0.01"
          value={currentValue}
          onChange={(e) => setCurrentValue(e.target.value)}
          className="form-input tabular-nums"
          style={inputStyle}
          placeholder="0.00"
        />
      </FormField>

      {accounts && accounts.length > 0 && (
        <FormField label="Linked Account" htmlFor="asset-account">
          <select
            id="asset-account"
            value={accountId}
            onChange={(e) => setAccountId(e.target.value)}
            className="form-input"
            style={inputStyle}
          >
            <option value="">None</option>
            {accounts.map((acc) => (
              <option key={acc.id} value={acc.id}>
                {acc.institution} - {acc.name}
              </option>
            ))}
          </select>
        </FormField>
      )}

      {assetType === "real_estate" && (
        <RealEstateFields
          purchasePrice={purchasePrice}
          onPurchasePriceChange={setPurchasePrice}
          downPaymentPercent={downPaymentPercent}
          onDownPaymentPercentChange={setDownPaymentPercent}
          mortgageRate={mortgageRate}
          onMortgageRateChange={setMortgageRate}
          mortgageTermYears={mortgageTermYears}
          onMortgageTermYearsChange={setMortgageTermYears}
          monthlyPayment={monthlyPayment}
          onMonthlyPaymentChange={setMonthlyPayment}
          purchaseDate={purchaseDate}
          onPurchaseDateChange={setPurchaseDate}
        />
      )}

      {assetType === "speculative" && (
        <SafeFields
          companyName={companyName}
          onCompanyNameChange={setCompanyName}
          ownershipPercentage={ownershipPercentage}
          onOwnershipPercentageChange={setOwnershipPercentage}
          valuationCap={valuationCap}
          onValuationCapChange={setValuationCap}
        />
      )}

      <div className="mt-2 flex gap-3">
        <button
          type="submit"
          disabled={isSubmitting}
          className="btn-primary flex-1 rounded-md px-4 py-2 text-sm font-medium transition-all disabled:opacity-50"
          style={{
            backgroundColor: "var(--color-accent)",
            color: "var(--color-background)",
            boxShadow: "var(--glow-accent)",
          }}
        >
          {isEditing ? "Save Changes" : "Add Asset"}
        </button>
        <button
          type="button"
          onClick={onCancel}
          className="rounded-md px-4 py-2 text-sm font-medium transition-colors"
          style={{ color: "var(--color-text-secondary)" }}
        >
          Cancel
        </button>
      </div>
    </form>
  );
}

function FormField({
  label,
  htmlFor,
  children,
}: {
  label: string;
  htmlFor: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <label
        htmlFor={htmlFor}
        className="mb-1 block text-sm font-medium"
        style={{ color: "var(--color-text-secondary)" }}
      >
        {label}
      </label>
      {children}
    </div>
  );
}

function RealEstateFields({
  purchasePrice,
  onPurchasePriceChange,
  downPaymentPercent,
  onDownPaymentPercentChange,
  mortgageRate,
  onMortgageRateChange,
  mortgageTermYears,
  onMortgageTermYearsChange,
  monthlyPayment,
  onMonthlyPaymentChange,
  purchaseDate,
  onPurchaseDateChange,
}: {
  purchasePrice: string;
  onPurchasePriceChange: (v: string) => void;
  downPaymentPercent: string;
  onDownPaymentPercentChange: (v: string) => void;
  mortgageRate: string;
  onMortgageRateChange: (v: string) => void;
  mortgageTermYears: string;
  onMortgageTermYearsChange: (v: string) => void;
  monthlyPayment: string;
  onMonthlyPaymentChange: (v: string) => void;
  purchaseDate: string;
  onPurchaseDateChange: (v: string) => void;
}) {
  return (
    <div
      className="flex flex-col gap-3 rounded-md p-3"
      style={{ backgroundColor: "var(--color-background)" }}
    >
      <p className="text-xs font-medium" style={{ color: "var(--color-text-secondary)" }}>
        Mortgage Details
      </p>
      <FormField label="Purchase Price" htmlFor="purchase-price">
        <input
          id="purchase-price"
          type="number"
          min="0"
          step="1"
          value={purchasePrice}
          onChange={(e) => onPurchasePriceChange(e.target.value)}
          className="form-input tabular-nums"
          style={inputStyle}
        />
      </FormField>
      <FormField label="Down Payment (%)" htmlFor="down-payment">
        <input
          id="down-payment"
          type="number"
          min="0"
          max="100"
          step="0.1"
          value={downPaymentPercent}
          onChange={(e) => onDownPaymentPercentChange(e.target.value)}
          className="form-input tabular-nums"
          style={inputStyle}
        />
      </FormField>
      <FormField label="Mortgage Rate (%)" htmlFor="mortgage-rate">
        <input
          id="mortgage-rate"
          type="number"
          min="0"
          step="0.001"
          value={mortgageRate}
          onChange={(e) => onMortgageRateChange(e.target.value)}
          className="form-input tabular-nums"
          style={inputStyle}
        />
      </FormField>
      <FormField label="Term (years)" htmlFor="mortgage-term">
        <input
          id="mortgage-term"
          type="number"
          min="1"
          max="50"
          value={mortgageTermYears}
          onChange={(e) => onMortgageTermYearsChange(e.target.value)}
          className="form-input tabular-nums"
          style={inputStyle}
        />
      </FormField>
      <FormField label="Monthly Payment" htmlFor="monthly-payment">
        <input
          id="monthly-payment"
          type="number"
          min="0"
          step="0.01"
          value={monthlyPayment}
          onChange={(e) => onMonthlyPaymentChange(e.target.value)}
          className="form-input tabular-nums"
          style={inputStyle}
        />
      </FormField>
      <FormField label="Purchase Date" htmlFor="purchase-date">
        <input
          id="purchase-date"
          type="date"
          value={purchaseDate}
          onChange={(e) => onPurchaseDateChange(e.target.value)}
          className="form-input"
          style={inputStyle}
        />
      </FormField>
    </div>
  );
}

function SafeFields({
  companyName,
  onCompanyNameChange,
  ownershipPercentage,
  onOwnershipPercentageChange,
  valuationCap,
  onValuationCapChange,
}: {
  companyName: string;
  onCompanyNameChange: (v: string) => void;
  ownershipPercentage: string;
  onOwnershipPercentageChange: (v: string) => void;
  valuationCap: string;
  onValuationCapChange: (v: string) => void;
}) {
  return (
    <div
      className="flex flex-col gap-3 rounded-md p-3"
      style={{ backgroundColor: "var(--color-background)" }}
    >
      <p className="text-xs font-medium" style={{ color: "var(--color-text-secondary)" }}>
        SAFE Agreement Details
      </p>
      <FormField label="Company Name" htmlFor="company-name">
        <input
          id="company-name"
          type="text"
          value={companyName}
          onChange={(e) => onCompanyNameChange(e.target.value)}
          className="form-input"
          style={inputStyle}
        />
      </FormField>
      <FormField label="Ownership (%)" htmlFor="ownership">
        <input
          id="ownership"
          type="number"
          min="0"
          max="100"
          step="0.01"
          value={ownershipPercentage}
          onChange={(e) => onOwnershipPercentageChange(e.target.value)}
          className="form-input tabular-nums"
          style={inputStyle}
        />
      </FormField>
      <FormField label="Valuation Cap" htmlFor="valuation-cap">
        <input
          id="valuation-cap"
          type="number"
          min="0"
          step="1"
          value={valuationCap}
          onChange={(e) => onValuationCapChange(e.target.value)}
          className="form-input tabular-nums"
          style={inputStyle}
        />
      </FormField>
    </div>
  );
}
