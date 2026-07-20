import { auth } from "@/lib/auth";
import { rateLimit, getRateLimitHeaders } from "@/lib/rate-limit";
import {
  getServerTunnels,
  createServerTunnel,
  ServerApiError,
} from "@/lib/server-api";
import { NextRequest, NextResponse } from "next/server";
import { getClientIp, hasTrustedOrigin } from "@/lib/request-security";
import { getUsernameForUser } from "@/lib/username";
import { z } from "zod";

// This route is a thin proxy to the Go relay server's /api/tunnels.
// The relay server (SQLite) is the single source of truth for tunnel data;
// the dashboard no longer keeps its own copy in Postgres.

const createTunnelSchema = z.object({
  name: z
    .string()
    .trim()
    .toLowerCase()
    .min(1)
    .max(63)
    .regex(/^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/),
  host: z.string().trim().toLowerCase().min(1).max(253).regex(/^[a-z0-9.-]+$/),
  port: z.coerce.number().int().min(1).max(65535),
});

export async function GET(req: NextRequest) {
  const clientIp = getClientIp(req);
  const rateLimitResult = await rateLimit(`tunnels:get:${clientIp}`, {
    windowMs: 60 * 1000,
    maxRequests: 30,
  });

  const headers = getRateLimitHeaders(rateLimitResult);

  if (!rateLimitResult.success) {
    return NextResponse.json(
      { error: "Rate limit exceeded. Please try again later." },
      { status: 429, headers }
    );
  }

  const session = await auth.api.getSession({ headers: req.headers });
  if (!session) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401, headers });
  }

  try {
    const tunnels = await getServerTunnels(session.user.id);
    return NextResponse.json(tunnels, { headers });
  } catch (err) {
    console.error("Failed to list tunnels from relay:", err);
    return NextResponse.json(
      { error: "Failed to load tunnels" },
      { status: 502, headers }
    );
  }
}

export async function POST(req: NextRequest) {
  const clientIp = getClientIp(req);
  const rateLimitResult = await rateLimit(`tunnels:post:${clientIp}`, {
    windowMs: 60 * 1000,
    maxRequests: 10,
  });

  const headers = getRateLimitHeaders(rateLimitResult);

  if (!rateLimitResult.success) {
    return NextResponse.json(
      { error: "Rate limit exceeded. Please try again later." },
      { status: 429, headers }
    );
  }

  if (!hasTrustedOrigin(req)) {
    return NextResponse.json({ error: "Invalid origin" }, { status: 403, headers });
  }

  const session = await auth.api.getSession({ headers: req.headers });
  if (!session) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401, headers });
  }

  let body: unknown;
  try {
    body = await req.json();
  } catch {
    return NextResponse.json(
      { error: "Invalid JSON body" },
      { status: 400, headers }
    );
  }

  const parsed = createTunnelSchema.safeParse(body);
  if (!parsed.success) {
    return NextResponse.json(
      { error: "Invalid input", issues: parsed.error.issues },
      { status: 400, headers }
    );
  }

  const { name, host, port } = parsed.data;

  try {
    const username = await getUsernameForUser(session.user.id);
    if (!username) {
      return NextResponse.json(
        { error: "Create a username before reserving a tunnel" },
        { status: 409, headers },
      );
    }

    const tunnel = await createServerTunnel(
      session.user.id,
      { name, host, port },
      username,
    );
    return NextResponse.json(tunnel, { status: 201, headers });
  } catch (err) {
    if (err instanceof ServerApiError) {
      // 409 Conflict from the relay = host already taken
      if (err.status === 409) {
        return NextResponse.json(
          { error: "A tunnel with this host already exists" },
          { status: 409, headers }
        );
      }
      // 400 = validation failure on the server side
      if (err.status === 400) {
        return NextResponse.json(
          { error: err.message || "Invalid input" },
          { status: 400, headers }
        );
      }
      console.error("Relay rejected create tunnel:", err.status, err.message);
      return NextResponse.json(
        { error: "Failed to create tunnel" },
        { status: 502, headers }
      );
    }
    console.error("Failed to create tunnel:", err);
    return NextResponse.json(
      { error: "Failed to create tunnel" },
      { status: 500, headers }
    );
  }
}
