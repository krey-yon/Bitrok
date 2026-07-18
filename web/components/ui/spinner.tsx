import { Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";

/**
 * Spinner — single loading affordance with accent glow.
 */
export function Spinner({ className }: { className?: string }) {
  return (
    <Loader2
      className={cn(
        "size-4 animate-spin text-accent",
        className,
      )}
      aria-hidden
    />
  );
}
