import { auth } from "@/lib/auth";
import { mintCliToken } from "@/lib/cli-token";
import { prisma } from "@/lib/prisma";
import { rateLimit, getRateLimitHeaders } from "@/lib/rate-limit";
import { getTrustedOrigins } from "@/lib/app-url";
import { getUsernameForUser } from "@/lib/username";
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

    const sessionUser = session.user as {
      id: string;
      email?: string | null;
      name?: string | null;
      username?: string | null;
    };
    const username = await getUsernameForUser(sessionUser.id);
    if (!username) {
      return NextResponse.json(
        { error: "Choose a username before authorizing the CLI.", requiresUsername: true },
        { status: 409, headers },
      );
    }

    const token = await mintCliToken(
      {
        userId: sessionUser.id,
        email: sessionUser.email ?? undefined,
        username,
      },
      30 * 24 * 60 * 60,
    );

    // Update auth request
    await prisma.cliAuthRequest.update({
      where: { id: authReq.id },
      data: { status: "approved", userId: session.user.id },
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
