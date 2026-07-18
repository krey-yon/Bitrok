"use client";

import { useEffect, useState, Suspense } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Eyebrow } from "@/components/ui/eyebrow";
import { Logo } from "@/components/ui/logo";
import { Spinner } from "@/components/ui/spinner";
import { StatusGlyph } from "@/components/ui/status-glyph";
import { TerminalPanel, TerminalLine } from "@/components/ui/terminal";

type SessionData = { user?: { email?: string; name?: string } } | null;

function CliAuthContent() {
  const searchParams = useSearchParams();
  const state = searchParams.get("state");
  const callback = searchParams.get("callback");

  const [session, setSession] = useState<SessionData>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [authorizing, setAuthorizing] = useState(false);

  useEffect(() => {
    fetch("/api/auth/get-session")
      .then((r) => r.json())
      .then((data) => {
        setSession(data);
        setLoading(false);
      })
      .catch(() => {
        setLoading(false);
      });
  }, []);

  // Session loaded but missing → redirect to sign-in. Done in an effect
  // (not during render) to satisfy react-hooks/immutability.
  useEffect(() => {
    if (loading || session) return;
    const returnUrl = encodeURIComponent(
      `/cli-auth?state=${state || ""}&callback=${encodeURIComponent(callback || "")}`,
    );
    window.location.href = `/login?returnUrl=${returnUrl}`;
  }, [loading, session, state, callback]);

  const handleAuthorize = async () => {
    if (!state) return;
    setAuthorizing(true);
    try {
      const res = await fetch("/api/cli-auth/token", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ state }),
      });
      const data = await res.json();
      if (data.token && callback) {
        window.location.href = `${callback}?token=${encodeURIComponent(data.token)}&state=${encodeURIComponent(state)}`;
      } else if (data.error) {
        setError(data.error);
        setAuthorizing(false);
      }
    } catch {
      setError("Failed to authorize. Please try again.");
      setAuthorizing(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-full flex items-center justify-center">
        <p className="text-sm text-muted font-mono">loading…</p>
      </div>
    );
  }

  if (!session) {
    return (
      <div className="min-h-full flex items-center justify-center">
        <p className="text-sm text-muted font-mono">redirecting to sign in…</p>
      </div>
    );
  }

  return (
    <div className="relative min-h-full flex items-center justify-center px-6 py-16 overflow-hidden">
      <div
        className="absolute inset-0 bg-starfield [mask-image:radial-gradient(60%_50%_at_50%_50%,#000,transparent)] opacity-60"
        aria-hidden
      />
      <div className="relative w-full max-w-sm text-center">
        <div className="flex justify-center mb-4">
          <Link href="/">
            <Logo />
          </Link>
        </div>
        <Eyebrow ornament="·">authorize cli</Eyebrow>
        <h1 className="mt-3 text-2xl font-semibold tracking-tight mb-1">
          Authorize CLI.
        </h1>
        <p className="text-sm text-muted mb-8">
          Allow the bitrok CLI to access your account?
        </p>

        <TerminalPanel title="session" className="mb-8 text-left">
          <TerminalLine status={<StatusGlyph variant="active" pulse />}>
            <span className="text-foreground truncate">
              {session.user?.email || session.user?.name || "Unknown"}
            </span>
          </TerminalLine>
        </TerminalPanel>

        {error && (
          <p className="mb-6 flex items-center justify-center gap-2 text-sm text-danger">
            <StatusGlyph variant="danger" /> {error}
          </p>
        )}

        <div className="flex flex-col gap-3">
          <Button
            className="w-full"
            arrow={!authorizing}
            disabled={authorizing}
            onClick={handleAuthorize}
          >
            {authorizing ? (
              <>
                <Spinner /> authorizing
              </>
            ) : (
              "Authorize"
            )}
          </Button>
          <Button
            variant="ghost"
            className="w-full"
            onClick={() => window.close()}
          >
            Cancel
          </Button>
        </div>

        <p className="mt-8 text-xs text-muted font-mono">
          this generates an api token for cli access
        </p>
      </div>
    </div>
  );
}

export default function CliAuthPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-full flex items-center justify-center">
          <p className="text-sm text-muted font-mono">loading…</p>
        </div>
      }
    >
      <CliAuthContent />
    </Suspense>
  );
}
