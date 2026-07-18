"use client";

import { useState } from "react";
import { Copy, Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Eyebrow } from "@/components/ui/eyebrow";
import { Spinner } from "@/components/ui/spinner";
import { StatusGlyph } from "@/components/ui/status-glyph";
import { TerminalPanel } from "@/components/ui/terminal";

type GenerateResponse = {
  token?: string;
  expiresAt?: string;
  error?: string;
};

export function CliTokenGenerator() {
  const [token, setToken] = useState<string | null>(null);
  const [expiresAt, setExpiresAt] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const handleGenerate = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch("/api/cli-auth/generate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
      });
      const data = (await res.json()) as GenerateResponse;
      if (!res.ok || !data.token) {
        setError(data.error || "Failed to generate token");
        return;
      }
      setToken(data.token);
      setExpiresAt(data.expiresAt || null);
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = async () => {
    if (!token) return;
    try {
      await navigator.clipboard.writeText(token);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      setError("Failed to copy to clipboard");
    }
  };

  const expiryLabel = expiresAt
    ? new Date(expiresAt).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
      })
    : null;

  return (
    <div>
      {/* Generate / token block */}
      {!token ? (
        <div className="border-t border-b border-hairline py-8 space-y-6">
          <p className="text-sm text-muted">
            A JWT scoped to your account. Previous tokens stay valid until they
            expire.
          </p>
          <Button arrow disabled={loading} onClick={handleGenerate}>
            {loading ? (
              <>
                <Spinner /> Generating…
              </>
            ) : (
              "Generate token"
            )}
          </Button>
          {error && (
            <p className="flex items-center gap-2 text-sm text-danger">
              <StatusGlyph variant="danger" /> {error}
            </p>
          )}
        </div>
      ) : (
        <div className="border-t border-b border-hairline py-8 space-y-4">
          <p className="text-sm text-accent font-mono">✓ copy this token now.</p>
          <p className="text-sm text-muted">
            This is the only time you&apos;ll see it.
            {expiryLabel && <> Expires {expiryLabel}.</>}
          </p>
          <TerminalPanel title="token · copy once" className="text-left">
            <div className="relative">
              <pre className="text-xs font-mono break-all whitespace-pre-wrap pr-20">
                {token}
              </pre>
              <button
                onClick={handleCopy}
                className="absolute top-1 right-1 inline-flex items-center gap-1.5 text-xs text-muted hover:text-foreground border border-hairline rounded-[var(--radius)] px-2.5 py-1 transition-colors font-mono"
              >
                {copied ? (
                  <>
                    <Check className="w-3 h-3" /> Copied
                  </>
                ) : (
                  <>
                    <Copy className="w-3 h-3" /> Copy
                  </>
                )}
              </button>
            </div>
          </TerminalPanel>
          <button
            onClick={() => {
              setToken(null);
              setExpiresAt(null);
              setCopied(false);
              setError(null);
            }}
            className="text-sm text-accent font-mono hover:underline"
          >
            › generate another
          </button>
          {error && (
            <p className="flex items-center gap-2 text-sm text-danger">
              <StatusGlyph variant="danger" /> {error}
            </p>
          )}
        </div>
      )}

      {/* Usage */}
      <div className="mt-14 mb-3">
        <Eyebrow ornament="·">using this token</Eyebrow>
      </div>
      <div className="space-y-6">
        <div>
          <p className="text-sm font-medium mb-3">Interactive</p>
          <TerminalPanel title="interactive">
            <pre className="text-xs font-mono">bitrok login</pre>
          </TerminalPanel>
          <p className="text-xs text-muted mt-2 font-mono">
            opens the browser to this page, then paste the token when prompted
          </p>
        </div>
        <div>
          <p className="text-sm font-medium mb-3">Headless / CI</p>
          <TerminalPanel title="headless · ci">
            <pre className="text-xs font-mono break-all whitespace-pre-wrap">
              {`export BITROK_TOKEN=${token ?? "<paste-your-token>"}\nbitrok auth --server https://your-server.example.com`}
            </pre>
          </TerminalPanel>
          <p className="text-xs text-muted mt-2 font-mono">
            best for ci or ssh sessions. never commit the token to your repo
          </p>
        </div>
      </div>
    </div>
  );
}
