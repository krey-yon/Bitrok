import { cn } from "@/lib/utils";

export function Logo({ className, withMark = true }: { className?: string; withMark?: boolean }) {
  return (
    <span className={cn("inline-flex items-center gap-2.5 font-display font-bold tracking-[-0.04em] lowercase", className)} translate="no">
      {withMark && (
        <svg className="size-6 overflow-visible" viewBox="0 0 24 24" aria-hidden>
          <path d="M4 4v16M4 12h8c4.4 0 8-3.6 8-8M12 12c4.4 0 8 3.6 8 8" fill="none" stroke="var(--foreground)" strokeWidth="2.2" strokeLinecap="round" />
          <circle cx="4" cy="12" r="2.2" fill="var(--accent)" />
          <circle cx="20" cy="4" r="2.2" fill="var(--secondary)" />
          <circle cx="20" cy="20" r="2.2" fill="var(--accent)" />
        </svg>
      )}
      <span>bitrok</span>
    </span>
  );
}
