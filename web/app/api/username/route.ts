import { auth } from "@/lib/auth";
import { rateLimit, getRateLimitHeaders } from "@/lib/rate-limit";
import {
  checkUsernameAvailability,
  getUsernameForUser,
  setUsernameForUser,
  slugify,
} from "@/lib/username";
import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";

function getClientIp(req: NextRequest): string {
  const forwarded = req.headers.get("x-forwarded-for");
  return forwarded ? forwarded.split(",")[0].trim() : "unknown";
}

const bodySchema = z.object({
  username: z.string().min(1).max(64),
});

/** GET /api/username — current username for the signed-in user. */
export async function GET(req: NextRequest) {
  const clientIp = getClientIp(req);
  const rateLimitResult = rateLimit(`username:get:${clientIp}`, {
    windowMs: 60 * 1000,
    maxRequests: 60,
  });
  const headers = getRateLimitHeaders(rateLimitResult);
  if (!rateLimitResult.success) {
    return NextResponse.json(
      { error: "Rate limit exceeded. Please try again later." },
      { status: 429, headers },
    );
  }

  const session = await auth.api.getSession({ headers: req.headers });
  if (!session) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401, headers });
  }

  const candidate = req.nextUrl.searchParams.get("candidate");
  if (candidate !== null) {
    const result = await checkUsernameAvailability(session.user.id, candidate);
    return NextResponse.json(result, { headers });
  }

  const username = await getUsernameForUser(session.user.id);
  return NextResponse.json(
    {
      username,
      preview: username ? `myapp-${username}.bitrok.tech` : null,
    },
    { headers },
  );
}

/** PUT /api/username — create or update username. */
export async function PUT(req: NextRequest) {
  const clientIp = getClientIp(req);
  const rateLimitResult = rateLimit(`username:put:${clientIp}`, {
    windowMs: 60 * 1000,
    maxRequests: 10,
  });
  const headers = getRateLimitHeaders(rateLimitResult);
  if (!rateLimitResult.success) {
    return NextResponse.json(
      { error: "Rate limit exceeded. Please try again later." },
      { status: 429, headers },
    );
  }

  const session = await auth.api.getSession({ headers: req.headers });
  if (!session) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401, headers });
  }

  let body: unknown;
  try {
    body = await req.json();
  } catch {
    return NextResponse.json({ error: "Invalid JSON body" }, { status: 400, headers });
  }

  const parsed = bodySchema.safeParse(body);
  if (!parsed.success) {
    return NextResponse.json(
      { error: "Enter a username between 2 and 32 characters." },
      { status: 400, headers },
    );
  }

  const result = await setUsernameForUser(session.user.id, parsed.data.username);
  if (!result.ok) {
    return NextResponse.json({ error: result.error }, { status: 400, headers });
  }

  return NextResponse.json(
    {
      username: result.username,
      preview: `myapp-${result.username}.bitrok.tech`,
      normalized: slugify(parsed.data.username) !== parsed.data.username.trim().toLowerCase(),
    },
    { headers },
  );
}
