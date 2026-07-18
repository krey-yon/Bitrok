import * as React from "react";
import { cn } from "@/lib/utils";

/**
 * Input — underline-style field, terminal-cosmic.
 *
 * Borderless box; the bottom hairline brightens to the foreground on
 * focus. Mono placeholders so hints read as terminal prompts. Optional
 * <Label> primitive renders the small mono uppercase eyebrow.
 */
export type InputProps = React.InputHTMLAttributes<HTMLInputElement>;

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, ...props }, ref) => (
    <input
      ref={ref}
      type={type ?? "text"}
      className={cn(
        "w-full bg-transparent border-0 border-b border-hairline",
        "py-2.5 text-sm text-foreground",
        "placeholder:text-muted placeholder:font-mono placeholder:text-xs",
        "outline-none transition-colors duration-200 ease-[var(--ease-bitrok)]",
        "focus:border-foreground",
        "disabled:cursor-not-allowed disabled:opacity-50",
        className,
      )}
      {...props}
    />
  ),
);
Input.displayName = "Input";

const Label = React.forwardRef<
  HTMLLabelElement,
  React.LabelHTMLAttributes<HTMLLabelElement>
>(({ className, ...props }, ref) => (
  <label
    ref={ref}
    className={cn(
      "font-mono text-xs uppercase tracking-[0.22em] text-muted-foreground",
      className,
    )}
    {...props}
  />
));
Label.displayName = "Label";

export { Input, Label };
