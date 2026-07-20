import { auth } from "@/lib/auth";
import { mintCliToken } from "@/lib/cli-token";
import { rateLimit, getRateLimitHeaders } from "@/lib/rate-limit";
import { getClientIp, hasTrustedOrigin } from "@/lib/request-security";
import { getUsernameForUser } from "@/lib/username";
import { NextRequest, NextResponse } from "next/server";

const TOKEN_TTL_DAYS = 30;

export async function POST(req: NextRequest) {
  const clientIp = getClientIp(req);
  const rateLimitResult = await rateLimit(`cli-auth:generate:${clientIp}`, {
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

  if (!hasTrustedOrigin(req)) {
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

    // Always embed a username claim - CLI builds hosts as app-<username>.domain.
    const sessionUser = session.user as {
      id: string;
      email?: string | null;
      name?: string | null;
      username?: string | null;
    };
    const username = await getUsernameForUser(sessionUser.id);
    if (!username) {
      return NextResponse.json(
        { error: "Choose a username before generating a CLI token.", requiresUsername: true },
        { status: 409, headers },
      );
    }

    const expiresInSeconds = TOKEN_TTL_DAYS * 24 * 60 * 60;
    const token = await mintCliToken(
      {
        userId: sessionUser.id,
        email: sessionUser.email ?? undefined,
        username,
      },
      expiresInSeconds,
    );

    const expiresAt = new Date(Date.now() + expiresInSeconds * 1000).toISOString();
    return NextResponse.json({ token, expiresAt, username }, { headers });
  } catch (error) {
    console.error("cli-auth generate error:", error);
    return NextResponse.json(
      { error: "Failed to generate token" },
      { status: 500, headers }
    );
  }
}
