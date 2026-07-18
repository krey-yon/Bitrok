"use client";

import { useState, Suspense } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { authClient } from "@/lib/auth-client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Eyebrow } from "@/components/ui/eyebrow";
import { Logo } from "@/components/ui/logo";
import { Spinner } from "@/components/ui/spinner";
import { StatusGlyph } from "@/components/ui/status-glyph";

function LoginContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const returnUrl = searchParams.get("returnUrl") || "/dashboard";

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const validate = () => {
    if (!email.trim()) return "Email is required";
    if (!/^\S+@\S+\.\S+$/.test(email)) return "Please enter a valid email";
    if (!password) return "Password is required";
    return "";
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const validationError = validate();
    if (validationError) {
      setError(validationError);
      return;
    }

    setLoading(true);
    setError("");

    const { error: signInError } = await authClient.signIn.email({
      email: email.trim(),
      password,
      callbackURL: returnUrl,
    });

    if (signInError) {
      setError(signInError.message || "Invalid credentials");
      setLoading(false);
    } else {
      router.push(returnUrl);
      router.refresh();
    }
  };

  const handleGitHub = async () => {
    setLoading(true);
    await authClient.signIn.social({
      provider: "github",
      callbackURL: returnUrl,
    });
  };

  return (
    <div className="relative min-h-full flex flex-col items-center justify-center px-6 py-16 overflow-hidden">
      <div
        className="absolute inset-0 bg-starfield [mask-image:radial-gradient(60%_50%_at_50%_50%,#000,transparent)] opacity-60"
        aria-hidden
      />
      <div className="relative w-full max-w-xs">
        <div className="flex justify-center mb-6">
          <Link href="/">
            <Logo />
          </Link>
        </div>
        <div className="text-center">
          <Eyebrow ornament="·">sign in</Eyebrow>
          <h1 className="mt-3 text-2xl font-semibold tracking-tight">Sign in.</h1>
          <p className="mt-1.5 mb-10 text-sm text-muted font-mono">
            to continue to your tunnels
          </p>
        </div>

        {error && (
          <p className="mb-6 flex items-center justify-center gap-2 text-sm text-danger text-center">
            <StatusGlyph variant="danger" /> {error}
          </p>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          <Input
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="email"
          />
          <Input
            type="password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="password"
          />
          <Button className="w-full" arrow={!loading} disabled={loading}>
            {loading ? (
              <>
                <Spinner /> signing in
              </>
            ) : (
              "Sign in"
            )}
          </Button>
        </form>

        <div className="flex items-center gap-4 my-8 text-xs text-muted font-mono">
          <span className="flex-1 border-t border-hairline" />
          or
          <span className="flex-1 border-t border-hairline" />
        </div>

        <Button
          variant="ghost"
          className="w-full"
          onClick={handleGitHub}
          disabled={loading}
        >
          <svg
            className="w-4 h-4"
            viewBox="0 0 24 24"
            fill="currentColor"
            aria-hidden
          >
            <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
          </svg>
          Continue with GitHub
        </Button>

        <p className="text-center text-sm text-muted mt-10">
          New here?{" "}
          <Link href="/register" className="text-accent hover:underline">
            Create an account
          </Link>
        </p>
      </div>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-full flex items-center justify-center">
          <p className="text-sm text-muted font-mono">loading…</p>
        </div>
      }
    >
      <LoginContent />
    </Suspense>
  );
}
