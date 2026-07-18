"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Check, Globe2, Laptop, Route } from "lucide-react";
import { DashboardHeader } from "@/app/components/dashboard-header";
import { Button } from "@/components/ui/button";
import { Input, Label } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";

export default function NewTunnelPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [host, setHost] = useState("");
  const [port, setPort] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!name.trim()) return setError("Enter a tunnel name.");
    if (!/^[a-zA-Z0-9_-]+$/.test(name)) return setError("Use only letters, numbers, hyphens, and underscores in the tunnel name.");
    if (!host.trim()) return setError("Enter the public hostname reserved for this tunnel.");
    if (!/^[a-zA-Z0-9._-]+$/.test(host)) return setError("Enter a valid hostname using letters, numbers, dots, hyphens, or underscores.");
    const portNumber = Number.parseInt(port, 10);
    if (Number.isNaN(portNumber) || portNumber < 1 || portNumber > 65535) return setError("Enter a local port between 1 and 65,535.");
    setLoading(true); setError("");
    try {
      const response = await fetch("/api/tunnels", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ name: name.trim(), host: host.trim(), port: portNumber }) });
      if (!response.ok) { const data = await response.json(); throw new Error(data.error || "The tunnel could not be created."); }
      router.push("/dashboard"); router.refresh();
    } catch (caught: unknown) { setError(caught instanceof Error ? caught.message : "A network error interrupted tunnel creation. Try again."); setLoading(false); }
  };

  return (
    <div className="min-h-full bg-page-gradient">
      <DashboardHeader />
      <main id="main-content" className="section-shell py-10 sm:py-14">
        <Link href="/dashboard" className="inline-flex items-center gap-2 rounded-md text-sm text-muted-foreground transition-colors hover:text-foreground"><ArrowLeft className="size-4" aria-hidden />Back to Overview</Link>
        <div className="mt-8 grid gap-8 lg:grid-cols-[.9fr_1.1fr] lg:gap-14">
          <section><div className="signal-label">New route</div><h1 className="mt-5 text-balance text-4xl font-semibold tracking-[-.045em] sm:text-5xl">Connect a local service.</h1><p className="mt-4 max-w-lg text-pretty leading-7 text-muted-foreground">Reserve a public hostname and tell Bitrok which local port should receive its traffic.</p>
            <div className="mt-10 hidden space-y-6 lg:block">{[{ icon: Globe2, title: "Public endpoint", body: "A stable HTTPS address for your frontend and integrations." }, { icon: Route, title: "Encrypted relay", body: "Requests cross the authenticated CLI tunnel to your machine." }, { icon: Laptop, title: "Local service", body: "Your app continues running on localhost with no deployment." }].map(({ icon: Icon, title, body }, index) => <div key={title} className="flex gap-4"><div className="flex size-9 shrink-0 items-center justify-center rounded-lg border border-hairline bg-card text-accent"><Icon className="size-4" aria-hidden /></div><div><div className="flex items-center gap-2"><span className="font-mono text-[10px] text-muted-foreground">0{index + 1}</span><h2 className="font-semibold">{title}</h2></div><p className="mt-1 text-sm leading-6 text-muted-foreground">{body}</p></div></div>)}</div>
          </section>
          <section className="rounded-[var(--radius-xl)] border border-hairline bg-card/80 p-6 shadow-[0_24px_70px_rgb(0_0_0/8%)] sm:p-8">
            <h2 className="text-xl font-semibold">Tunnel details</h2><p className="mt-1 text-sm text-muted-foreground">You can start the connection after saving.</p>
            {error && <div role="alert" aria-live="polite" className="mt-6 rounded-lg border border-danger/30 bg-danger/8 px-4 py-3 text-sm text-danger">{error}</div>}
            <form onSubmit={handleSubmit} className="mt-7 space-y-6">
              <div><Label htmlFor="name">Tunnel name</Label><Input id="name" name="name" autoComplete="off" spellCheck={false} required maxLength={100} value={name} onChange={(e) => setName(e.target.value)} placeholder="excali-api…" className="mt-2" /><p className="mt-2 text-xs text-muted-foreground">Used by the CLI: <code className="font-mono text-foreground">bitrok up {name || "excali-api"}</code></p></div>
              <div><Label htmlFor="host">Public hostname</Label><Input id="host" name="host" autoComplete="off" spellCheck={false} required maxLength={255} value={host} onChange={(e) => setHost(e.target.value.toLowerCase())} placeholder="excali-kreyon.bitrok.tech…" className="mt-2" /><p className="mt-2 text-xs text-muted-foreground">This is the permanent address your frontend will use.</p></div>
              <div><Label htmlFor="port">Local port</Label><Input id="port" name="port" type="number" inputMode="numeric" autoComplete="off" required min={1} max={65535} value={port} onChange={(e) => setPort(e.target.value)} placeholder="3000…" className="mt-2" /></div>
              <div className="rounded-lg border border-hairline bg-background/70 p-4"><div className="flex items-center gap-2 text-xs font-medium"><Check className="size-3.5 text-success" aria-hidden />Route preview</div><div className="mt-3 space-y-2 font-mono text-xs"><div className="truncate text-accent">https://{host || "your-project.bitrok.tech"}</div><div className="text-muted-foreground">↓ encrypted tunnel</div><div>http://localhost:{port || "3000"}</div></div></div>
              <Button type="submit" variant="accent" className="w-full" arrow={!loading} disabled={loading}>{loading ? <><Spinner /> Creating tunnel…</> : "Create Tunnel"}</Button>
            </form>
          </section>
        </div>
      </main>
    </div>
  );
}
