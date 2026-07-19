import { betterAuth } from "better-auth";
import { prismaAdapter } from "better-auth/adapters/prisma";
import { prisma } from "@/lib/prisma";

function requireEnv(name: string): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(`Missing required environment variable: ${name}`);
  }
  return value;
}

export const auth = betterAuth({
  database: prismaAdapter(prisma, {
    provider: "postgresql",
  }),
  secret: requireEnv("BETTER_AUTH_SECRET"),
  baseURL: requireEnv("BETTER_AUTH_URL"),
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
        // Fallback: if GitHub returns no email (private/unverified),
        // construct a deterministic one from the GitHub ID so auth can proceed.
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
