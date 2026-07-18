import { cn } from "@/lib/utils";

/**
 * Logo — the Bitrok wordmark. Lowercase bold Inter with a small amber dot
 * as the mark. Minimal, no terminal glyph. Sized via className; inherits
 * text color.
 */
export function Logo({
  className,
  withMark = true,
}: {
  className?: string;
  withMark?: boolean;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 font-bold tracking-tight lowercase",
        className,
      )}
    >
      {withMark && (
        <span
          className="inline-block size-1.5 rounded-full bg-accent"
          aria-hidden
        />
      )}
      <span>bitrok</span>
    </span>
  );
}
