"use client";

import { useTheme } from "next-themes";
import { Moon, Sun } from "lucide-react";
import { useEffect, useState } from "react";
import { cn } from "@/lib/utils";

/**
 * ThemeToggle — dark/light switch. Renders a single icon button that flips
 * the theme. next-themes is client-only, so resolvedTheme is undefined on
 * the server; we keep aria-label/title/icon stable until mounted to avoid
 * a hydration mismatch on the button attributes.
 */
export function ThemeToggle({ className }: { className?: string }) {
  const { resolvedTheme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => setMounted(true), []);

  const isDark = resolvedTheme === "dark";
  // Stable until mounted so SSR + first client render match exactly.
  const label = mounted
    ? isDark
      ? "Switch to light theme"
      : "Switch to dark theme"
    : "Toggle theme";
  const title = mounted ? (isDark ? "Light" : "Dark") : "Toggle theme";

  return (
    <button
      type="button"
      aria-label={label}
      title={title}
      onClick={() => setTheme(resolvedTheme === "dark" ? "light" : "dark")}
      className={cn(
        "inline-flex h-9 w-9 items-center justify-center rounded-[calc(var(--radius)*0.9)]",
        "text-muted-foreground hover:text-foreground hover:bg-foreground/[0.04]",
        "cursor-pointer transition-colors duration-200 ease-[var(--ease-bitrok)]",
        "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring",
        className,
      )}
    >
      {mounted ? (
        isDark ? (
          <Sun className="size-4" aria-hidden />
        ) : (
          <Moon className="size-4" aria-hidden />
        )
      ) : (
        <span className="size-4" aria-hidden />
      )}
    </button>
  );
}
