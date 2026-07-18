"use client";

import { Suspense, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { GitHubMark } from "@/components/ui/github-mark";
import { authClient } from "@/lib/auth-client";
import { AuthShell } from "@/app/components/auth-shell";
import { Button } from "@/components/ui/button";
import { Input, Label } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";

function LoginContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const returnUrl = searchParams.get("returnUrl") || "/dashboard";
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!email.trim() || !/^\S+@\S+\.\S+$/.test(email)) return setError("Enter a valid email address.");
    if (!password) return setError("Enter your password.");
    setLoading(true); setError("");
    const { error: signInError } = await authClient.signIn.email({ email: email.trim(), password, callbackURL: returnUrl });
    if (signInError) { setError(signInError.message || "Those credentials did not match. Try again."); setLoading(false); }
    else { router.push(returnUrl); router.refresh(); }
  };

  const handleGitHub = async () => {
    setLoading(true); setError("");
    await authClient.signIn.social({ provider: "github", callbackURL: returnUrl });
  };

  return (
    <AuthShell eyebrow="Welcome back" title="Sign in to your network." description="Manage permanent endpoints, see active tunnels, and connect the CLI.">
      {error && <div role="alert" aria-live="polite" className="mb-6 rounded-lg border border-danger/30 bg-danger/8 px-4 py-3 text-sm text-danger">{error}</div>}
      <form onSubmit={handleSubmit} className="space-y-5">
        <div><Label htmlFor="email">Email address</Label><Input id="email" name="email" type="email" autoComplete="email" spellCheck={false} required value={email} onChange={(e) => setEmail(e.target.value)} placeholder="you@example.com…" className="mt-2" /></div>
        <div><div className="flex items-center justify-between"><Label htmlFor="password">Password</Label></div><Input id="password" name="password" type="password" autoComplete="current-password" required value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Your password…" className="mt-2" /></div>
        <Button type="submit" variant="accent" className="w-full" arrow={!loading} disabled={loading}>{loading ? <><Spinner /> Signing in…</> : "Sign In"}</Button>
      </form>
      <div className="my-7 flex items-center gap-4 text-[11px] uppercase tracking-[.14em] text-muted-foreground"><span className="h-px flex-1 bg-hairline" />or continue with<span className="h-px flex-1 bg-hairline" /></div>
      <Button variant="ghost" className="w-full" onClick={handleGitHub} disabled={loading}><GitHubMark />GitHub</Button>
      <p className="mt-8 text-center text-sm text-muted-foreground">New to Bitrok? <Link href="/register" className="font-medium text-foreground underline decoration-accent decoration-2 underline-offset-4 hover:text-accent">Create an account</Link></p>
    </AuthShell>
  );
}

export default function LoginPage() {
  return <Suspense fallback={<div className="flex min-h-full items-center justify-center" role="status">Loading…</div>}><LoginContent /></Suspense>;
}
