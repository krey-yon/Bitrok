import * as React from "react";
import { cn } from "@/lib/utils";

/**
 * Badge — compact mono status/label pill. Used for tunnel status, log
 * methods, feature tags. Solid status variants share the status palette
 * and can carry a leading status glyph (✓ ✗ …).
 *
 * Terminal-cosmic: mono font, hairline border, translucent tinted bg.
 * Squared, not fully rounded — reads as a terminal tag, not a bubble.
 */
const variants = {
  default:
    "border border-hairline bg-transparent text-muted-foreground",
  accent:
    "border border-accent/30 bg-accent/10 text-accent",
  success:
    "border border-success/30 bg-success/10 text-success",
  warning:
    "border border-warning/30 bg-warning/10 text-warning",
  danger:
    "border border-danger/30 bg-danger/10 text-danger",
} as const;

const glyphs = {
  default: null,
  accent: "·",
  success: "✓",
  warning: "…",
  danger: "✗",
} as const;

export interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: keyof typeof variants;
  glyph?: boolean;
}

export function Badge({
  className,
  variant = "default",
  glyph = false,
  children,
  ...props
}: BadgeProps) {
  const g = glyph ? glyphs[variant] : null;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-[calc(var(--radius)*0.5)] px-2 py-0.5",
        "font-mono text-[0.7rem] leading-none tracking-tight whitespace-nowrap",
        "transition-colors duration-200",
        variants[variant],
        className,
      )}
      {...props}
    >
      {g && <span aria-hidden>{g}</span>}
      {children}
    </span>
  );
}

/**
 * StatusDot — the ●/○ indicator shared by tunnels everywhere.
 * Matches the CLI's status-pill helper (cli/internal/ui/status.go).
 * Active dot pulses gently (terminal heartbeat).
 */
export function StatusDot({
  active,
  className,
}: {
  active: boolean;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "inline-block w-1.5 h-1.5 rounded-full shrink-0",
        active
          ? "bg-success animate-pulse-dot"
          : "bg-muted-foreground/40",
        className,
      )}
      aria-hidden
    />
  );
}
