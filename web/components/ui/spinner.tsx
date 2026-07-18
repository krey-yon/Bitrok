import { Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";

/**
 * Spinner — single loading affordance. Uses lucide Loader2 (the only
 * icon lib in deps). Pair with "animate-spin". Muted, terminal-quiet.
 */
export function Spinner({ className }: { className?: string }) {
  return (
    <Loader2
      className={cn("size-4 animate-spin text-muted-foreground", className)}
      aria-hidden
    />
  );
}
