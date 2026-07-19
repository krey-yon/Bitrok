import { betterAuth } from "better-auth";
import { prismaAdapter } from "better-auth/adapters/prisma";
import { prisma } from "@/lib/prisma";
import { getAuthBaseURL, getTrustedOrigins } from "@/lib/app-url";

function requireEnv(name: string): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(`Missing required environment variable: ${name}`);
  }
  return value;
}

// Never throw on module load during `next build` — Vercel collects page data
// with NODE_ENV=production even when BETTER_AUTH_URL is still the local default.
// getAuthBaseURL() rewrites localhost → https://bitrok.tech on Vercel/prod.
const baseURL = getAuthBaseURL();

if (
  process.env.VERCEL_ENV === "production" &&
  /localhost|127\.0\.0\.1/.test(process.env.BETTER_AUTH_URL || "")
) {
  // Soft warning only — build must succeed; runtime uses rewritten baseURL.
  console.warn(
    `[bitrok] BETTER_AUTH_URL is localhost on Vercel production. Using ${baseURL}. ` +
      `Set BETTER_AUTH_URL=https://bitrok.tech and NEXT_PUBLIC_APP_URL=https://bitrok.tech in project env.`,
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
