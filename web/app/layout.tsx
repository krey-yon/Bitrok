import type { Metadata } from "next";
import { Inter, Space_Grotesk, JetBrains_Mono } from "next/font/google";
import Script from "next/script";
import { ThemeProvider } from "@/components/theme-provider";
import "./globals.css";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  display: "swap",
});

const spaceGrotesk = Space_Grotesk({
  variable: "--font-space-grotesk",
  subsets: ["latin"],
  display: "swap",
});

const jetbrainsMono = JetBrains_Mono({
  variable: "--font-jetbrains-mono",
  subsets: ["latin"],
  display: "swap",
});

const baseUrl = process.env.NEXT_PUBLIC_APP_URL || "https://bitrok.dev";

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
        url: `${baseUrl}/og-image.png`,
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
    images: [`${baseUrl}/og-image.png`],
    creator: "@bitrok",
  },
  alternates: {
    canonical: baseUrl,
  },
  verification: {
    google: process.env.NEXT_PUBLIC_GOOGLE_SITE_VERIFICATION,
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${inter.variable} ${spaceGrotesk.variable} ${jetbrainsMono.variable} h-full antialiased`}
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
              logo: `${baseUrl}/logo.png`,
              sameAs: [
                "https://github.com/bitrok",
                "https://twitter.com/bitrok",
              ],
              contactPoint: {
                "@type": "ContactPoint",
                email:
                  process.env.NEXT_PUBLIC_SUPPORT_EMAIL ||
                  "support@example.com",
                contactType: "customer service",
              },
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
              potentialAction: {
                "@type": "SearchAction",
                target: {
                  "@type": "EntryPoint",
                  urlTemplate: `${baseUrl}/search?q={search_term_string}`,
                },
                "query-input": "required name=search_term_string",
              },
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
