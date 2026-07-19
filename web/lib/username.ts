import { prisma } from "@/lib/prisma";

/** Host labels that must never be claimed as a user slug. */
const RESERVED = new Set([
  "api",
  "www",
  "app",
  "admin",
  "dashboard",
  "static",
  "assets",
  "cdn",
  "mail",
  "ftp",
  "status",
  "support",
  "help",
  "docs",
  "blog",
  "auth",
  "login",
  "register",
  "bitrok",
  "tunnel",
  "tunnels",
  "cli",
  "null",
  "undefined",
]);

export type UsernameResult =
  | { ok: true; username: string }
  | { ok: false; error: string };

/**
 * Explicit create/update of a username from the dashboard.
 * Validates format + uniqueness, then persists.
 */
export async function setUsernameForUser(
  userId: string,
  raw: string,
): Promise<UsernameResult> {
  const username = slugify(raw);
  if (!username) {
    return {
      ok: false,
      error: "Username must use letters, numbers, or hyphens (e.g. kreyon).",
    };
  }
  if (username.length < 2) {
    return { ok: false, error: "Username must be at least 2 characters." };
  }
  if (username.length > 32) {
    return { ok: false, error: "Username must be 32 characters or fewer." };
  }
  if (RESERVED.has(username)) {
    return { ok: false, error: "That username is reserved. Pick another." };
  }
  if (/^\d+$/.test(username)) {
    return {
      ok: false,
      error: "Username cannot be only numbers. Add a letter.",
    };
  }

  const taken = await prisma.user.findFirst({
    where: { username, NOT: { id: userId } },
    select: { id: true },
  });
  if (taken) {
    return { ok: false, error: "That username is already taken." };
  }

  await prisma.user.update({
    where: { id: userId },
    data: { username },
  });

  return { ok: true, username };
}

/** Load the current username from the DB (session may not include additional fields). */
export async function getUsernameForUser(userId: string): Promise<string | null> {
  const user = await prisma.user.findUnique({
    where: { id: userId },
    select: { username: true },
  });
  if (!user?.username) return null;
  const s = slugify(user.username);
  return s || null;
}

/**
 * Resolve a stable URL slug for tunnel hosts: <app>-<username>.bitrok.tech.
 *
 * Prefer the stored `user.username`. If missing (legacy accounts, incomplete
 * GitHub map), derive one, persist it, and return it so CLI JWTs always carry
 * the claim.
 */
export async function resolveUsernameForUser(
  userId: string,
  hints?: { email?: string | null; name?: string | null; sessionUsername?: string | null },
): Promise<string> {
  const user = await prisma.user.findUnique({
    where: { id: userId },
    select: { id: true, email: true, name: true, username: true },
  });
  if (!user) {
    throw new Error("user not found");
  }

  const fromSession = slugify(hints?.sessionUsername ?? "");
  if (user.username) {
    const stored = slugify(user.username);
    if (stored) return stored;
  }
  if (fromSession) {
    return await ensureUniqueUsername(userId, fromSession);
  }

  // GitHub noreply: 12345@users.noreply.github.com — prefer name or short id.
  const candidates = [
    slugify(user.name ?? ""),
    slugify(hints?.name ?? ""),
    slugify(emailLocalPart(user.email)),
    slugify(emailLocalPart(hints?.email ?? "")),
    slugify(user.id.slice(0, 10)),
  ].filter(Boolean);

  for (const candidate of candidates) {
    const taken = await prisma.user.findFirst({
      where: { username: candidate, NOT: { id: userId } },
      select: { id: true },
    });
    if (!taken) {
      return await ensureUniqueUsername(userId, candidate);
    }
  }

  // Last resort: user id tail + random nibble
  const fallback = slugify(`${user.id.slice(-8)}-${Math.random().toString(36).slice(2, 5)}`);
  return await ensureUniqueUsername(userId, fallback || "user");
}

async function ensureUniqueUsername(userId: string, base: string): Promise<string> {
  let candidate = base.slice(0, 32);
  for (let i = 0; i < 20; i++) {
    const taken = await prisma.user.findFirst({
      where: { username: candidate, NOT: { id: userId } },
      select: { id: true },
    });
    if (!taken) {
      await prisma.user.update({
        where: { id: userId },
        data: { username: candidate },
      });
      return candidate;
    }
    candidate = `${base.slice(0, 24)}-${i + 2}`;
  }
  const last = `${base.slice(0, 20)}-${Date.now().toString(36)}`.slice(0, 32);
  await prisma.user.update({
    where: { id: userId },
    data: { username: last },
  });
  return last;
}

function emailLocalPart(email: string): string {
  if (!email) return "";
  const local = email.split("@")[0] || "";
  // Skip pure-numeric github noreply ids as primary brand slug when possible
  // (still usable as fallback later in the candidate list).
  return local;
}

/** DNS-label-safe slug: lowercase, a-z0-9-, max 32. */
export function slugify(input: string): string {
  return input
    .toLowerCase()
    .trim()
    .replace(/[_\s.]+/g, "-")
    .replace(/[^a-z0-9-]/g, "")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "")
    .slice(0, 32)
    .replace(/-$/g, "");
}
