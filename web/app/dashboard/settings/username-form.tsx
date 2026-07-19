"use client";

import { useMemo, useState } from "react";
import { AtSign, Check, Link2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input, Label } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";

type Props = {
  initialUsername: string | null;
};

export function UsernameForm({ initialUsername }: Props) {
  const [value, setValue] = useState(initialUsername ?? "");
  const [saved, setSaved] = useState(initialUsername);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const previewSlug = useMemo(() => {
    return value
      .toLowerCase()
      .trim()
      .replace(/[_\s.]+/g, "-")
      .replace(/[^a-z0-9-]/g, "")
      .replace(/-+/g, "-")
      .replace(/^-|-$/g, "")
      .slice(0, 32);
  }, [value]);

  const dirty = (saved ?? "") !== previewSlug;

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError("");
    setSuccess("");
    if (!previewSlug) {
      setError("Enter a username using letters, numbers, or hyphens.");
      return;
    }
    setLoading(true);
    try {
      const res = await fetch("/api/username", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username: previewSlug }),
      });
      const data = (await res.json()) as {
        username?: string;
        error?: string;
        preview?: string;
      };
      if (!res.ok || !data.username) {
        setError(data.error || "Could not save username.");
        return;
      }
      setSaved(data.username);
      setValue(data.username);
      setSuccess(
        saved
          ? "Username updated. Generate a new CLI token so tunnels use the new slug."
          : "Username saved. Your tunnels will use this slug.",
      );
    } catch {
      setError("Network error. Try again.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div>
        <Label htmlFor="username">Username</Label>
        <p className="mt-1 text-xs text-muted-foreground">
          Public slug in every tunnel URL. Letters, numbers, hyphens — 2–32 characters.
        </p>
        <div className="relative mt-3">
          <AtSign
            className="pointer-events-none absolute left-3.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
            aria-hidden
          />
          <Input
            id="username"
            name="username"
            autoComplete="username"
            spellCheck={false}
            maxLength={32}
            value={value}
            onChange={(e) => {
              setValue(e.target.value);
              setError("");
              setSuccess("");
            }}
            placeholder="kreyon"
            className="pl-10 font-mono"
            aria-describedby="username-preview"
          />
        </div>
      </div>

      <div
        id="username-preview"
        className="rounded-lg border border-hairline bg-background/70 p-4"
      >
        <div className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
          <Link2 className="size-3.5 text-accent" aria-hidden />
          Tunnel URL preview
        </div>
        <p className="mt-2 truncate font-mono text-sm text-accent">
          https://myapp-{previewSlug || "you"}.bitrok.tech
        </p>
        {saved && (
          <p className="mt-2 text-xs text-muted-foreground">
            Current: <span className="font-mono text-foreground">{saved}</span>
          </p>
        )}
      </div>

      {error && (
        <div
          role="alert"
          aria-live="polite"
          className="rounded-lg border border-danger/30 bg-danger/8 px-4 py-3 text-sm text-danger"
        >
          {error}
        </div>
      )}
      {success && (
        <div
          role="status"
          aria-live="polite"
          className="flex items-start gap-2 rounded-lg border border-success/30 bg-success/8 px-4 py-3 text-sm text-success"
        >
          <Check className="mt-0.5 size-4 shrink-0" aria-hidden />
          <span>{success}</span>
        </div>
      )}

      <Button
        type="submit"
        variant="accent"
        disabled={loading || !dirty || !previewSlug}
        arrow={!loading}
      >
        {loading ? (
          <>
            <Spinner /> Saving…
          </>
        ) : saved ? (
          "Update username"
        ) : (
          "Create username"
        )}
      </Button>
    </form>
  );
}
