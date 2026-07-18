import * as React from "react";
import { cn } from "@/lib/utils";

export type InputProps = React.InputHTMLAttributes<HTMLInputElement>;

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, ...props }, ref) => (
    <input
      ref={ref}
      type={type ?? "text"}
      className={cn(
        "w-full h-12 rounded-[var(--radius)] border border-input bg-card/70 px-4 text-[15px] text-foreground shadow-[0_1px_0_rgb(255_255_255/4%)_inset]",
        "placeholder:text-muted-foreground/65 transition-[border-color,background-color,box-shadow] duration-200 ease-[var(--ease-bitrok)]",
        "hover:border-foreground/25 focus-visible:border-accent focus-visible:bg-card focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring focus-visible:shadow-[0_0_0_4px_var(--accent-glow)]",
        "disabled:cursor-not-allowed disabled:opacity-50",
        className,
      )}
      {...props}
    />
  ),
);
Input.displayName = "Input";

const Label = React.forwardRef<HTMLLabelElement, React.LabelHTMLAttributes<HTMLLabelElement>>(
  ({ className, ...props }, ref) => (
    <label ref={ref} className={cn("block text-sm font-medium text-foreground", className)} {...props} />
  ),
);
Label.displayName = "Label";

export { Input, Label };
