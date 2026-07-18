import * as React from "react";
import { cn } from "@/lib/utils";

/**
 * Card — the Quantum Grid surface.
 *
 * Glassmorphism: translucent bg with backdrop blur, subtle gradient border,
 * and a soft glow on hover. The hover is a border brighten + glow, not a shadow.
 *
 * variant="terminal" renders the faux-window chrome with scanline effect.
 */
type CardProps = React.HTMLAttributes<HTMLDivElement> & {
  variant?: "default" | "terminal" | "glass";
  terminalTitle?: string;
};

const Card = React.forwardRef<HTMLDivElement, CardProps>(
  ({ className, variant = "default", terminalTitle, children, ...props }, ref) => {
    if (variant === "terminal") {
      return (
        <div
          ref={ref}
          className={cn(
            "terminal terminal-scanlines overflow-hidden text-card-foreground",
                        "transition-[border-color,background-color,box-shadow,transform] duration-200 ease-[var(--ease-bitrok)]",
            className,
          )}
          {...props}
        >
          <div className="terminal-header flex items-center gap-2 px-3.5 h-9">
            <span className="flex items-center gap-1.5" aria-hidden>
              <span className="size-2 rounded-full bg-danger/70" />
              <span className="size-2 rounded-full bg-warning/70" />
              <span className="size-2 rounded-full bg-success/70" />
            </span>
            {terminalTitle && (
              <span className="ml-2 text-xs font-mono text-muted-foreground truncate">
                {terminalTitle}
              </span>
            )}
          </div>
          {children}
        </div>
      );
    }
    if (variant === "glass") {
      return (
        <div
          ref={ref}
          className={cn(
            "glass rounded-card text-card-foreground",
                        "transition-[border-color,background-color,box-shadow,transform] duration-200 ease-[var(--ease-bitrok)]",
            "hover:border-accent/20 hover:glow-sm",
            className,
          )}
          {...props}
        />
      );
    }
    return (
      <div
        ref={ref}
        className={cn(
          "border border-hairline rounded-card bg-card text-card-foreground",
                    "transition-[border-color,background-color,box-shadow,transform] duration-200 ease-[var(--ease-bitrok)]",
          "hover:border-border hover:shadow-[0_0_20px_var(--accent-glow)]",
          className,
        )}
        {...props}
      />
    );
  },
);
Card.displayName = "Card";

const CardHeader = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn("flex flex-col gap-1.5 p-6", className)}
    {...props}
  />
));
CardHeader.displayName = "CardHeader";

const CardTitle = React.forwardRef<
  HTMLHeadingElement,
  React.HTMLAttributes<HTMLHeadingElement>
>(({ className, ...props }, ref) => (
  <h3
    ref={ref}
    className={cn(
      "text-base font-semibold leading-tight tracking-tight font-display",
      className,
    )}
    {...props}
  />
));
CardTitle.displayName = "CardTitle";

const CardDescription = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLParagraphElement>
>(({ className, ...props }, ref) => (
  <p
    ref={ref}
    className={cn("text-sm text-muted-foreground leading-relaxed", className)}
    {...props}
  />
));
CardDescription.displayName = "CardDescription";

const CardContent = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div ref={ref} className={cn("p-6 pt-0", className)} {...props} />
));
CardContent.displayName = "CardContent";

const CardFooter = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn("flex items-center p-6 pt-0", className)}
    {...props}
  />
));
CardFooter.displayName = "CardFooter";

export { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter };
