"use client";

import { useState } from "react";
import { Check, Copy, KeyRound, RefreshCw, TerminalSquare, TriangleAlert } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

type GenerateResponse = { token?: string; expiresAt?: string; error?: string };

export function CliTokenGenerator() {
  const [token, setToken] = useState<string | null>(null);
  const [expiresAt, setExpiresAt] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const handleGenerate = async () => {
    setLoading(true); setError(null);
    try {
      const response = await fetch("/api/cli-auth/generate", { method: "POST", headers: { "Content-Type": "application/json" } });
      const data = (await response.json()) as GenerateResponse;
      if (response.status === 409) { window.location.href = "/dashboard/settings?onboarding=1&returnUrl=/dashboard/cli-token"; return; }
      if (!response.ok || !data.token) { setError(data.error || "The token could not be generated. Try again."); return; }
      setToken(data.token); setExpiresAt(data.expiresAt || null);
    } catch { setError("The server could not be reached. Check your connection and try again."); }
    finally { setLoading(false); }
  };

  const handleCopy = async () => {
    if (!token) return;
    try { await navigator.clipboard.writeText(token); setCopied(true); window.setTimeout(() => setCopied(false), 2000); }
    catch { setError("Clipboard access failed. Select and copy the token manually."); }
  };

  const expiryLabel = expiresAt ? new Intl.DateTimeFormat(undefined, { year: "numeric", month: "short", day: "numeric" }).format(new Date(expiresAt)) : null;

  if (!token) return <div><div className="flex size-11 items-center justify-center rounded-xl border border-hairline bg-background text-accent"><KeyRound className="size-5" aria-hidden /></div><h2 className="mt-5 text-2xl font-semibold tracking-tight">Generate a CLI token</h2><p className="mt-2 max-w-lg text-sm leading-6 text-muted-foreground">The token is scoped to your account. Existing tokens remain valid until they expire.</p><div className="mt-6 rounded-lg border border-warning/25 bg-warning/8 p-4 text-sm"><div className="flex gap-3"><TriangleAlert className="mt-0.5 size-4 shrink-0 text-warning" aria-hidden /><p><strong>Store it safely.</strong> The full value appears only once and grants access to your tunnels.</p></div></div>{error && <p role="alert" aria-live="polite" className="mt-5 text-sm text-danger">{error}</p>}<Button variant="accent" className="mt-6" arrow={!loading} disabled={loading} onClick={handleGenerate}>{loading ? <><Spinner /> Generating token…</> : "Generate Token"}</Button><Usage /></div>;

  return <div><div aria-live="polite" className="flex items-start justify-between gap-4"><div><div className="inline-flex items-center gap-2 text-sm font-semibold text-success"><Check className="size-4" aria-hidden />Token generated</div><h2 className="mt-3 text-2xl font-semibold tracking-tight">Copy it before you leave.</h2></div>{expiryLabel && <span className="rounded-full border border-hairline px-3 py-1.5 text-xs text-muted-foreground">Expires {expiryLabel}</span>}</div><div className="relative mt-6 overflow-hidden rounded-xl border border-border bg-[#0c0f0a] p-5 pr-20 text-[#eff5e5]"><pre className="max-h-40 overflow-auto whitespace-pre-wrap break-all font-mono text-xs leading-6">{token}</pre><button type="button" onClick={handleCopy} aria-label={copied ? "Token copied" : "Copy CLI token"} className="absolute right-3 top-3 inline-flex h-9 items-center gap-1.5 rounded-md border border-white/15 bg-white/8 px-3 text-xs transition-colors hover:bg-white/14 focus-visible:outline-2 focus-visible:outline-accent">{copied ? <Check className="size-3.5 text-[#b8f34a]" aria-hidden /> : <Copy className="size-3.5" aria-hidden />}{copied ? "Copied" : "Copy"}</button></div>{error && <p role="alert" aria-live="polite" className="mt-4 text-sm text-danger">{error}</p>}<button type="button" onClick={() => { setToken(null); setExpiresAt(null); setCopied(false); setError(null); }} className="mt-5 inline-flex items-center gap-2 rounded-md text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"><RefreshCw className="size-3.5" aria-hidden />Generate another token</button><Usage token={token} /></div>;
}

function Usage({ token }: { token?: string | null }) {
  return <div className="mt-10 border-t border-hairline pt-8"><div className="flex items-center gap-2"><TerminalSquare className="size-4 text-accent" aria-hidden /><h3 className="font-semibold">Use it with the CLI</h3></div><div className="mt-4 space-y-3"><CodeBlock label="Interactive" code="bitrok login" /><CodeBlock label="Headless / CI" code={`export BITROK_TOKEN=${token ?? "<your-token>"}\nbitrok auth --server https://your-server.example.com`} /></div><p className="mt-4 text-xs leading-5 text-muted-foreground">Never commit a token to source control or expose it in client-side environment variables.</p></div>;
}

function CodeBlock({ label, code }: { label: string; code: string }) {
  return <div className="overflow-hidden rounded-lg border border-hairline bg-background"><div className="border-b border-hairline px-4 py-2 font-mono text-[10px] uppercase tracking-[.13em] text-muted-foreground">{label}</div><pre className="overflow-x-auto p-4 font-mono text-xs leading-6"><span className="text-accent">$ </span>{code}</pre></div>;
}
