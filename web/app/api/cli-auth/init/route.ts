import { prisma } from "@/lib/prisma";
import { rateLimit, getRateLimitHeaders } from "@/lib/rate-limit";
import { getClientIp } from "@/lib/request-security";
import { NextRequest, NextResponse } from "next/server";
import crypto from "crypto";

// Validate callback URL to prevent open redirect
function isValidCallbackUrl(url: string): boolean {
  try {
    const parsed = new URL(url);
    const loopback =
      parsed.hostname === "localhost" ||
      parsed.hostname === "127.0.0.1" ||
      parsed.hostname === "[::1]";
    return (
      parsed.protocol === "http:" &&
      loopback &&
      parsed.port !== "" &&
      parsed.username === "" &&
      parsed.password === "" &&
      parsed.hash === ""
    );
  } catch {
    return false;
  }
}

export async function POST(req: NextRequest) {
  const clientIp = getClientIp(req);
  const rateLimitResult = await rateLimit(`cli-auth:init:${clientIp}`, {
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
    await prisma.cliAuthRequest.deleteMany({ where: { expiresAt: { lt: new Date() } } });

    await prisma.cliAuthRequest.create({
      data: {
        state,
        status: "pending",
        callbackUrl: callback,
        expiresAt: new Date(Date.now() + 10 * 60 * 1000), // 10 minutes
      },
    });

    const authUrl = `/cli-auth?state=${state}`;

    return NextResponse.json({ state, authUrl }, { headers });
  } catch (error) {
    console.error("cli-auth init error:", error);
    return NextResponse.json(
      { error: "Failed to initialize auth request" },
      { status: 500, headers }
    );
  }
}
