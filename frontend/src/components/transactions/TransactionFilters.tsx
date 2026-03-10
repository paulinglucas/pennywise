import { Search } from "lucide-react";
import type { AccountResponse, TransactionFilters as Filters, TransactionType } from "@/api/client";
import { useCategories } from "@/hooks/useCategories";

interface TransactionFiltersProps {
  filters: Filters;
  accounts: AccountResponse[];
  onFiltersChange: (filters: Filters) => void;
}

const selectStyle = {
  backgroundColor: "var(--color-background)",
  borderColor: "var(--color-border)",
  color: "var(--color-text-primary)",
};

export default function TransactionFilters({
  filters,
  accounts,
  onFiltersChange,
}: TransactionFiltersProps) {
  const { data: categories = [] } = useCategories();

  function updateFilter(key: keyof Filters, value: string | number | undefined) {
    onFiltersChange({ ...filters, [key]: value || undefined, page: 1 });
  }

  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-end">
      <div className="relative flex-1 sm:min-w-48">
        <Search
          size={16}
          className="absolute left-3 top-1/2 -translate-y-1/2"
          style={{ color: "var(--color-text-secondary)" }}
        />
        <input
          type="text"
          value={filters.search ?? ""}
          onChange={(event) => updateFilter("search", event.target.value)}
          placeholder="Search transactions..."
          aria-label="Search transactions"
          className="form-input w-full rounded-md border py-2 pl-9 pr-3 text-sm transition-shadow"
          style={selectStyle}
        />
      </div>

      <FilterSelect
        label="Account"
        value={filters.account_id ?? ""}
        onChange={(val) => updateFilter("account_id", val)}
      >
        <option value="">All accounts</option>
        {accounts.map((account) => (
          <option key={account.id} value={account.id}>
            {account.name}
          </option>
        ))}
      </FilterSelect>

      <FilterSelect
        label="Type"
        value={filters.type ?? ""}
        onChange={(val) => updateFilter("type", val as TransactionType)}
      >
        <option value="">All types</option>
        <option value="expense">Expense</option>
        <option value="deposit">Deposit</option>
        <option value="transfer">Transfer</option>
      </FilterSelect>

      <FilterSelect
        label="Category"
        value={filters.category ?? ""}
        onChange={(val) => updateFilter("category", val)}
      >
        <option value="">All categories</option>
        {categories.map((cat) => (
          <option key={cat} value={cat}>
            {cat}
          </option>
        ))}
      </FilterSelect>

      <FilterInput
        label="From"
        type="date"
        value={filters.date_from ?? ""}
        onChange={(val) => updateFilter("date_from", val)}
      />

      <FilterInput
        label="To"
        type="date"
        value={filters.date_to ?? ""}
        onChange={(val) => updateFilter("date_to", val)}
      />
    </div>
  );
}

function FilterSelect({
  label,
  value,
  onChange,
  children,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  children: React.ReactNode;
}) {
  const id = `filter-${label.toLowerCase()}`;
  return (
    <div className="flex flex-col gap-1">
      <label
        htmlFor={id}
        className="text-xs font-medium"
        style={{ color: "var(--color-text-secondary)" }}
      >
        {label}
      </label>
      <select
        id={id}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className="form-input rounded-md border px-3 py-2 text-sm transition-shadow"
        style={selectStyle}
      >
        {children}
      </select>
    </div>
  );
}

function FilterInput({
  label,
  value,
  placeholder,
  type = "text",
  onChange,
}: {
  label: string;
  value: string;
  placeholder?: string;
  type?: string;
  onChange: (value: string) => void;
}) {
  const id = `filter-${label.toLowerCase()}`;
  return (
    <div className="flex flex-col gap-1">
      <label
        htmlFor={id}
        className="text-xs font-medium"
        style={{ color: "var(--color-text-secondary)" }}
      >
        {label}
      </label>
      <input
        id={id}
        type={type}
        value={value}
        placeholder={placeholder}
        onChange={(event) => onChange(event.target.value)}
        className="form-input rounded-md border px-3 py-2 text-sm transition-shadow"
        style={selectStyle}
      />
    </div>
  );
}
