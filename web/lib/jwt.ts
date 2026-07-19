import jwt from "jsonwebtoken";

// Server-only JWT minter. Signs tokens with BITROK_JWT_SECRET using the exact
// claims the Go relay server's AuthMiddleware validates:
//   - aud: "bitrok-cli"
//   - iss: "bitrok"
//   - sub: the user id (the server scopes tunnels/logs by this)
//   - exp: now + ttlSeconds
//
// The dashboard mints short-lived tokens (60s) for its own server-to-server
// calls to the relay, and long-lived tokens (30d) for the CLI. Same claims,
// same secret — the server can't tell them apart, by design.

const JWT_AUDIENCE = "bitrok-cli";
const JWT_ISSUER = "bitrok";

/**
 * Mint a JWT the Go relay server will accept.
 *
 * @param userId    The user id, placed in the `sub` claim.
 * @param email     Optional email, placed in the payload for traceability.
 * @param ttlSeconds Lifetime in seconds. Default 60 (server-to-server).
 *                   CLI tokens pass 30 * 24 * 60 * 60.
 * @param username  Optional URL slug, placed in the payload so the CLI can
 *                  build deterministic hosts (app-username.bitrok.tech) without
 *                  an extra API round-trip. Only set on long-lived CLI tokens.
 */
export function mintServerToken(
  userId: string,
  email?: string,
  ttlSeconds: number = 60,
  username?: string,
): string {
  const secret = process.env.BITROK_JWT_SECRET;
  if (!secret) {
    throw new Error(
      "BITROK_JWT_SECRET is not configured. The web dashboard needs it to talk to the relay server.",
    );
  }

  const payload: { sub: string; email?: string; username?: string; type: string } = {
    sub: userId,
    type: "cli",
  };
  if (email) {
    payload.email = email;
  }
  if (username) {
    payload.username = username;
  }

  return jwt.sign(payload, secret, {
    expiresIn: ttlSeconds,
    audience: JWT_AUDIENCE,
    issuer: JWT_ISSUER,
  });
}
