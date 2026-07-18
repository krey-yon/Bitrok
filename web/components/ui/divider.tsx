import * as React from "react";
import { cn } from "@/lib/utils";

/**
 * Divider — a simple hairline rule with a single centered dot.
 */
export function Divider({
  className,
}: {
  className?: string;
}) {
  return (
    <div
      role="separator"
      aria-orientation="horizontal"
      className={cn("divider-dot py-2", className)}
    >
      <span aria-hidden className="text-xs leading-none text-accent/50">·</span>
    </div>
  );
}
