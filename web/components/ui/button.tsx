import * as React from "react";
import { cn } from "@/lib/utils";

/**
 * Button — the terminal-cosmic action primitive.
 *
 * The signature is text + arrow for the primary action (conifer DNA),
 * a hairline border for secondary, and a quiet ghost for tertiary/nav.
 * No filled pills, no gradient. Subtle radius, not round. Mono variant
 * for in-terminal actions (renders in Geist Mono with a leading glyph).
 *
 * Variants:
 *   primary — foreground fill, the default action (still solid so it
 *             reads as the CTA, but squared + arrow, not a glossy pill)
 *   accent  — amber fill, the brand CTA (landing/marketing only)
 *   ghost   — hairline border, secondary actions
 *   subtle  — quiet hover-bg, tertiary / nav
 *   mono    — terminal-styled: mono font, leading `›` glyph, hairline
 *   danger  — red fill, destructive confirm
 *
 * Sizes: sm | md | lg | icon
 *
 * Pass `arrow` to append a `→` that nudges right on hover.
 */
const variants = {
  primary:
    "bg-foreground text-background hover:opacity-90 active:opacity-100",
  accent:
    "bg-accent text-accent-foreground hover:bg-accent-light",
  ghost:
    "border border-hairline bg-transparent hover:bg-foreground/[0.04] text-foreground",
  subtle:
    "bg-transparent hover:bg-foreground/[0.04] text-muted-foreground hover:text-foreground",
  mono:
    "border border-hairline bg-transparent hover:border-border hover:bg-foreground/[0.03] text-foreground font-mono",
  danger:
    "bg-danger text-danger-foreground hover:opacity-90",
} as const;

const sizes = {
  sm: "h-8 px-3.5 text-xs gap-1.5 rounded-[calc(var(--radius)*0.85)]",
  md: "h-10 px-5 text-sm gap-2 rounded-[var(--radius)]",
  lg: "h-12 px-6 text-sm gap-2 rounded-[var(--radius)]",
  icon: "h-9 w-9 rounded-[calc(var(--radius)*0.85)]",
} as const;

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: keyof typeof variants;
  size?: keyof typeof sizes;
  arrow?: boolean;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    { className, variant = "primary", size = "md", arrow = false, type, children, ...props },
    ref,
  ) => (
    <button
      ref={ref}
      type={type ?? "button"}
      className={cn(
        "group/btn inline-flex items-center justify-center font-medium tracking-tight",
        "cursor-pointer select-none whitespace-nowrap",
        "transition-all duration-200 ease-[var(--ease-bitrok)]",
        "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring",
        "disabled:pointer-events-none disabled:opacity-50",
        variants[variant],
        sizes[size],
        className,
      )}
      {...props}
    >
      {variant === "mono" && (
        <span aria-hidden className="text-accent leading-none">›</span>
      )}
      {children}
      {arrow && (
        <span
          aria-hidden
          className="transition-transform duration-200 ease-[var(--ease-bitrok)] group-hover/btn:translate-x-0.5"
        >
          →
        </span>
      )}
    </button>
  ),
);
Button.displayName = "Button";
