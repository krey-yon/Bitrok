import { auth } from "@/lib/auth";
import { mintServerToken } from "@/lib/jwt";
import { rateLimit, getRateLimitHeaders } from "@/lib/rate-limit";
import { NextRequest, NextResponse } from "next/server";

function getClientIp(req: NextRequest): string {
  const forwarded = req.headers.get("x-forwarded-for");
  return forwarded ? forwarded.split(",")[0].trim() : "unknown";
}

// Validate Origin header for CSRF protection.
function validateOrigin(req: NextRequest): boolean {
  const origin = req.headers.get("origin");
  const allowedOrigins = [
    process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3000",
  ];
  return !origin || allowedOrigins.includes(origin);
}

const TOKEN_TTL_DAYS = 30;

export async function POST(req: NextRequest) {
  const clientIp = getClientIp(req);
  const rateLimitResult = rateLimit(`cli-auth:generate:${clientIp}`, {
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

    const jwtSecret = process.env.BITROK_JWT_SECRET;
    if (!jwtSecret) {
      console.error("BITROK_JWT_SECRET not configured");
      return NextResponse.json(
        { error: "Server configuration error" },
        { status: 500, headers }
      );
    }

    const expiresInSeconds = TOKEN_TTL_DAYS * 24 * 60 * 60;
    const token = mintServerToken(
      session.user.id,
      session.user.email,
      expiresInSeconds,
    );

    const expiresAt = new Date(Date.now() + expiresInSeconds * 1000).toISOString();
    return NextResponse.json({ token, expiresAt }, { headers });
  } catch (error) {
    console.error("cli-auth generate error:", error);
    return NextResponse.json(
      { error: "Failed to generate token" },
      { status: 500, headers }
    );
  }
}
