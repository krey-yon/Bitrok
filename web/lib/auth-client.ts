"use client";

import { createAuthClient } from "better-auth/react";

/**
 * Auth client base URL — MUST be same-origin in the browser.
 *
 * The live failure mode was:
 *   origin: https://www.bitrok.tech
 *   fetch:  http://localhost:3000/api/auth/...
 * → blocked by Private Network Access / CORS.
 *
 * That happens when NEXT_PUBLIC_APP_URL is missing or set to localhost in the
 * production build. Prefer the current page origin whenever we have a window.
 */
function appBaseURL(): string {
  if (typeof window !== "undefined") {
    return window.location.origin;
  }

  const fromEnv = process.env.NEXT_PUBLIC_APP_URL?.replace(/\/+$/, "");
  if (
    fromEnv &&
    !fromEnv.includes("localhost") &&
    !fromEnv.includes("127.0.0.1")
  ) {
    return fromEnv;
  }

  return fromEnv || "http://localhost:3000";
}

export const authClient = createAuthClient({
  baseURL: appBaseURL(),
});

export const { signIn, signUp, signOut, useSession } = authClient;
