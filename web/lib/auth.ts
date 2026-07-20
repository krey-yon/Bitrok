import { betterAuth } from "better-auth";
import { prismaAdapter } from "better-auth/adapters/prisma";
import { prisma } from "@/lib/prisma";
import { getAuthBaseURL, getTrustedOrigins } from "@/lib/app-url";
import { selectVerifiedGithubEmail } from "@/lib/github-identity";

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
    // `username` is the immutable URL namespace claimed from dashboard settings.
    additionalFields: {
      username: { type: "string", required: false, input: false, returned: true },
    },
  },
  socialProviders: {
    github: {
      clientId: requireEnv("GITHUB_CLIENT_ID"),
      clientSecret: requireEnv("GITHUB_CLIENT_SECRET"),
      scope: ["read:user", "user:email"],
      getUserInfo: async (token) => {
        if (!token.accessToken) return null;
        const headers = {
          Accept: "application/vnd.github+json",
          Authorization: `Bearer ${token.accessToken}`,
          "X-GitHub-Api-Version": "2022-11-28",
        };
        const [profileResponse, emailsResponse] = await Promise.all([
          fetch("https://api.github.com/user", { headers }),
          fetch("https://api.github.com/user/emails", { headers }),
        ]);
        if (!profileResponse.ok || !emailsResponse.ok) return null;

        const profile = (await profileResponse.json()) as {
          id?: number | string;
          login?: string;
          name?: string | null;
          avatar_url?: string;
        };
        const emails = (await emailsResponse.json()) as Array<{
          email?: string;
          primary?: boolean;
          verified?: boolean;
        }>;
        const email = selectVerifiedGithubEmail(emails);
        if (!profile.id || !profile.login || !email) return null;

        return {
          user: {
            id: String(profile.id),
            email,
            emailVerified: true,
            name: profile.name || profile.login,
            image: profile.avatar_url,
          },
          data: profile,
        };
      },
    },
  },
  emailAndPassword: {
    // Password auth needs verified-email delivery and a recovery flow. Launch
    // with one verified identity provider until those controls are available.
    enabled: false,
  },
  advanced: {
    cookiePrefix: "bitrok",
  },
});
