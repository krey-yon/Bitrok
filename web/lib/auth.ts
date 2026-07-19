import { betterAuth } from "better-auth";
import { prismaAdapter } from "better-auth/adapters/prisma";
import { prisma } from "@/lib/prisma";
import { getAppURL, getTrustedOrigins } from "@/lib/app-url";

function requireEnv(name: string): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(`Missing required environment variable: ${name}`);
  }
  return value;
}

// Prefer explicit BETTER_AUTH_URL; fall back to public app URL so a missing
// env doesn't silently use localhost on Vercel.
const baseURL = (process.env.BETTER_AUTH_URL || getAppURL()).replace(/\/+$/, "");

if (process.env.NODE_ENV === "production" && /localhost|127\.0\.0\.1/.test(baseURL)) {
  throw new Error(
    `BETTER_AUTH_URL/baseURL must not be localhost in production (got ${baseURL}). ` +
      `Set BETTER_AUTH_URL=https://bitrok.tech and NEXT_PUBLIC_APP_URL=https://bitrok.tech on Vercel.`,
  );
}

export const auth = betterAuth({
  database: prismaAdapter(prisma, {
    provider: "postgresql",
  }),
  secret: requireEnv("BETTER_AUTH_SECRET"),
  baseURL,
  trustedOrigins: getTrustedOrigins(),
  user: {
    // `username` is the URL slug used in deterministic tunnel hosts
    // (app-username.bitrok.tech). Populated from the GitHub login at signup.
    additionalFields: {
      username: { type: "string", required: false, input: true, returned: true },
    },
  },
  socialProviders: {
    github: {
      clientId: requireEnv("GITHUB_CLIENT_ID"),
      clientSecret: requireEnv("GITHUB_CLIENT_SECRET"),
      scope: ["read:user", "user:email"],
      mapProfileToUser: (profile) => {
        return {
          email: profile.email || `${profile.id}@users.noreply.github.com`,
          name: profile.name || profile.login,
          image: profile.avatar_url,
          username: profile.login,
        };
      },
    },
  },
  emailAndPassword: {
    enabled: true,
    autoSignIn: true,
  },
  advanced: {
    cookiePrefix: "bitrok",
  },
});
