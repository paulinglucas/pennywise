import { useState, useRef, useEffect } from "react";
import { useCategories } from "@/hooks/useCategories";

interface CategoryComboboxProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  id?: string;
}

export default function CategoryCombobox({
  value,
  onChange,
  placeholder = "Category",
  id,
}: CategoryComboboxProps) {
  const { data: categories = [] } = useCategories();
  const [open, setOpen] = useState(false);
  const [highlightIndex, setHighlightIndex] = useState(-1);
  const containerRef = useRef<HTMLDivElement>(null);

  const filtered = categories.filter((cat) => cat.toLowerCase().includes(value.toLowerCase()));

  const exactMatch = categories.some((cat) => cat.toLowerCase() === value.toLowerCase());
  const showAddOption = value.trim() !== "" && !exactMatch;

  const options = showAddOption ? [...filtered, `__add__${value.trim()}`] : filtered;

  useEffect(() => {
    function handleClickOutside(event: globalThis.MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as globalThis.Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  function selectOption(option: string) {
    if (option.startsWith("__add__")) {
      onChange(option.replace("__add__", ""));
    } else {
      onChange(option);
    }
    setOpen(false);
    setHighlightIndex(-1);
  }

  function handleKeyDown(event: React.KeyboardEvent) {
    if (!open) {
      if (event.key === "ArrowDown" || event.key === "ArrowUp") {
        setOpen(true);
        event.preventDefault();
      }
      return;
    }

    if (event.key === "ArrowDown") {
      event.preventDefault();
      setHighlightIndex((prev) => (prev + 1) % options.length);
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      setHighlightIndex((prev) => (prev <= 0 ? options.length - 1 : prev - 1));
    } else if (event.key === "Enter" && highlightIndex >= 0 && highlightIndex < options.length) {
      event.preventDefault();
      selectOption(options[highlightIndex]!);
    } else if (event.key === "Escape") {
      setOpen(false);
      setHighlightIndex(-1);
    }
  }

  return (
    <div ref={containerRef} className="relative flex-1" style={{ minWidth: 0 }}>
      <input
        id={id}
        role="combobox"
        type="text"
        value={value}
        placeholder={placeholder}
        aria-expanded={open}
        onChange={(e) => {
          onChange(e.target.value);
          setOpen(true);
          setHighlightIndex(-1);
        }}
        onFocus={() => setOpen(true)}
        onKeyDown={handleKeyDown}
        className="form-input w-full"
        style={{
          backgroundColor: "var(--color-background)",
          borderColor: "var(--color-border)",
          color: "var(--color-text-primary)",
        }}
      />
      {open && options.length > 0 && (
        <ul
          className="absolute z-50 mt-1 max-h-48 w-full overflow-auto rounded-md py-1"
          style={{
            backgroundColor: "var(--color-surface)",
            border: "1px solid var(--color-border)",
          }}
        >
          {options.map((option, index) => {
            const isAdd = option.startsWith("__add__");
            const label = isAdd ? `Add "${option.replace("__add__", "")}"` : option;
            return (
              <li
                key={option}
                role="option"
                aria-selected={index === highlightIndex}
                onMouseDown={(e) => e.preventDefault()}
                onClick={() => selectOption(option)}
                onMouseEnter={() => setHighlightIndex(index)}
                className="cursor-pointer px-3 py-1.5 text-sm"
                style={{
                  backgroundColor: index === highlightIndex ? "var(--color-border)" : "transparent",
                  color: isAdd ? "var(--color-accent)" : "var(--color-text-primary)",
                }}
              >
                {label}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
