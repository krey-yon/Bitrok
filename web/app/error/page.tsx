"use client";

import { Suspense } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { ArrowRight, CircleAlert, ExternalLink } from "lucide-react";
import { AuthShell } from "@/app/components/auth-shell";
import { buttonClassName } from "@/components/ui/button";

const errorMessages: Record<string, { title: string; message: string; action?: string }> = {
  email_not_found: { title: "GitHub did not share a verified email.", message: "Bitrok needs a verified email address to create your account.", action: "Verify at least one email in GitHub, revoke the existing Bitrok authorization, then try again." },
  invalid_code: { title: "That sign-in link expired.", message: "GitHub returned an invalid or expired authorization code.", action: "Start a fresh sign-in request." },
  no_callback_url: { title: "The sign-in request is incomplete.", message: "No callback URL was attached to this authentication request.", action: "Return to the login page and try again." },
  oauth_provider_not_found: { title: "GitHub sign-in is unavailable.", message: "The OAuth provider is not configured correctly.", action: "Contact support if the problem continues." },
  service_unavailable: { title: "The dashboard is temporarily unavailable.", message: "Bitrok could not reach the session database.", action: "Wait a moment, then try again. Your tunnels and account data are unaffected." },
  unable_to_get_user_info: { title: "GitHub profile access failed.", message: "Bitrok could not retrieve the account information required to sign you in.", action: "Check GitHub's availability, then start a fresh sign-in request." },
  state_not_found: { title: "The sign-in session expired.", message: "The authentication state is missing or no longer valid.", action: "Start a fresh sign-in request." },
};

function ErrorContent() {
  const code = useSearchParams().get("error") || "unknown";
  const info = errorMessages[code] || { title: "Something interrupted sign-in.", message: "Bitrok encountered an unexpected authentication error.", action: "Try signing in again." };
  return <AuthShell eyebrow="Authentication error" title={info.title} description={info.message} asideTitle="Your tunnels are still safe." asideBody="A failed authentication attempt does not expose tunnel credentials or change your reserved endpoints.">
    <div role="alert" className="flex gap-4 rounded-xl border border-danger/25 bg-danger/8 p-4"><CircleAlert className="mt-0.5 size-5 shrink-0 text-danger" aria-hidden /><div><strong className="text-sm">What to do next</strong><p className="mt-1 text-sm leading-6 text-muted-foreground">{info.action}</p></div></div>
    {code === "email_not_found" && <ol className="mt-5 space-y-3 rounded-xl border border-hairline bg-background/65 p-5 text-sm text-muted-foreground"><li className="flex gap-3"><span className="font-mono text-accent">01</span><span>Verify a primary email in <a href="https://github.com/settings/emails" target="_blank" rel="noopener noreferrer" className="font-medium text-foreground underline decoration-accent underline-offset-4">GitHub Email Settings <ExternalLink className="inline size-3" aria-hidden /></a>.</span></li><li className="flex gap-3"><span className="font-mono text-accent">02</span><span>Revoke Bitrok under GitHub&apos;s authorized OAuth apps.</span></li><li className="flex gap-3"><span className="font-mono text-accent">03</span><span>Return here and start a fresh sign-in.</span></li></ol>}
    <div className="mt-6 grid gap-3 sm:grid-cols-2"><Link href="/" className={buttonClassName({ variant: "ghost" })}>Return Home</Link><Link href="/login" className={buttonClassName({ variant: "accent" })}>Try Sign In Again <ArrowRight className="size-4" aria-hidden /></Link></div>
  </AuthShell>;
}

export default function ErrorPage() { return <Suspense fallback={<div className="flex min-h-full items-center justify-center" role="status">Loading…</div>}><ErrorContent /></Suspense>; }
