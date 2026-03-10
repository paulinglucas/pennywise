function ClownCoinIcon({ size = 32 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 32 32"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
      style={{ filter: "drop-shadow(0 0 6px var(--color-accent-muted))" }}
    >
      <circle cx="16" cy="16" r="15" stroke="currentColor" strokeWidth="1.5" opacity="0.6" />
      <circle cx="16" cy="16" r="12" stroke="currentColor" strokeWidth="1" opacity="0.3" />
      <circle cx="16" cy="8" r="3" fill="#ef4444" opacity="0.9" />
      <circle cx="11.5" cy="14" r="2" fill="currentColor" />
      <circle cx="11.5" cy="14" r="0.8" fill="var(--color-background)" />
      <circle cx="20.5" cy="14" r="2" fill="currentColor" />
      <circle cx="20.5" cy="14" r="0.8" fill="var(--color-background)" />
      <path
        d="M9 12.5 Q11.5 11 14 12.5"
        stroke="currentColor"
        strokeWidth="0.8"
        fill="none"
        strokeLinecap="round"
      />
      <path
        d="M18 12.5 Q20.5 11 23 12.5"
        stroke="currentColor"
        strokeWidth="0.8"
        fill="none"
        strokeLinecap="round"
      />
      <circle cx="16" cy="17" r="2" fill="#ef4444" opacity="0.9" />
      <path
        d="M10 21 Q13 25 16 23 Q19 25 22 21"
        stroke="currentColor"
        strokeWidth="1.2"
        fill="none"
        strokeLinecap="round"
      />
      <line x1="12" y1="22.5" x2="12" y2="21.5" stroke="currentColor" strokeWidth="0.6" />
      <line x1="14.5" y1="23.5" x2="14.5" y2="22.5" stroke="currentColor" strokeWidth="0.6" />
      <line x1="17.5" y1="23.5" x2="17.5" y2="22.5" stroke="currentColor" strokeWidth="0.6" />
      <line x1="20" y1="22.5" x2="20" y2="21.5" stroke="currentColor" strokeWidth="0.6" />
      <path d="M7 10 Q5 7 8 7" stroke="currentColor" strokeWidth="0.8" fill="none" opacity="0.5" />
      <path
        d="M25 10 Q27 7 24 7"
        stroke="currentColor"
        strokeWidth="0.8"
        fill="none"
        opacity="0.5"
      />
    </svg>
  );
}

interface BrandLogoProps {
  size?: "sm" | "lg";
}

export default function BrandLogo({ size = "sm" }: BrandLogoProps) {
  const textClass = size === "lg" ? "text-3xl" : "text-2xl";
  const iconSize = size === "lg" ? 40 : 32;

  return (
    <span
      className={`inline-flex items-center gap-2 ${textClass} tracking-wide`}
      style={{
        color: "var(--color-accent)",
        textShadow: "0 0 20px var(--color-accent-muted)",
        fontFamily: "'Pacifico', cursive",
      }}
    >
      Pennywise
      <ClownCoinIcon size={iconSize} />
    </span>
  );
}
