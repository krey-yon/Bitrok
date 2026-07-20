import type { Metadata } from "next";
import Script from "next/script";
import { ThemeProvider } from "@/components/theme-provider";
import "./globals.css";

const baseUrl = process.env.NEXT_PUBLIC_APP_URL || "https://bitrok.tech";

export const metadata: Metadata = {
  metadataBase: new URL(baseUrl),
  title: {
    default: "Bitrok — Self-Hosted Tunnels",
    template: "%s | Bitrok",
  },
  description:
    "Deterministic tunnels with custom subdomains. Your infra, your URLs, your rules.",
  keywords: [
    "tunnel",
    "localhost",
    "ngrok alternative",
    "self-hosted",
    "reverse proxy",
    "developer tools",
  ],
  authors: [{ name: "Bitrok" }],
  creator: "Bitrok",
  publisher: "Bitrok",
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
  },
  openGraph: {
    type: "website",
    locale: "en_US",
    url: baseUrl,
    siteName: "Bitrok",
    title: "Bitrok — Self-Hosted Tunnels",
    description:
      "Deterministic tunnels with custom subdomains. Your infra, your URLs, your rules.",
    images: [
      {
        url: `${baseUrl}/opengraph-image`,
        width: 1200,
        height: 630,
        alt: "Bitrok — Self-Hosted Tunneling Infrastructure",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Bitrok — Self-Hosted Tunnels",
    description:
      "Deterministic tunnels with custom subdomains. Your infra, your URLs, your rules.",
    images: [`${baseUrl}/opengraph-image`],
    creator: "@Krey_yon",
  },
  alternates: {
    canonical: baseUrl,
  },
  ...(process.env.NEXT_PUBLIC_GOOGLE_SITE_VERIFICATION
    ? { verification: { google: process.env.NEXT_PUBLIC_GOOGLE_SITE_VERIFICATION } }
    : {}),
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className="h-full antialiased"
      suppressHydrationWarning
    >
      <head>
        <link rel="canonical" href={baseUrl} />
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              "@context": "https://schema.org",
              "@type": "Organization",
              name: "Bitrok",
              url: baseUrl,
              logo: `${baseUrl}/icon.svg`,
              sameAs: [
                "https://github.com/krey-yon/Bitrok",
                "https://x.com/Krey_yon",
              ],
              ...(process.env.NEXT_PUBLIC_SUPPORT_EMAIL
                ? {
                    contactPoint: {
                      "@type": "ContactPoint",
                      email: process.env.NEXT_PUBLIC_SUPPORT_EMAIL,
                      contactType: "customer service",
                    },
                  }
                : {}),
            }),
          }}
        />
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              "@context": "https://schema.org",
              "@type": "WebSite",
              name: "Bitrok",
              url: baseUrl,
            }),
          }}
        />
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              "@context": "https://schema.org",
              "@type": "SoftwareApplication",
              name: "Bitrok",
              applicationCategory: "DeveloperApplication",
              operatingSystem: "Cross-platform",
              offers: {
                "@type": "Offer",
                price: "0",
                priceCurrency: "USD",
              },
            }),
          }}
        />
      </head>
      <body className="min-h-full flex flex-col bg-background text-foreground">
        <ThemeProvider>
          <a href="#main-content" className="skip-link">Skip to content</a>
          {children}
        </ThemeProvider>
        {process.env.NEXT_PUBLIC_UMAMI_WEBSITE_ID && (
          <Script
            defer
            src={`${process.env.NEXT_PUBLIC_UMAMI_DOMAIN || "https://cloud.umami.is"}/script.js`}
            data-website-id={process.env.NEXT_PUBLIC_UMAMI_WEBSITE_ID}
            strategy="lazyOnload"
          />
        )}
      </body>
    </html>
  );
}
