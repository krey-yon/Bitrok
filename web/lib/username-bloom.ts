import "server-only";

import { prisma } from "@/lib/prisma";
import { redis } from "@/lib/redis";

const BLOOM_KEY = "bitrok:usernames:v1";
const SEEDED_KEY = `${BLOOM_KEY}:seeded`;
const SEED_LOCK_KEY = `${BLOOM_KEY}:seed-lock`;

async function ensureSeeded(): Promise<boolean> {
  if (!redis) return false;

  try {
    if (await redis.get(SEEDED_KEY)) return true;
    const ownsLock = await redis.set(SEED_LOCK_KEY, "1", "EX", 30, "NX");
    if (!ownsLock) return false;

    try {
      const users = await prisma.user.findMany({
        where: { username: { not: null } },
        select: { username: true },
      });
      const pipeline = redis.pipeline();
      for (const user of users) {
        if (user.username) pipeline.call("BF.ADD", BLOOM_KEY, user.username);
      }
      pipeline.set(SEEDED_KEY, "1");
      await pipeline.exec();
      return true;
    } finally {
      await redis.del(SEED_LOCK_KEY).catch(() => undefined);
    }
  } catch {
    return false;
  }
}

// false is definitive only after the filter has been seeded. true remains a
// probabilistic result and must be confirmed against PostgreSQL.
export async function usernameMayBeTaken(username: string): Promise<boolean | null> {
  if (!(await ensureSeeded()) || !redis) return null;
  try {
    return Number(await redis.call("BF.EXISTS", BLOOM_KEY, username)) === 1;
  } catch {
    return null;
  }
}

export async function markUsernameTaken(username: string): Promise<void> {
  if (!redis) return;
  try {
    await redis.call("BF.ADD", BLOOM_KEY, username);
  } catch {
    // Redis is an optimization; PostgreSQL's unique constraint is authoritative.
  }
}
