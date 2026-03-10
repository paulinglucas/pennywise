import { useEffect, useRef, useId, useCallback, type ReactNode } from "react";
import { X } from "lucide-react";

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
}

function getFocusableElements(container: HTMLElement): HTMLElement[] {
  const selectors = [
    "a[href]",
    "button:not([disabled])",
    "input:not([disabled])",
    "select:not([disabled])",
    "textarea:not([disabled])",
    '[tabindex]:not([tabindex="-1"])',
  ].join(",");
  return Array.from(container.querySelectorAll<HTMLElement>(selectors));
}

export default function Modal({ isOpen, onClose, title, children }: ModalProps) {
  const dialogRef = useRef<HTMLDivElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);
  const titleId = useId();

  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
        return;
      }

      if (event.key !== "Tab" || !dialogRef.current) return;

      const focusable = getFocusableElements(dialogRef.current);
      if (focusable.length === 0) return;

      const first: HTMLElement | undefined = focusable[0];
      const last: HTMLElement | undefined = focusable[focusable.length - 1];
      if (!first || !last) return;

      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    },
    [onClose],
  );

  useEffect(() => {
    if (!isOpen) return;

    previousFocusRef.current = document.activeElement as HTMLElement;
    document.addEventListener("keydown", handleKeyDown);
    document.body.style.overflow = "hidden";

    const timer = setTimeout(() => {
      if (!dialogRef.current) return;
      const focusable = getFocusableElements(dialogRef.current);
      const firstFocusable: HTMLElement | undefined = focusable[0];
      if (firstFocusable) {
        firstFocusable.focus();
      }
    }, 0);

    return () => {
      clearTimeout(timer);
      document.removeEventListener("keydown", handleKeyDown);
      document.body.style.overflow = "";
      if (previousFocusRef.current && typeof previousFocusRef.current.focus === "function") {
        previousFocusRef.current.focus();
      }
    };
  }, [isOpen, handleKeyDown]);

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      data-testid="modal-backdrop"
      onClick={onClose}
      style={{ backgroundColor: "rgba(0, 0, 0, 0.7)" }}
    >
      <div
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        onClick={(event) => event.stopPropagation()}
        className="w-full max-w-lg rounded-lg"
        style={{
          backgroundColor: "var(--color-surface)",
          border: "1px solid var(--color-accent-muted)",
          boxShadow: "var(--glow-lg)",
        }}
      >
        <div
          className="flex items-center justify-between border-b px-6 py-4"
          style={{ borderColor: "var(--color-border)" }}
        >
          <h2
            id={titleId}
            className="text-lg font-semibold"
            style={{ color: "var(--color-text-primary)" }}
          >
            {title}
          </h2>
          <button
            onClick={onClose}
            className="btn-icon rounded-md p-1 transition-all"
            style={{ color: "var(--color-text-secondary)" }}
            aria-label="Close"
          >
            <X size={18} />
          </button>
        </div>
        <div className="px-6 py-4">{children}</div>
      </div>
    </div>
  );
}
