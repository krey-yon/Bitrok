"use client";

import { Suspense, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { Check, KeyRound, TerminalSquare, UserRound } from "lucide-react";
import { AuthShell } from "@/app/components/auth-shell";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

type SessionData = { user?: { email?: string; name?: string } } | null;

function CliAuthContent() {
  const searchParams = useSearchParams();
  const state = searchParams.get("state");
  const callback = searchParams.get("callback");
  const [session, setSession] = useState<SessionData>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [authorizing, setAuthorizing] = useState(false);

  useEffect(() => { fetch("/api/auth/get-session").then((response) => response.json()).then((data) => { setSession(data); setLoading(false); }).catch(() => setLoading(false)); }, []);
  useEffect(() => { if (loading || session) return; const returnUrl = encodeURIComponent(`/cli-auth?state=${state || ""}&callback=${encodeURIComponent(callback || "")}`); window.location.href = `/login?returnUrl=${returnUrl}`; }, [loading, session, state, callback]);

  const handleAuthorize = async () => {
    if (!state || !callback) { setError("This authorization request is incomplete. Return to the CLI and run bitrok login again."); return; }
    setAuthorizing(true); setError("");
    try { const response = await fetch("/api/cli-auth/token", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ state }) }); const data = await response.json(); if (response.status === 409 && data.requiresUsername) { window.location.href = `/dashboard/settings?onboarding=1&returnUrl=${encodeURIComponent(`/cli-auth?state=${state}&callback=${encodeURIComponent(callback)}`)}`; return; } if (data.token) window.location.href = `${callback}?token=${encodeURIComponent(data.token)}&state=${encodeURIComponent(state)}`; else { setError(data.error || "Authorization failed. Run bitrok login again."); setAuthorizing(false); } }
    catch { setError("The authorization server could not be reached. Check your connection and try again."); setAuthorizing(false); }
  };

  if (loading || !session) return <div className="flex min-h-full items-center justify-center" role="status"><Spinner /> <span className="ml-2 text-sm text-muted-foreground">{loading ? "Checking session…" : "Redirecting to sign in…"}</span></div>;

  return <AuthShell eyebrow="CLI authorization" title="Connect this terminal?" description="Approve a short-lived credential so the Bitrok CLI can create and run tunnels for your account." asideTitle="Your browser is the trust boundary." asideBody="The CLI receives a scoped token only after you confirm the request from an authenticated Bitrok session.">
    <div className="space-y-3 rounded-xl border border-hairline bg-background/65 p-4">
      <div className="flex items-center gap-3"><div className="flex size-9 items-center justify-center rounded-lg bg-accent/10 text-accent"><UserRound className="size-4" aria-hidden /></div><div className="min-w-0"><div className="text-xs text-muted-foreground">Signed in as</div><div className="truncate text-sm font-medium">{session.user?.email || session.user?.name || "Unknown account"}</div></div><Check className="ml-auto size-4 text-success" aria-hidden /></div>
      <div className="flex items-center gap-3 border-t border-hairline pt-3"><div className="flex size-9 items-center justify-center rounded-lg bg-secondary/10 text-secondary"><TerminalSquare className="size-4" aria-hidden /></div><div className="min-w-0"><div className="text-xs text-muted-foreground">Requesting access</div><div className="truncate font-mono text-xs" title={callback || undefined}>{callback || "Bitrok CLI callback missing"}</div></div></div>
    </div>
    <div className="mt-5 flex gap-3 rounded-lg border border-warning/25 bg-warning/8 p-4 text-sm"><KeyRound className="mt-0.5 size-4 shrink-0 text-warning" aria-hidden /><p>This grants CLI access to your tunnel configuration. Approve only if you started <code className="font-mono">bitrok login</code>.</p></div>
    {error && <p role="alert" aria-live="polite" className="mt-5 text-sm text-danger">{error}</p>}
    <div className="mt-6 grid gap-3 sm:grid-cols-2"><Button variant="ghost" onClick={() => window.close()}>Cancel</Button><Button variant="accent" arrow={!authorizing} disabled={authorizing} onClick={handleAuthorize}>{authorizing ? <><Spinner /> Authorizing…</> : "Authorize CLI"}</Button></div>
  </AuthShell>;
}

export default function CliAuthPage() { return <Suspense fallback={<div className="flex min-h-full items-center justify-center" role="status">Loading…</div>}><CliAuthContent /></Suspense>; }
