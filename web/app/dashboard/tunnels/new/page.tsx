"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Eyebrow } from "@/components/ui/eyebrow";
import { Logo } from "@/components/ui/logo";
import { Spinner } from "@/components/ui/spinner";
import { StatusGlyph } from "@/components/ui/status-glyph";

export default function NewTunnelPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [host, setHost] = useState("");
  const [port, setPort] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const validate = () => {
    if (!name.trim()) return "Name is required";
    if (name.length > 100) return "Name must be 100 characters or less";
    if (!host.trim()) return "Host is required";
    if (host.length > 255) return "Host must be 255 characters or less";
    const portNum = parseInt(port, 10);
    if (Number.isNaN(portNum) || portNum < 1 || portNum > 65535) {
      return "Port must be a number between 1 and 65535";
    }
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

    try {
      const res = await fetch("/api/tunnels", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: name.trim(),
          host: host.trim(),
          port: parseInt(port, 10),
        }),
      });

      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || "Failed to create tunnel");
      }

      router.push("/dashboard");
      router.refresh();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Something went wrong";
      setError(message);
      setLoading(false);
    }
  };

  return (
    <div className="min-h-full flex flex-col">
      <nav className="sticky top-0 z-50 border-b border-hairline bg-background/80 backdrop-blur">
        <div className="max-w-3xl mx-auto px-6 h-12 flex items-center text-sm">
          <Link href="/dashboard" className="font-mono">
            <Logo />
          </Link>
        </div>
      </nav>

      <main className="flex-1 max-w-xs mx-auto px-6 py-14 w-full">
        <Link
          href="/dashboard"
          className="text-sm text-muted hover:text-foreground transition-colors font-mono"
        >
          ← tunnels
        </Link>

        <div className="mt-4">
          <Eyebrow ornament="·">new tunnel</Eyebrow>
        </div>
        <h1 className="mt-3 text-2xl font-semibold tracking-tight mb-1">
          New tunnel.
        </h1>
        <p className="text-sm text-muted mb-10 font-mono">
          then run <span className="text-accent">bitrok up {name || "…"}</span>
        </p>

        {error && (
          <p className="mb-6 flex items-center gap-2 text-sm text-danger">
            <StatusGlyph variant="danger" /> {error}
          </p>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          <Input
            type="text"
            required
            maxLength={100}
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="name — my-api"
          />
          <Input
            type="text"
            required
            maxLength={255}
            value={host}
            onChange={(e) => setHost(e.target.value)}
            placeholder="host — api.myapp.bitrok.tech"
          />
          <Input
            type="number"
            required
            min={1}
            max={65535}
            value={port}
            onChange={(e) => setPort(e.target.value)}
            placeholder="local port — 3000"
          />
          <Button className="w-full" arrow={!loading} disabled={loading}>
            {loading ? (
              <>
                <Spinner /> creating
              </>
            ) : (
              "Create tunnel"
            )}
          </Button>
        </form>
      </main>
    </div>
  );
}
