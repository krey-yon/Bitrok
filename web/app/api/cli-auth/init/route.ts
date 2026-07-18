import { prisma } from "@/lib/prisma";
import { rateLimit, getRateLimitHeaders } from "@/lib/rate-limit";
import { NextRequest, NextResponse } from "next/server";
import crypto from "crypto";

function getClientIp(req: NextRequest): string {
  const forwarded = req.headers.get("x-forwarded-for");
  return forwarded ? forwarded.split(",")[0].trim() : "unknown";
}

// Validate callback URL to prevent open redirect
function isValidCallbackUrl(url: string): boolean {
  try {
    const parsed = new URL(url);
    // Allow only http/https protocols
    if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
      return false;
    }
    // Optionally: restrict to specific domains
    // const allowedDomains = ["localhost", "bitrok.dev"];
    // return allowedDomains.includes(parsed.hostname);
    return true;
  } catch {
    return false;
  }
}

export async function POST(req: NextRequest) {
  const clientIp = getClientIp(req);
  const rateLimitResult = rateLimit(`cli-auth:init:${clientIp}`, {
    windowMs: 60 * 1000,
    maxRequests: 5,
  });

  const headers = getRateLimitHeaders(rateLimitResult);

  if (!rateLimitResult.success) {
    return NextResponse.json(
      { error: "Rate limit exceeded. Please try again later." },
      { status: 429, headers }
    );
  }

  try {
    const { callback } = await req.json();

    if (!callback || typeof callback !== "string") {
      return NextResponse.json(
        { error: "callback URL is required" },
        { status: 400, headers }
      );
    }

    if (!isValidCallbackUrl(callback)) {
      return NextResponse.json(
        { error: "Invalid callback URL" },
        { status: 400, headers }
      );
    }

    const state = crypto.randomBytes(32).toString("hex");

    await prisma.cliAuthRequest.create({
      data: {
        state,
        status: "pending",
        expiresAt: new Date(Date.now() + 10 * 60 * 1000), // 10 minutes
      },
    });

    const authUrl = `/cli-auth?state=${state}&callback=${encodeURIComponent(callback)}`;

    return NextResponse.json({ state, authUrl }, { headers });
  } catch (error) {
    console.error("cli-auth init error:", error);
    return NextResponse.json(
      { error: "Failed to initialize auth request" },
      { status: 500, headers }
    );
  }
}
