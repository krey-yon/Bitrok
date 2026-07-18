import * as React from "react";
import { cn } from "@/lib/utils";

/**
 * Eyebrow — the small uppercase label that leads every section.
 *
 * Quantum Grid: a mono uppercase tag with a leading glowing accent dot,
 * wide tracking, muted. Sits above the big statement.
 */
export function Eyebrow({
  children,
  ornament = "dot",
  className,
}: {
  children: React.ReactNode;
  ornament?: "dot" | "·" | "✦" | "+" | "›" | false;
  className?: string;
}) {
  return (
    <p
      className={cn(
        "inline-flex items-center gap-2 font-mono text-xs uppercase tracking-[0.2em] text-muted-foreground",
        className,
      )}
    >
      {ornament !== false &&
        (ornament === "dot" ? (
          <span className="relative inline-flex" aria-hidden>
            <span className="inline-block size-1.5 rounded-full bg-accent" />
            <span className="absolute inset-0 inline-block size-1.5 rounded-full bg-accent blur-[2px] opacity-50" />
          </span>
        ) : (
          <span aria-hidden className="text-accent">
            {ornament}
          </span>
        ))}
      {children}
    </p>
  );
}
