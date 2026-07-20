import "server-only";

import { redis } from "@/lib/redis";

interface RateLimitEntry {
  count: number;
  resetTime: number;
}

interface RateLimitOptions {
  windowMs?: number;
  maxRequests?: number;
}

export type RateLimitResult = {
  success: boolean;
  limit: number;
  remaining: number;
  resetTime: number;
};

const localStore = new Map<string, RateLimitEntry>();
let lastLocalCleanup = 0;

const FIXED_WINDOW_SCRIPT = `
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("PTTL", KEYS[1])
return {current, ttl}
`;

export async function rateLimit(
  identifier: string,
  options: RateLimitOptions = {},
): Promise<RateLimitResult> {
  const { windowMs = 60_000, maxRequests = 10 } = options;
  const now = Date.now();

  if (redis) {
    try {
      const result = (await redis.eval(
        FIXED_WINDOW_SCRIPT,
        1,
        `bitrok:rate-limit:${identifier}`,
        windowMs,
      )) as [number, number];
      const count = Number(result[0]);
      const ttl = Math.max(Number(result[1]), 0);
      return {
        success: count <= maxRequests,
        limit: maxRequests,
        remaining: Math.max(maxRequests - count, 0),
        resetTime: now + ttl,
      };
    } catch (error) {
      console.warn("Redis rate limiter unavailable; using local fallback", error);
    }
  }

  return localRateLimit(identifier, windowMs, maxRequests, now);
}

function localRateLimit(
  identifier: string,
  windowMs: number,
  maxRequests: number,
  now: number,
): RateLimitResult {
  if (now-lastLocalCleanup > 5 * 60_000) {
    for (const [key, entry] of localStore.entries()) {
      if (entry.resetTime <= now) localStore.delete(key);
    }
    lastLocalCleanup = now;
  }

  const entry = localStore.get(identifier);
  if (!entry || entry.resetTime <= now) {
    const created = { count: 1, resetTime: now + windowMs };
    localStore.set(identifier, created);
    return {
      success: true,
      limit: maxRequests,
      remaining: maxRequests - 1,
      resetTime: created.resetTime,
    };
  }

  entry.count += 1;
  return {
    success: entry.count <= maxRequests,
    limit: maxRequests,
    remaining: Math.max(maxRequests - entry.count, 0),
    resetTime: entry.resetTime,
  };
}

export function getRateLimitHeaders(result: RateLimitResult): Record<string, string> {
  return {
    "X-RateLimit-Limit": result.limit.toString(),
    "X-RateLimit-Remaining": result.remaining.toString(),
    "X-RateLimit-Reset": Math.ceil(result.resetTime / 1000).toString(),
  };
}
