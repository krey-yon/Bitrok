"use client";

import { Suspense, useState } from "react";
import { useSearchParams } from "next/navigation";
import { GitHubMark } from "@/components/ui/github-mark";
import { authClient } from "@/lib/auth-client";
import { AuthShell } from "@/app/components/auth-shell";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { safeReturnPath } from "@/lib/safe-return-path";

function LoginContent() {
  const searchParams = useSearchParams();
  const returnUrl = safeReturnPath(searchParams.get("returnUrl"), "/dashboard");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleGitHub = async () => {
    setLoading(true); setError("");
    const { error: signInError } = await authClient.signIn.social({ provider: "github", callbackURL: returnUrl });
    if (signInError) {
      setError(signInError.message || "GitHub sign-in could not be started. Try again.");
      setLoading(false);
    }
  };

  return (
    <AuthShell eyebrow="Welcome back" title="Sign in to your network." description="Manage permanent endpoints, see active tunnels, and connect the CLI.">
      {error && <div role="alert" aria-live="polite" className="mb-6 rounded-lg border border-danger/30 bg-danger/8 px-4 py-3 text-sm text-danger">{error}</div>}
      <Button variant="accent" className="w-full" onClick={handleGitHub} disabled={loading}>
        {loading ? <><Spinner /> Connecting to GitHub…</> : <><GitHubMark />Continue with GitHub</>}
      </Button>
      <p className="mt-5 text-center text-xs leading-5 text-muted-foreground">New accounts are created automatically after GitHub verifies your identity.</p>
    </AuthShell>
  );
}

export default function LoginPage() {
  return <Suspense fallback={<div className="flex min-h-full items-center justify-center" role="status">Loading…</div>}><LoginContent /></Suspense>;
}
