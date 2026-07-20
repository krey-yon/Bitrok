import "server-only";

import Redis from "ioredis";

const globalForRedis = globalThis as unknown as {
  bitrokRedisV1: Redis | null | undefined;
};

function createRedis(): Redis | null {
  const connectionString = (
    process.env.BITROK_REDIS_URL || process.env.UPSTASH_REDIS_KEY
  )?.trim();
  if (!connectionString) return null;

  return new Redis(connectionString, {
    lazyConnect: true,
    maxRetriesPerRequest: 2,
    connectTimeout: 5_000,
  });
}

export const redis = globalForRedis.bitrokRedisV1 ?? createRedis();

if (process.env.NODE_ENV !== "production") {
  globalForRedis.bitrokRedisV1 = redis;
}
