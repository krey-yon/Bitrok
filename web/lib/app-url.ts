/**
 * Public site origin helpers.
 *
 * Production:
 *   https://bitrok.tech  (and/or https://www.bitrok.tech)
 * Relay is separate:
 *   https://api.bitrok.tech
 */

const LOCAL = "http://localhost:3000";
const PROD = "https://bitrok.tech";

/** Canonical public app URL for metadata, redirects, auth. */
export function getAppURL(): string {
  const explicit = process.env.NEXT_PUBLIC_APP_URL?.replace(/\/+$/, "");
  if (explicit && !isLoopback(explicit)) {
    return explicit;
  }

  const betterAuth = process.env.BETTER_AUTH_URL?.replace(/\/+$/, "");
  if (betterAuth && !isLoopback(betterAuth)) {
    return betterAuth;
  }

  // Vercel system env (preview + production deployments)
  const vercel = process.env.VERCEL_URL?.replace(/\/+$/, "");
  if (vercel && !vercel.includes("localhost")) {
    // Prefer production domain over ephemeral *.vercel.app when on production.
    if (process.env.VERCEL_ENV === "production") {
      return PROD;
    }
    return `https://${vercel}`;
  }

  if (process.env.NODE_ENV === "production" || process.env.VERCEL_ENV === "production") {
    return PROD;
  }

  return explicit || LOCAL;
}

/** Auth server base URL (Better Auth). Never localhost on Vercel production. */
export function getAuthBaseURL(): string {
  const raw = process.env.BETTER_AUTH_URL?.replace(/\/+$/, "");
  if (raw && !isLoopback(raw)) {
    return raw;
  }

  // Misconfigured / missing in production → fall back to public app URL.
  if (
    process.env.NODE_ENV === "production" ||
    process.env.VERCEL === "1" ||
    process.env.VERCEL_ENV === "production"
  ) {
    // During `next build` on Vercel, env may still be localhost from a bad
    // dashboard value — never use loopback in that case.
    if (!raw || isLoopback(raw)) {
      return getAppURL();
    }
  }

  return raw || LOCAL;
}

/** Origins that should be allowed for auth cookies / CSRF / OAuth return. */
export function getTrustedOrigins(): string[] {
  const primary = getAuthBaseURL();
  const set = new Set<string>([primary, getAppURL()]);

  try {
    for (const origin of [...set]) {
      const u = new URL(origin);
      if (u.hostname === "bitrok.tech") {
        set.add("https://www.bitrok.tech");
      }
      if (u.hostname === "www.bitrok.tech") {
        set.add("https://bitrok.tech");
      }
    }
  } catch {
    /* ignore */
  }

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
