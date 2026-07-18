import { auth } from "@/lib/auth";
import { rateLimit, getRateLimitHeaders } from "@/lib/rate-limit";
import {
  getServerTunnels,
  createServerTunnel,
  ServerApiError,
} from "@/lib/server-api";
import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";

// This route is a thin proxy to the Go relay server's /api/tunnels.
// The relay server (SQLite) is the single source of truth for tunnel data;
// the dashboard no longer keeps its own copy in Postgres.

const createTunnelSchema = z.object({
  name: z.string().min(1).max(100).regex(/^[a-zA-Z0-9-_]+$/),
  host: z.string().min(1).max(255).regex(/^[a-zA-Z0-9-_.]+$/),
  port: z.coerce.number().int().min(1).max(65535),
});

function getClientIp(req: NextRequest): string {
  const forwarded = req.headers.get("x-forwarded-for");
  return forwarded ? forwarded.split(",")[0].trim() : "unknown";
}

export async function GET(req: NextRequest) {
  const clientIp = getClientIp(req);
  const rateLimitResult = rateLimit(`tunnels:get:${clientIp}`, {
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
  const rateLimitResult = rateLimit(`tunnels:post:${clientIp}`, {
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
    const tunnel = await createServerTunnel(session.user.id, { name, host, port });
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
