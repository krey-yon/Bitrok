import * as React from "react";
import { cn } from "@/lib/utils";

/**
 * TerminalPanel — the faux terminal window with scanline effect.
 *
 * Glassmorphism surface with animated scanline overlay. The header has
 * three dots + optional path/title. Body is mono.
 */
export function TerminalPanel({
  title,
  children,
  className,
  bodyClassName,
  tone = "default",
}: {
  title?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
  bodyClassName?: string;
  tone?: "default" | "dim";
}) {
  return (
    <div
      className={cn(
        "terminal terminal-scanlines overflow-hidden",
        className,
      )}
    >
      <div className="terminal-header flex items-center gap-2 px-3.5 h-9">
        <span className="flex items-center gap-1.5" aria-hidden>
          <span className="size-2 rounded-full bg-danger/70" />
          <span className="size-2 rounded-full bg-warning/70" />
          <span className="size-2 rounded-full bg-success/70" />
        </span>
        {title && (
          <span className="ml-2 text-xs font-mono text-muted-foreground truncate">
            {title}
          </span>
        )}
      </div>
      <div
        className={cn(
          "px-4 py-3.5 font-mono text-sm leading-relaxed",
          tone === "dim" && "bg-background/40",
          bodyClassName,
        )}
      >
        {children}
      </div>
    </div>
  );
}

/**
 * TerminalLine — one row in a gateway log.
 */
export function TerminalLine({
  label,
  children,
  status,
  className,
}: {
  label?: React.ReactNode;
  children: React.ReactNode;
  status?: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("flex items-start gap-3", className)}>
      {status && (
        <span className="select-none w-4 shrink-0 text-center">{status}</span>
      )}
      {label && (
        <span className="text-muted-foreground select-none shrink-0">{label}</span>
      )}
      <span className="min-w-0 break-all">{children}</span>
    </div>
  );
}
