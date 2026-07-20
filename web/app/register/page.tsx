"use client";

import { useState } from "react";
import Link from "next/link";
import { GitHubMark } from "@/components/ui/github-mark";
import { authClient } from "@/lib/auth-client";
import { AuthShell } from "@/app/components/auth-shell";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

export default function RegisterPage() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleGitHub = async () => {
    setLoading(true); setError("");
    const { error: signUpError } = await authClient.signIn.social({ provider: "github", callbackURL: "/dashboard/settings?onboarding=1" });
    if (signUpError) {
      setError(signUpError.message || "GitHub registration could not be started. Try again.");
      setLoading(false);
    }
  };

  return (
    <AuthShell eyebrow="Start free" title="Claim your permanent endpoint." description="Create an account, reserve a project URL, and connect localhost in less than a minute." asideTitle="Configure the frontend once." asideBody="Your deterministic endpoint stays the same across restarts, network changes, and every new demo session.">
      {error && <div role="alert" aria-live="polite" className="mb-6 rounded-lg border border-danger/30 bg-danger/8 px-4 py-3 text-sm text-danger">{error}</div>}
      <Button variant="accent" className="w-full" onClick={handleGitHub} disabled={loading}>
        {loading ? <><Spinner /> Connecting to GitHub…</> : <><GitHubMark />Create account with GitHub</>}
      </Button>
      <p className="mt-5 text-center text-xs leading-5 text-muted-foreground">Bitrok uses your verified GitHub email for account identity.</p>
      <p className="mt-8 text-center text-sm text-muted-foreground">Already have an account? <Link href="/login" className="font-medium text-foreground underline decoration-accent decoration-2 underline-offset-4 hover:text-accent">Sign in</Link></p>
    </AuthShell>
  );
}
