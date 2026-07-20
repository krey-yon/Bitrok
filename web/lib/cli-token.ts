import "server-only";

import { createHash, randomBytes } from "node:crypto";
import { mintServerToken } from "@/lib/jwt";
import { redis } from "@/lib/redis";

const TOKEN_PREFIX = "br_sk_";

type CliTokenRecord = {
  userId: string;
  email?: string;
  username: string;
};

export async function mintCliToken(
  record: CliTokenRecord,
  ttlSeconds: number,
): Promise<string> {
  if (!redis) {
    return mintServerToken(record.userId, record.email, ttlSeconds, record.username);
  }

  const token = TOKEN_PREFIX + randomBytes(16).toString("hex");
  const digest = createHash("sha256").update(token).digest("hex");
  try {
    await redis.set(`bitrok:cli-token:${digest}`, JSON.stringify(record), "EX", ttlSeconds);
    return token;
  } catch (error) {
    // When Redis is configured, silently issuing a non-revocable JWT changes
    // the credential's security contract and can hide a split web/relay setup.
    throw new Error("Could not store the revocable CLI token", { cause: error });
  }
}
