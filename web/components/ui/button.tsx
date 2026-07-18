import * as React from "react";
import { cn } from "@/lib/utils";

export const buttonStyles = {
  primary: "bg-foreground text-background hover:bg-foreground/88 shadow-[0_1px_0_rgb(255_255_255/15%)_inset]",
  accent: "bg-accent text-accent-foreground hover:bg-accent-light shadow-[0_14px_36px_var(--accent-glow)]",
  ghost: "border border-border bg-card/55 text-foreground hover:bg-card hover:border-foreground/25",
  subtle: "bg-transparent text-muted-foreground hover:bg-foreground/[0.06] hover:text-foreground",
  mono: "border border-hairline bg-card/50 text-foreground font-mono hover:border-accent/50 hover:text-accent-soft",
  danger: "bg-danger text-danger-foreground hover:brightness-110",
} as const;

export const buttonSizes = {
  sm: "h-9 px-3.5 text-xs gap-1.5 rounded-[var(--radius-sm)]",
  md: "h-11 px-5 text-sm gap-2 rounded-[var(--radius)]",
  lg: "h-12 px-6 text-sm gap-2 rounded-[var(--radius)]",
  icon: "size-11 rounded-[var(--radius)]",
} as const;

export function buttonClassName({ variant = "primary", size = "md", className }: { variant?: keyof typeof buttonStyles; size?: keyof typeof buttonSizes; className?: string } = {}) {
  return cn("group/btn inline-flex items-center justify-center whitespace-nowrap font-semibold tracking-[-0.01em] cursor-pointer select-none transition-[color,background-color,border-color,box-shadow,transform,filter] duration-200 ease-[var(--ease-bitrok)] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:pointer-events-none disabled:opacity-45 active:translate-y-px", buttonStyles[variant], buttonSizes[size], className);
}

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: keyof typeof buttonStyles;
  size?: keyof typeof buttonSizes;
  arrow?: boolean;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "primary", size = "md", arrow = false, type, children, ...props }, ref) => (
    <button ref={ref} type={type ?? "button"} className={buttonClassName({ variant, size, className })} {...props}>
      {variant === "mono" && <span aria-hidden className="text-accent leading-none">›</span>}
      {children}
      {arrow && <span aria-hidden className="transition-transform duration-200 ease-[var(--ease-bitrok)] group-hover/btn:translate-x-1">→</span>}
    </button>
  ),
);
Button.displayName = "Button";
