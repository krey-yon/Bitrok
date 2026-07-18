import * as React from "react";
import { cn } from "@/lib/utils";

/**
 * TerminalPanel — the faux terminal window.
 *
 * A flat, hairline-bordered surface with a mono header strip (three dots
 * + optional path/title) and a mono body. This is the onecli.sh hero
 * device and the home for the gateway-log aesthetic across the app.
 *
 * Use `lines` for static content, or children for composed bodies.
 * `tone="dim"` drops the body to a slightly recessed bg for contrast
 * against the header.
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
        "terminal overflow-hidden",
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
 * TerminalLine — one row in a gateway log. Mono, optional status glyph,
 * optional dim label prefix. Renders as a grid row for crisp alignment.
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
