import "server-only";

import { isIP } from "node:net";
import type { NextRequest } from "next/server";
import { getTrustedOrigins } from "@/lib/app-url";

export function getClientIp(req: NextRequest): string {
  const candidates = [
    req.headers.get("x-vercel-forwarded-for"),
    req.headers.get("x-forwarded-for"),
    req.headers.get("x-real-ip"),
  ];
  for (const value of candidates) {
    const candidate = value?.split(",")[0]?.trim();
    if (candidate && isIP(candidate)) return candidate;
  }
  return "unknown";
}

export function hasTrustedOrigin(req: NextRequest): boolean {
  const origin = req.headers.get("origin");
  if (!origin) return false;
  return getTrustedOrigins().includes(origin);
}
