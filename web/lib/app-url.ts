/**
 * Public site origin helpers.
 *
 * Production:
 *   https://bitrok.tech  (and/or https://www.bitrok.tech)
 * Relay is separate:
 *   https://api.bitrok.tech
 */

const LOCAL = "http://localhost:3000";

/** Canonical public app URL for metadata, redirects, auth. */
export function getAppURL(): string {
  const explicit = process.env.NEXT_PUBLIC_APP_URL?.replace(/\/+$/, "");
  if (explicit && !isLoopback(explicit)) {
    return explicit;
  }

  // Vercel system env (no protocol)
  const vercel = process.env.VERCEL_URL?.replace(/\/+$/, "");
  if (vercel && !vercel.includes("localhost")) {
    return `https://${vercel}`;
  }

  const betterAuth = process.env.BETTER_AUTH_URL?.replace(/\/+$/, "");
  if (betterAuth && !isLoopback(betterAuth)) {
    return betterAuth;
  }

  return explicit || LOCAL;
}

/** Origins that should be allowed for auth cookies / CSRF / OAuth return. */
export function getTrustedOrigins(): string[] {
  const primary = getAppURL();
  const set = new Set<string>([primary]);

  try {
    const u = new URL(primary);
    if (u.hostname === "bitrok.tech") {
      set.add("https://www.bitrok.tech");
    }
    if (u.hostname === "www.bitrok.tech") {
      set.add("https://bitrok.tech");
    }
  } catch {
    /* ignore */
  }

  // Always allow local dev when not in production builds.
  if (process.env.NODE_ENV !== "production") {
    set.add(LOCAL);
  }

  return [...set];
}

function isLoopback(url: string): boolean {
  return (
    url.includes("localhost") ||
    url.includes("127.0.0.1") ||
    url.includes("[::1]")
  );
}
