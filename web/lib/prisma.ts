import "server-only";

import { PrismaClient } from "@prisma/client";

const globalForPrisma = globalThis as unknown as {
  prisma: PrismaClient | undefined;
};

function positiveIntegerEnv(name: string, fallback: number): string {
  const value = Number.parseInt(process.env[name] ?? "", 10);
  return String(Number.isFinite(value) && value > 0 ? value : fallback);
}

function getDatasourceUrl(): string | undefined {
  const value = process.env.DATABASE_URL;
  if (!value) return undefined;

  const url = new URL(value);
  if (url.protocol !== "postgres:" && url.protocol !== "postgresql:") {
    return value;
  }

  // Prisma otherwise sizes every runtime instance from the host CPU count.
  // In Next dev/HMR and serverless deployments that can overwhelm a shared
  // Postgres/Neon pool before any one instance is doing meaningful work.
  if (!url.searchParams.has("connection_limit")) {
    url.searchParams.set(
      "connection_limit",
      positiveIntegerEnv("PRISMA_CONNECTION_LIMIT", 5),
    );
  }
  if (!url.searchParams.has("pool_timeout")) {
    url.searchParams.set("pool_timeout", positiveIntegerEnv("PRISMA_POOL_TIMEOUT", 15));
  }
  if (!url.searchParams.has("connect_timeout")) {
    url.searchParams.set(
      "connect_timeout",
      positiveIntegerEnv("PRISMA_CONNECT_TIMEOUT", 10),
    );
  }

  return url.toString();
}

export const prisma =
  globalForPrisma.prisma ??
  new PrismaClient({
    datasourceUrl: getDatasourceUrl(),
  });

if (process.env.NODE_ENV !== "production") globalForPrisma.prisma = prisma;
