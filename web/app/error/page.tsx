"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Suspense } from "react";
import { Button } from "@/components/ui/button";
import { Eyebrow } from "@/components/ui/eyebrow";
import { Logo } from "@/components/ui/logo";
import { StatusGlyph } from "@/components/ui/status-glyph";

const errorMessages: Record<
  string,
  { title: string; message: string; action?: string }
> = {
  email_not_found: {
    title: "Email not found",
    message:
      "We couldn't retrieve an email from your GitHub account. This usually happens when your email is private or not verified.",
    action:
      "Make sure your GitHub account has at least one verified email. You may also need to revoke app access and try again.",
  },
  invalid_code: {
    title: "Authentication failed",
    message: "The authorization code from GitHub was invalid or expired.",
    action: "Please try signing in again.",
  },
  no_callback_url: {
    title: "Authentication failed",
    message: "No callback URL was found for this authentication request.",
    action: "Please try signing in again.",
  },
  oauth_provider_not_found: {
    title: "Authentication failed",
    message: "The OAuth provider configuration could not be found.",
    action: "Please contact support if this persists.",
  },
  unable_to_get_user_info: {
    title: "Authentication failed",
    message: "We couldn't retrieve your account information from GitHub.",
    action: "Please try signing in again.",
  },
  state_not_found: {
    title: "Authentication failed",
    message: "The authentication state was missing or expired.",
    action: "Please try signing in again.",
  },
};

function ErrorPageContent() {
  const searchParams = useSearchParams();
  const errorCode = searchParams.get("error") || "unknown";
  const errorInfo = errorMessages[errorCode] || {
    title: "Something went wrong",
    message: "We encountered an unexpected error during sign in.",
    action: "Please try again.",
  };

  return (
    <div className="relative min-h-full flex items-center justify-center px-6 py-16 overflow-hidden">
      <div
        className="absolute inset-0 bg-starfield [mask-image:radial-gradient(60%_50%_at_50%_50%,#000,transparent)] opacity-60"
        aria-hidden
      />
      <div className="relative w-full max-w-xs text-center">
        <div className="flex justify-center mb-4">
          <Link href="/">
            <Logo />
          </Link>
        </div>
        <div className="flex justify-center mb-4">
          <StatusGlyph variant="danger" className="text-4xl" />
        </div>
        <Eyebrow ornament="·">error</Eyebrow>
        <h1 className="mt-3 text-2xl font-semibold tracking-tight mb-4">
          {errorInfo.title}
        </h1>

        <p className="text-sm text-danger mb-2">{errorInfo.message}</p>
        {errorInfo.action && (
          <p className="text-sm text-muted mb-8">{errorInfo.action}</p>
        )}

        {errorCode === "email_not_found" && (
          <ol className="text-left text-sm text-muted space-y-2 mb-8 list-decimal list-inside font-mono">
            <li>
              Verify a primary email in{" "}
              <a
                href="https://github.com/settings/emails"
                target="_blank"
                rel="noopener noreferrer"
                className="text-accent hover:underline"
              >
                GitHub Email Settings
              </a>
              .
            </li>
            <li>
              Revoke Bitrok in{" "}
              <a
                href="https://github.com/settings/applications"
                target="_blank"
                rel="noopener noreferrer"
                className="text-accent hover:underline"
              >
                Authorized OAuth Apps
              </a>
              .
            </li>
            <li>Return here and sign in again.</li>
          </ol>
        )}

        <div className="flex flex-col gap-3">
          <Link href="/login">
            <Button className="w-full" arrow>
              Try again
            </Button>
          </Link>
          <Link href="/">
            <Button variant="ghost" className="w-full">
              Go home
            </Button>
          </Link>
        </div>
      </div>
    </div>
  );
}

export default function ErrorPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-full flex items-center justify-center">
          <p className="text-sm text-muted font-mono">loading…</p>
        </div>
      }
    >
      <ErrorPageContent />
    </Suspense>
  );
}
