"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { GitHubMark } from "@/components/ui/github-mark";
import { authClient } from "@/lib/auth-client";
import { AuthShell } from "@/app/components/auth-shell";
import { Button } from "@/components/ui/button";
import { Input, Label } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";

export default function RegisterPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!name.trim()) return setError("Enter your name.");
    if (name.length > 100) return setError("Keep your name under 100 characters.");
    if (!email.trim() || !/^\S+@\S+\.\S+$/.test(email)) return setError("Enter a valid email address.");
    if (password.length < 8) return setError("Use at least 8 characters for your password.");
    setLoading(true); setError("");
    const { error: signUpError } = await authClient.signUp.email({ name: name.trim(), email: email.trim(), password, callbackURL: "/dashboard/settings?onboarding=1" });
    if (signUpError) { setError(signUpError.message || "We could not create your account. Try again."); setLoading(false); }
    else { router.push("/dashboard/settings?onboarding=1"); router.refresh(); }
  };

  const handleGitHub = async () => {
    setLoading(true); setError("");
    await authClient.signIn.social({ provider: "github", callbackURL: "/dashboard/settings?onboarding=1" });
  };

  return (
    <AuthShell eyebrow="Start free" title="Claim your permanent endpoint." description="Create an account, reserve a project URL, and connect localhost in less than a minute." asideTitle="Configure the frontend once." asideBody="Your deterministic endpoint stays the same across restarts, network changes, and every new demo session.">
      {error && <div role="alert" aria-live="polite" className="mb-6 rounded-lg border border-danger/30 bg-danger/8 px-4 py-3 text-sm text-danger">{error}</div>}
      <form onSubmit={handleSubmit} className="space-y-5">
        <div><Label htmlFor="name">Your name</Label><Input id="name" name="name" type="text" autoComplete="name" required maxLength={100} value={name} onChange={(e) => setName(e.target.value)} placeholder="Kreyon…" className="mt-2" /></div>
        <div><Label htmlFor="email">Email address</Label><Input id="email" name="email" type="email" autoComplete="email" spellCheck={false} required value={email} onChange={(e) => setEmail(e.target.value)} placeholder="you@example.com…" className="mt-2" /></div>
        <div><Label htmlFor="password">Password</Label><Input id="password" name="password" type="password" autoComplete="new-password" required minLength={8} aria-describedby="password-hint" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="At least 8 characters…" className="mt-2" /><p id="password-hint" className="mt-2 text-xs text-muted-foreground">Use 8 or more characters.</p></div>
        <Button type="submit" variant="accent" className="w-full" arrow={!loading} disabled={loading}>{loading ? <><Spinner /> Creating account…</> : "Create Account"}</Button>
      </form>
      <div className="my-7 flex items-center gap-4 text-[11px] uppercase tracking-[.14em] text-muted-foreground"><span className="h-px flex-1 bg-hairline" />or continue with<span className="h-px flex-1 bg-hairline" /></div>
      <Button variant="ghost" className="w-full" onClick={handleGitHub} disabled={loading}><GitHubMark />GitHub</Button>
      <p className="mt-8 text-center text-sm text-muted-foreground">Already have an account? <Link href="/login" className="font-medium text-foreground underline decoration-accent decoration-2 underline-offset-4 hover:text-accent">Sign in</Link></p>
    </AuthShell>
  );
}
