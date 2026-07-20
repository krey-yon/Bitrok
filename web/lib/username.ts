import { Prisma } from "@prisma/client";
import { prisma } from "@/lib/prisma";
import { markUsernameTaken, usernameMayBeTaken } from "@/lib/username-bloom";

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

export type UsernameAvailability =
  | { available: true; username: string }
  | { available: false; username: string; error: string };

function validateUsername(raw: string): UsernameResult {
  const username = slugify(raw);
  if (!username) return { ok: false, error: "Use letters, numbers, or hyphens." };
  if (username.length < 2) return { ok: false, error: "Use at least 2 characters." };
  if (username.length > 32) return { ok: false, error: "Use 32 characters or fewer." };
  if (RESERVED.has(username)) return { ok: false, error: "That username is reserved." };
  if (/^\d+$/.test(username)) return { ok: false, error: "Add at least one letter." };
  return { ok: true, username };
}

export async function checkUsernameAvailability(
  userId: string,
  raw: string,
): Promise<UsernameAvailability> {
  const validated = validateUsername(raw);
  if (!validated.ok) return { available: false, username: slugify(raw), error: validated.error };

  const maybeTaken = await usernameMayBeTaken(validated.username);
  if (maybeTaken === false) return { available: true, username: validated.username };

  const taken = await prisma.user.findFirst({
    where: { username: validated.username, NOT: { id: userId } },
    select: { id: true },
  });
  return taken
    ? { available: false, username: validated.username, error: "That username is already taken." }
    : { available: true, username: validated.username };
}

/**
 * Explicit create/update of a username from the dashboard.
 * Validates format + uniqueness, then persists.
 */
export async function setUsernameForUser(
  userId: string,
  raw: string,
): Promise<UsernameResult> {
  const validated = validateUsername(raw);
  if (!validated.ok) return validated;
  const username = validated.username;

  const taken = await prisma.user.findFirst({
    where: { username, NOT: { id: userId } },
    select: { id: true },
  });
  if (taken) {
    return { ok: false, error: "That username is already taken." };
  }

  try {
    await prisma.user.update({ where: { id: userId }, data: { username } });
  } catch (error) {
    if (error instanceof Prisma.PrismaClientKnownRequestError && error.code === "P2002") {
      return { ok: false, error: "That username is already taken." };
    }
    throw error;
  }

  await markUsernameTaken(username);

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
