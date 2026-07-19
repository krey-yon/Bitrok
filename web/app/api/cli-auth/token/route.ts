import { auth } from "@/lib/auth";
import { mintServerToken } from "@/lib/jwt";
import { prisma } from "@/lib/prisma";
import { rateLimit, getRateLimitHeaders } from "@/lib/rate-limit";
import { getTrustedOrigins } from "@/lib/app-url";
import { NextRequest, NextResponse } from "next/server";

function getClientIp(req: NextRequest): string {
  const forwarded = req.headers.get("x-forwarded-for");
  return forwarded ? forwarded.split(",")[0].trim() : "unknown";
}

// Validate Origin header for CSRF protection (www + apex + local).
function validateOrigin(req: NextRequest): boolean {
  const origin = req.headers.get("origin");
  if (!origin) return true;
  return getTrustedOrigins().includes(origin);
}

export async function POST(req: NextRequest) {
  const clientIp = getClientIp(req);
  const rateLimitResult = rateLimit(`cli-auth:token:${clientIp}`, {
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

  // CSRF protection
  if (!validateOrigin(req)) {
    return NextResponse.json(
      { error: "Invalid origin" },
      { status: 403, headers }
    );
  }

  try {
    const session = await auth.api.getSession({ headers: req.headers });
    if (!session) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401, headers });
    }

    const { state } = await req.json();

    if (!state || typeof state !== "string" || !/^[a-f0-9]{64}$/.test(state)) {
      return NextResponse.json(
        { error: "Invalid state format" },
        { status: 400, headers }
      );
    }

    // Validate state exists and is pending
    const authReq = await prisma.cliAuthRequest.findUnique({
      where: { state, status: "pending" },
    });

    if (!authReq) {
      return NextResponse.json(
        { error: "Invalid or expired state" },
        { status: 400, headers }
      );
    }

    if (authReq.expiresAt < new Date()) {
      return NextResponse.json(
        { error: "State expired" },
        { status: 400, headers }
      );
    }

    const jwtSecret = process.env.BITROK_JWT_SECRET;
    if (!jwtSecret) {
      console.error("BITROK_JWT_SECRET not configured");
      return NextResponse.json(
        { error: "Server configuration error" },
        { status: 500, headers }
      );
    }

    // Generate JWT (30-day CLI token, same claims the relay server validates)
    const token = mintServerToken(
      session.user.id,
      session.user.email,
      30 * 24 * 60 * 60,
      session.user.username ?? undefined,
    );

    // Update auth request
    await prisma.cliAuthRequest.update({
      where: { id: authReq.id },
      data: { status: "approved", token, userId: session.user.id },
    });

    return NextResponse.json({ token }, { headers });
  } catch (error) {
    console.error("cli-auth token error:", error);
    return NextResponse.json(
      { error: "Failed to generate token" },
      { status: 500, headers }
    );
  }
}
