"use client";

import { ThemeProvider as NextThemesProvider } from "next-themes";
import type { ComponentProps } from "react";

/**
 * ThemeProvider — class-based dark/light via next-themes.
 * Light is the default; users can switch to dark and keep their preference.
 * Replaces the previous inert theme class + prefers-color-scheme approach,
 * which offered no user override.
 */
export function ThemeProvider({
  children,
  ...props
}: ComponentProps<typeof NextThemesProvider>) {
  return (
    <NextThemesProvider
      attribute="class"
      defaultTheme="light"
      enableSystem={false}
      disableTransitionOnChange
      {...props}
    >
      {children}
    </NextThemesProvider>
  );
}
