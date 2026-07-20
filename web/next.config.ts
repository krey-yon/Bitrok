import type { NextConfig } from "next";

const isProduction = process.env.NODE_ENV === "production";
const umamiWebsiteId = process.env.NEXT_PUBLIC_UMAMI_WEBSITE_ID;
const umamiDomain = process.env.NEXT_PUBLIC_UMAMI_DOMAIN || "https://cloud.umami.is";
let umamiOrigin = "";
try {
  umamiOrigin = new URL(umamiDomain).origin;
} catch {
  // Invalid optional analytics configuration must never weaken the CSP.
}

const scriptSources = ["'self'", "'unsafe-inline'"];
const connectSources = ["'self'"];
if (umamiWebsiteId && umamiOrigin) {
  scriptSources.push(umamiOrigin);
  connectSources.push(umamiOrigin);
}

const nextConfig: NextConfig = {
  poweredByHeader: false,
  async headers() {
    return [
      {
        source: "/(.*)",
        headers: [
          {
            key: "X-DNS-Prefetch-Control",
            value: "on",
          },
          {
            key: "X-Frame-Options",
            value: "DENY",
          },
          {
            key: "X-Content-Type-Options",
            value: "nosniff",
          },
          {
            key: "Referrer-Policy",
            value: "strict-origin-when-cross-origin",
          },
          {
            key: "Permissions-Policy",
            value: "camera=(), microphone=(), geolocation=()",
          },
          // Strict-Transport-Security (HSTS) - only in production
          ...(isProduction
            ? [
                {
                  key: "Strict-Transport-Security",
                  value: "max-age=31536000; includeSubDomains; preload",
                },
              ]
            : []),
          // Content-Security-Policy
          {
            key: "Content-Security-Policy",
            value: [
              "default-src 'self'",
              `script-src ${scriptSources.join(" ")}`,
              "style-src 'self' 'unsafe-inline'",
              "img-src 'self' https: data:",
              "font-src 'self'",
              `connect-src ${connectSources.join(" ")}`,
              "object-src 'none'",
              "frame-ancestors 'none'",
              "base-uri 'self'",
              "form-action 'self'",
              ...(isProduction ? ["upgrade-insecure-requests"] : []),
            ].join("; "),
          },
        ],
      },
    ];
  },
};

export default nextConfig;
