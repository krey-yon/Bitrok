import Link from "next/link";
import { Activity, ArrowUpRight, CircleDot, Clock3, ExternalLink, Network, Plus, Route } from "lucide-react";
import { requireAuth } from "@/lib/session";
import { getServerLogs, getServerTunnels } from "@/lib/server-api";
import { DashboardHeader } from "@/app/components/dashboard-header";
import { SignOutButton } from "./sign-out-button";
import { buttonClassName } from "@/components/ui/button";

export default async function DashboardPage() {
  const session = await requireAuth();
  const [tunnelsResult, logsResult] = await Promise.allSettled([getServerTunnels(session.user.id), getServerLogs(session.user.id, 10)]);
  const tunnels = tunnelsResult.status === "fulfilled" ? tunnelsResult.value : [];
  const logs = logsResult.status === "fulfilled" ? logsResult.value.logs : [];
  const totalRequests = logsResult.status === "fulfilled" ? logsResult.value.total : 0;
  const relayUnavailable = tunnelsResult.status === "rejected" || logsResult.status === "rejected";
  if (tunnelsResult.status === "rejected") console.error("dashboard: failed to load tunnels:", tunnelsResult.reason);
  if (logsResult.status === "rejected") console.error("dashboard: failed to load logs:", logsResult.reason);
  const activeTunnels = tunnels.filter((tunnel) => tunnel.active).length;

  return (
    <div className="min-h-full bg-page-gradient">
      <DashboardHeader email={session.user.email ?? undefined} signOut={<SignOutButton />} />
      <main id="main-content" className="section-shell py-10 sm:py-14">
        <div className="flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
          <div><div className="signal-label">Network overview</div><h1 className="mt-5 text-balance text-4xl font-semibold tracking-[-.045em] sm:text-5xl">Your tunnels.</h1><p className="mt-3 text-muted-foreground">Stable routes from the public internet to your machine.</p></div>
          <Link href="/dashboard/tunnels/new" className={buttonClassName({ variant: "accent", size: "lg" })}><Plus className="size-4" aria-hidden />Create Tunnel</Link>
        </div>

        {relayUnavailable && <div role="status" className="mt-8 flex items-start gap-3 rounded-xl border border-warning/30 bg-warning/8 p-4 text-sm"><CircleDot className="mt-0.5 size-4 shrink-0 text-warning" aria-hidden /><div><strong>Relay data is temporarily unavailable.</strong><p className="mt-1 text-muted-foreground">Your saved tunnels may not appear until the relay reconnects.</p></div></div>}

        <section aria-label="Tunnel statistics" className="mt-10 grid gap-3 sm:grid-cols-3">
          <Stat icon={Network} label="Total tunnels" value={String(tunnels.length)} detail="reserved endpoints" />
          <Stat icon={Activity} label="Live now" value={String(activeTunnels)} detail={activeTunnels === 1 ? "active connection" : "active connections"} active />
          <Stat icon={Route} label="Requests served" value={new Intl.NumberFormat("en").format(totalRequests)} detail="all-time relay traffic" />
        </section>

        <section className="mt-14" aria-labelledby="tunnel-heading">
          <div className="mb-5 flex items-center justify-between"><div><div className="signal-label">Endpoints</div><h2 id="tunnel-heading" className="mt-3 text-2xl font-semibold tracking-tight">Reserved routes</h2></div><span className="font-mono text-xs text-muted-foreground">{activeTunnels}/{tunnels.length} live</span></div>
          {tunnels.length === 0 && !relayUnavailable ? (
            <div className="relative overflow-hidden rounded-[var(--radius-xl)] border border-dashed border-border bg-card/65 px-6 py-16 text-center"><div className="absolute inset-0 bg-grid opacity-50" aria-hidden /><div className="relative"><div className="mx-auto flex size-12 items-center justify-center rounded-xl border border-hairline bg-background text-accent"><Route className="size-5" aria-hidden /></div><h3 className="mt-5 text-xl font-semibold">No tunnel routes yet</h3><p className="mx-auto mt-2 max-w-md text-sm leading-6 text-muted-foreground">Reserve a deterministic URL, map it to a local port, then connect it with the CLI.</p><Link href="/dashboard/tunnels/new" className={buttonClassName({ variant: "accent", className: "mt-6" })}>Create Your First Tunnel <ArrowUpRight className="size-4" aria-hidden /></Link></div></div>
          ) : (
            <div className="grid gap-3 md:grid-cols-2">
              {tunnels.map((tunnel) => <article key={tunnel.id} className="group rounded-xl border border-hairline bg-card/75 p-5 transition-[border-color,background-color,transform] duration-200 hover:-translate-y-0.5 hover:border-foreground/25 hover:bg-card">
                <div className="flex items-start justify-between gap-4"><div className="min-w-0"><div className="flex items-center gap-2"><span className={`size-2 rounded-full ${tunnel.active ? "animate-pulse-dot bg-success" : "bg-muted"}`} /><h3 className="truncate font-semibold">{tunnel.name}</h3></div><p className="mt-1 pl-4 text-xs text-muted-foreground">{tunnel.active ? "Connected & receiving traffic" : "Waiting for CLI connection"}</p></div><span className={`rounded-full border px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[.1em] ${tunnel.active ? "border-success/30 bg-success/10 text-success" : "border-hairline text-muted-foreground"}`}>{tunnel.active ? "Live" : "Offline"}</span></div>
                <div className="mt-5 rounded-lg border border-hairline bg-background/70 p-3"><div className="flex min-w-0 items-center gap-2 font-mono text-xs"><ExternalLink className="size-3.5 shrink-0 text-accent" aria-hidden /><span className="truncate text-accent">https://{tunnel.host}</span></div><div className="mt-2 flex items-center gap-2 border-t border-hairline pt-2 font-mono text-xs text-muted-foreground"><span>→</span><span>localhost:{tunnel.port}</span></div></div>
              </article>)}
            </div>
          )}
        </section>

        <section className="mt-14" aria-labelledby="activity-heading">
          <div className="mb-5 flex items-end justify-between"><div><div className="signal-label">Traffic</div><h2 id="activity-heading" className="mt-3 text-2xl font-semibold tracking-tight">Recent activity</h2></div><div className="hidden items-center gap-2 text-xs text-muted-foreground sm:flex"><Clock3 className="size-3.5" aria-hidden />Last 10 requests</div></div>
          <div className="overflow-x-auto rounded-xl border border-hairline bg-card/75">
            {logs.length === 0 ? <div className="p-8 text-center text-sm text-muted-foreground">Requests will appear here when a tunnel starts receiving traffic.</div> : <table className="w-full min-w-[620px] text-left text-sm"><thead><tr className="border-b border-hairline text-[10px] uppercase tracking-[.13em] text-muted-foreground"><th className="px-4 py-3 font-medium">Status</th><th className="px-4 py-3 font-medium">Method</th><th className="px-4 py-3 font-medium">Path</th><th className="px-4 py-3 text-right font-medium">Latency</th><th className="px-4 py-3 text-right font-medium">Time</th></tr></thead><tbody>{logs.map((log) => <tr key={log.id} className="border-b border-hairline last:border-0 hover:bg-foreground/[0.025]"><td className="px-4 py-3"><span className={`inline-flex items-center gap-2 font-mono text-xs ${log.status >= 400 ? "text-danger" : log.status >= 300 ? "text-warning" : "text-success"}`}><span className="size-1.5 rounded-full bg-current" />{log.status}</span></td><td className="px-4 py-3 font-mono text-xs font-semibold">{log.method}</td><td className="max-w-80 truncate px-4 py-3 font-mono text-xs" title={log.path}>{log.path}</td><td className="px-4 py-3 text-right font-mono text-xs tabular-nums text-muted-foreground">{log.latency_ms}ms</td><td className="px-4 py-3 text-right font-mono text-xs tabular-nums text-muted-foreground">{new Intl.DateTimeFormat("en", { hour: "2-digit", minute: "2-digit", second: "2-digit" }).format(new Date(log.ts))}</td></tr>)}</tbody></table>}
          </div>
        </section>
      </main>
    </div>
  );
}

function Stat({ icon: Icon, label, value, detail, active = false }: { icon: typeof Activity; label: string; value: string; detail: string; active?: boolean }) {
  return <div className="rounded-xl border border-hairline bg-card/75 p-5"><div className="flex items-center justify-between"><span className="text-xs font-medium text-muted-foreground">{label}</span><Icon className={`size-4 ${active ? "text-success" : "text-accent"}`} aria-hidden /></div><div className="mt-5 font-display text-4xl font-semibold tracking-[-.05em] tabular-nums">{value}</div><div className="mt-1 text-xs text-muted-foreground">{detail}</div></div>;
}
