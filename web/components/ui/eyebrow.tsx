import * as React from "react";
import { cn } from "@/lib/utils";

/**
 * Eyebrow — the small uppercase label that leads every section.
 *
 * Minimal YC rhythm: a quiet sans uppercase tag with a leading amber dot,
 * wide tracking, muted. Sits above the big statement.
 */
export function Eyebrow({
  children,
  ornament = "dot",
  className,
}: {
  children: React.ReactNode;
  /**
   * "dot" renders a round amber dot (the new minimal mark). The char
   * options ("·", "✦", "+", "›") render that glyph in accent and are kept
   * for backward compatibility with existing pages. Pass `false` for none.
   */
  ornament?: "dot" | "·" | "✦" | "+" | "›" | false;
  className?: string;
}) {
  return (
    <p
      className={cn(
        "inline-flex items-center gap-2 text-xs uppercase tracking-[0.2em] text-muted-foreground",
        className,
      )}
    >
      {ornament !== false &&
        (ornament === "dot" ? (
          <span
            aria-hidden
            className="inline-block size-1.5 rounded-full bg-accent"
          />
        ) : (
          <span aria-hidden className="text-accent">
            {ornament}
          </span>
        ))}
      {children}
    </p>
  );
}
