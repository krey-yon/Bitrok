import { mintServerToken } from "@/lib/jwt";

// Server-only client for the Go relay server's REST API.
//
// The dashboard mints a short-lived JWT (60s) per call, signed with the same
// BITROK_JWT_SECRET the relay validates, so the server scopes every response
// by the user's id (the `sub` claim). No long-lived service token, no shared
// static secret — the trust path is identical to the CLI's.
//
// All functions are server-only: they read process.env and must never be
// imported from a client component.

const DEFAULT_TIMEOUT_MS = 10_000;

/** A tunnel row from the relay server (GET/POST /api/tunnels). */
export type TunnelDTO = {
  id: string;
  user_id: string;
  name: string;
  host: string;
  port: number;
  active: boolean;
  created_at: string;
  updated_at: string;
};

/** A request-log row from the relay server (GET /api/logs). */
export type LogDTO = {
  id: number;
  tunnel_id: string;
  tunnel_name?: string;
  method: string;
  path: string;
  status: number;
  latency_ms: number;
  bytes_in: number;
  bytes_out: number;
  ts: string;
};

export type LogsResponse = {
  total: number;
  logs: LogDTO[];
};

/**
 * Resolve the relay server base URL.
 * 1. BITROK_SERVER_URL if set.
 * 2. http://localhost:8080 in dev (NEXT_PUBLIC_USE_LOCALHOST truthy).
 * 3. http://localhost:8080 as a last-resort fallback.
 */
function serverBaseUrl(): string {
  const explicit = process.env.BITROK_SERVER_URL;
  if (explicit) return explicit.replace(/\/+$/, "");

  const useLocal =
    process.env.NEXT_PUBLIC_USE_LOCALHOST &&
    process.env.NEXT_PUBLIC_USE_LOCALHOST !== "false" &&
    process.env.NEXT_PUBLIC_USE_LOCALHOST !== "0";
  if (useLocal) return "http://localhost:8080";

  return "http://localhost:8080";
}

class ServerApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
    this.name = "ServerApiError";
  }
}

async function serverFetch(
  userId: string,
  path: string,
  init: RequestInit = {},
): Promise<Response> {
  const token = mintServerToken(userId);
  const base = serverBaseUrl();

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), DEFAULT_TIMEOUT_MS);

  try {
    const res = await fetch(`${base}${path}`, {
      ...init,
      headers: {
        Authorization: `Bearer ${token}`,
        ...(init.body ? { "Content-Type": "application/json" } : {}),
        ...(init.headers || {}),
      },
      signal: controller.signal,
      // Always hit the server fresh — this is server-to-server, no caching.
      cache: "no-store",
    });
    return res;
  } finally {
    clearTimeout(timer);
  }
}

/** Extract a human-readable error message from a non-2xx response. */
async function errorMessage(res: Response): Promise<string> {
  try {
    const data = await res.json();
    if (data && typeof data.error === "string") return data.error;
    if (data && typeof data.message === "string") return data.message;
  } catch {
    /* not JSON */
  }
  return `relay server returned HTTP ${res.status}`;
}

/** GET /api/tunnels — list the user's tunnels (active flag computed by server). */
export async function getServerTunnels(userId: string): Promise<TunnelDTO[]> {
  const res = await serverFetch(userId, "/api/tunnels");
  if (!res.ok) throw new ServerApiError(res.status, await errorMessage(res));

  const data = await res.json();
  // Server returns { tunnels: [...] }; be defensive about a bare array too.
  if (Array.isArray(data)) return data as TunnelDTO[];
  if (data && Array.isArray(data.tunnels)) return data.tunnels as TunnelDTO[];
  return [];
}

/** POST /api/tunnels — create a tunnel. */
export async function createServerTunnel(
  userId: string,
  body: { name: string; host: string; port: number },
): Promise<TunnelDTO> {
  const res = await serverFetch(userId, "/api/tunnels", {
    method: "POST",
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new ServerApiError(res.status, await errorMessage(res));
  return (await res.json()) as TunnelDTO;
}

/** GET /api/logs?limit=N — recent request logs + total count. */
export async function getServerLogs(
  userId: string,
  limit = 10,
): Promise<LogsResponse> {
  const res = await serverFetch(
    userId,
    `/api/logs?limit=${encodeURIComponent(limit)}`,
  );
  if (!res.ok) throw new ServerApiError(res.status, await errorMessage(res));

  const data = (await res.json()) as LogsResponse;
  // Defend against missing fields so the dashboard degrades gracefully.
  return {
    total: typeof data.total === "number" ? data.total : 0,
    logs: Array.isArray(data.logs) ? data.logs : [],
  };
}

export { ServerApiError };
