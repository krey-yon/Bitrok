import * as React from "react";
import { cn } from "@/lib/utils";

/**
 * StatusGlyph — the ✓ ✗ … ● ○ vocabulary with subtle glow on active.
 *
 *   success  →  ✓   green   — connected, ok, active
 *   danger   →  ✗   red     — blocked, failed, error
 *   warning  →  …   amber   — pending, held, connecting
 *   active   →  ●   green   — live dot (tunnels, requests)
 *   idle     →  ○   muted   — inactive, off
 */
type GlyphVariant = "success" | "danger" | "warning" | "active" | "idle";

const GLYPHS: Record<GlyphVariant, { char: string; cls: string }> = {
  success: { char: "✓", cls: "text-success" },
  danger: { char: "✗", cls: "text-danger" },
  warning: { char: "…", cls: "text-warning" },
  active: { char: "●", cls: "text-success" },
  idle: { char: "○", cls: "text-muted-foreground" },
};

export function StatusGlyph({
  variant,
  pulse = false,
  className,
}: {
  variant: GlyphVariant;
  pulse?: boolean;
  className?: string;
}) {
  const { char, cls } = GLYPHS[variant];
  return (
    <span
      aria-hidden
      className={cn(
        "inline-flex items-center justify-center font-mono leading-none tabular-nums",
        cls,
        pulse && variant === "active" && "animate-pulse-dot",
        className,
      )}
    >
      {char}
    </span>
  );
}

export type { GlyphVariant };
