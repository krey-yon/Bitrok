import Link from "next/link";
import { redirect } from "next/navigation";
import {
  Activity,
  CircleDot,
  Clock3,
  ExternalLink,
  Network,
  Route,
  Terminal,
} from "lucide-react";
import { requireAuth } from "@/lib/session";
import { getServerLogs, getServerTunnels } from "@/lib/server-api";
import { getUsernameForUser } from "@/lib/username";
import { dashboardView, type DashboardView } from "@/lib/dashboard-view";
import { DashboardHeader } from "@/app/components/dashboard-header";
import { SignOutButton } from "./sign-out-button";
import { buttonClassName } from "@/components/ui/button";
import { CopyButton } from "@/components/ui/copy-button";

const INSTALL_CMD = "curl -fsSL bitrok.tech/install | sh";

export default async function DashboardPage() {
  const session = await requireAuth();
  const [tunnelsResult, logsResult, username] = await Promise.all([
    getServerTunnels(session.user.id).then(
      (v) => ({ status: "fulfilled" as const, value: v }),
      (reason) => ({ status: "rejected" as const, reason }),
    ),
    getServerLogs(session.user.id, 50).then(
      (v) => ({ status: "fulfilled" as const, value: v }),
      (reason) => ({ status: "rejected" as const, reason }),
    ),
    getUsernameForUser(session.user.id),
  ]);

  if (!username) redirect("/dashboard/settings?onboarding=1");
  if (tunnelsResult.status === "rejected") console.error("dashboard: failed to load tunnels:", tunnelsResult.reason);
  if (logsResult.status === "rejected") console.error("dashboard: failed to load logs:", logsResult.reason);

  const view = dashboardView(username, tunnelsResult, logsResult);

  return (
    <div className="min-h-full bg-page-gradient">
      <DashboardHeader
        email={session.user.email ?? undefined}
        username={username}
        signOut={<SignOutButton />}
      />
      <main id="main-content" className="section-shell py-10 sm:py-14">
        <DashboardMain view={view} />
      </main>
    </div>
  );
}

function DashboardMain({ view }: { view: DashboardView }) {
  switch (view.kind) {
    case "first-run":
      return <FirstRunShell username={view.username} />;
    case "relay-down":
      return <RelayDownShell />;
    case "operating":
      return <OperatingShell view={view} />;
  }
}

function FirstRunShell({ username }: { username: string }) {
  const host = `myapp-${username}.bitrok.tech`;
  const hostUrl = `https://${host}`;
  const startCmd = "bitrok myapp 3000";

  return (
    <>
      <div>
        <div className="signal-label">First route</div>
        <h1 className="mt-5 text-balance text-4xl font-semibold tracking-[-.045em] sm:text-5xl">
          Carve a public URL.
        </h1>
        <p className="mt-3 max-w-xl text-muted-foreground">
          Hosts are deterministic. Install the CLI, log in, then expose a local port.
        </p>
      </div>

      <section
        className="mt-10 rounded-[var(--radius-xl)] border border-hairline bg-card/75 p-6 sm:p-8"
        aria-labelledby="first-run-heading"
      >
        <div className="signal-label" id="first-run-heading">
          Setup
        </div>
        <div className="mt-5 rounded-lg border border-hairline bg-background/70 px-4 py-3">
          <div className="font-mono text-[10px] uppercase tracking-[.15em] text-muted-foreground">
            Your first host
          </div>
          <div className="mt-2 flex min-w-0 items-center gap-2">
            <span className="min-w-0 truncate font-mono text-sm text-accent">{hostUrl}</span>
            <CopyButton text={hostUrl} className="shrink-0" />
          </div>
        </div>

        <div className="mt-5 space-y-2">
          <CommandCopy label="1. Install" command={INSTALL_CMD} />
          <CommandCopy label="2. Log in" command="bitrok login" />
          <CommandCopy label="3. Start a tunnel" command={startCmd} />
          <CommandCopy label="4. Hit the URL" command={`curl ${hostUrl}`} />
        </div>

        <p className="mt-4 font-mono text-xs text-muted-foreground">
          {startCmd} → {hostUrl} → localhost:3000
        </p>

        <Link
          href="/dashboard/cli-token"
          className={buttonClassName({ variant: "accent", size: "lg", className: "mt-6" })}
        >
          <Terminal className="size-4" aria-hidden />
          Get CLI token
        </Link>
      </section>
    </>
  );
}

function RelayDownShell() {
  return (
    <>
      <div className="flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <div className="signal-label">Network overview</div>
          <h1 className="mt-5 text-balance text-4xl font-semibold tracking-[-.045em] sm:text-5xl">
            Your tunnels.
          </h1>
          <p className="mt-3 text-muted-foreground">
            Stable routes from the public internet to your machine. Create them with the CLI.
          </p>
        </div>
        <Link
          href="/dashboard/cli-token"
          className={buttonClassName({ variant: "accent", size: "lg" })}
        >
          <Terminal className="size-4" aria-hidden />
          CLI setup
        </Link>
      </div>

      <div
        role="status"
        className="mt-8 flex items-start gap-3 rounded-xl border border-warning/30 bg-warning/8 p-4 text-sm"
      >
        <CircleDot className="mt-0.5 size-4 shrink-0 text-warning" aria-hidden />
        <div>
          <strong>Relay data is temporarily unavailable.</strong>
          <p className="mt-1 text-muted-foreground">
            Your tunnels may not appear until the relay reconnects.
          </p>
        </div>
      </div>
    </>
  );
}

function OperatingShell({ view }: { view: Extract<DashboardView, { kind: "operating" }> }) {
  const { tunnels, logs, totalRequests, logsUnavailable } = view;
  const activeTunnels = tunnels.filter((tunnel) => tunnel.active).length;
  const tunnelNameById = new Map(tunnels.map((t) => [t.id, t.name]));

  return (
    <>
      <div className="flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <div className="signal-label">Network overview</div>
          <h1 className="mt-5 text-balance text-4xl font-semibold tracking-[-.045em] sm:text-5xl">
            Your tunnels.
          </h1>
          <p className="mt-3 text-muted-foreground">
            Stable routes from the public internet to your machine. Create them with the CLI.
          </p>
        </div>
        <Link
          href="/dashboard/cli-token"
          className={buttonClassName({ variant: "accent", size: "lg" })}
        >
          <Terminal className="size-4" aria-hidden />
          CLI setup
        </Link>
      </div>

      <section aria-label="Tunnel statistics" className="mt-10 grid gap-3 sm:grid-cols-3">
        <Stat icon={Network} label="Total tunnels" value={String(tunnels.length)} detail="reserved endpoints" />
        <Stat
          icon={Activity}
          label="Live now"
          value={String(activeTunnels)}
          detail={activeTunnels === 1 ? "active connection" : "active connections"}
          active
        />
        <Stat
          icon={Route}
          label="Requests served"
          value={logsUnavailable ? "—" : new Intl.NumberFormat("en").format(totalRequests)}
          detail={logsUnavailable ? "logs unreachable" : "retained relay traffic"}
        />
      </section>

      <section className="mt-14" aria-labelledby="tunnel-heading">
        <div className="mb-5 flex items-center justify-between">
          <div>
            <div className="signal-label">Endpoints</div>
            <h2 id="tunnel-heading" className="mt-3 text-2xl font-semibold tracking-tight">
              Active routes
            </h2>
          </div>
          <span className="font-mono text-xs text-muted-foreground">
            {activeTunnels}/{tunnels.length} live
          </span>
        </div>
        <div className="grid gap-3 md:grid-cols-2">
          {tunnels.map((tunnel) => (
            <article
              key={tunnel.id}
              className="group rounded-xl border border-hairline bg-card/75 p-5 transition-[border-color,background-color,transform] duration-200 hover:-translate-y-0.5 hover:border-foreground/25 hover:bg-card"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <span
                      className={`size-2 rounded-full ${tunnel.active ? "animate-pulse-dot bg-success" : "bg-muted"}`}
                    />
                    <h3 className="truncate font-semibold">{tunnel.name}</h3>
                  </div>
                  <p className="mt-1 pl-4 text-xs text-muted-foreground">
                    {tunnel.active ? "Connected & receiving traffic" : "Waiting for CLI connection"}
                  </p>
                </div>
                <span
                  className={`rounded-full border px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[.1em] ${
                    tunnel.active
                      ? "border-success/30 bg-success/10 text-success"
                      : "border-hairline text-muted-foreground"
                  }`}
                >
                  {tunnel.active ? "Live" : "Offline"}
                </span>
              </div>
              <div className="mt-5 rounded-lg border border-hairline bg-background/70 p-3">
                <div className="flex min-w-0 items-center gap-2 font-mono text-xs">
                  <ExternalLink className="size-3.5 shrink-0 text-accent" aria-hidden />
                  <span className="truncate text-accent">https://{tunnel.host}</span>
                </div>
                <div className="mt-2 flex items-center gap-2 border-t border-hairline pt-2 font-mono text-xs text-muted-foreground">
                  <span>→</span>
                  <span>localhost:{tunnel.port}</span>
                </div>
              </div>
              {tunnel.active && (
                <p className="mt-3 font-mono text-[10px] text-muted-foreground">
                  stop · bitrok stop {tunnel.name}
                </p>
              )}
            </article>
          ))}
        </div>
      </section>

      <section className="mt-14" aria-labelledby="activity-heading">
        <div className="mb-5 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <div className="signal-label">Traffic</div>
            <h2 id="activity-heading" className="mt-3 text-2xl font-semibold tracking-tight">
              Tunnel logs
            </h2>
            <p className="mt-1 text-sm text-muted-foreground">
              Recent requests through your active tunnels.
            </p>
          </div>
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <Clock3 className="size-3.5" aria-hidden />
            {logsUnavailable ? "Logs unreachable" : `Last ${logs.length || 50} requests`}
          </div>
        </div>
        <div className="overflow-x-auto rounded-xl border border-hairline bg-card/75">
          {logsUnavailable ? (
            <div className="p-8 text-center text-sm text-muted-foreground">
              Request logs could not be loaded from the relay.
            </div>
          ) : logs.length === 0 ? (
            <div className="p-8 text-center text-sm text-muted-foreground">
              Requests will appear here when a tunnel starts receiving traffic.
              <div className="mt-3 font-mono text-xs text-muted-foreground/80">
                $ bitrok myapp 3000
              </div>
            </div>
          ) : (
            <table className="w-full min-w-[720px] text-left text-sm">
              <thead>
                <tr className="border-b border-hairline text-[10px] uppercase tracking-[.13em] text-muted-foreground">
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium">Tunnel</th>
                  <th className="px-4 py-3 font-medium">Method</th>
                  <th className="px-4 py-3 font-medium">Path</th>
                  <th className="px-4 py-3 text-right font-medium">Latency</th>
                  <th className="px-4 py-3 text-right font-medium">Size</th>
                  <th className="px-4 py-3 text-right font-medium">Time</th>
                </tr>
              </thead>
              <tbody>
                {logs.map((log) => {
                  const tunnelLabel =
                    log.tunnel_name || tunnelNameById.get(log.tunnel_id) || log.tunnel_id.slice(0, 8);
                  const sizeLabel =
                    log.bytes_out > 0 || log.bytes_in > 0
                      ? formatBytes(log.bytes_out || log.bytes_in)
                      : "—";
                  return (
                    <tr
                      key={log.id}
                      className="border-b border-hairline last:border-0 hover:bg-foreground/[0.025]"
                    >
                      <td className="px-4 py-3">
                        <span
                          className={`inline-flex items-center gap-2 font-mono text-xs ${
                            log.status >= 400
                              ? "text-danger"
                              : log.status >= 300
                                ? "text-warning"
                                : "text-success"
                          }`}
                        >
                          <span className="size-1.5 rounded-full bg-current" />
                          {log.status}
                        </span>
                      </td>
                      <td className="max-w-28 truncate px-4 py-3 font-mono text-xs text-muted-foreground" title={tunnelLabel}>
                        {tunnelLabel}
                      </td>
                      <td className="px-4 py-3 font-mono text-xs font-semibold">{log.method}</td>
                      <td className="max-w-80 truncate px-4 py-3 font-mono text-xs" title={log.path}>
                        {log.path}
                      </td>
                      <td className="px-4 py-3 text-right font-mono text-xs tabular-nums text-muted-foreground">
                        {log.latency_ms}ms
                      </td>
                      <td className="px-4 py-3 text-right font-mono text-xs tabular-nums text-muted-foreground">
                        {sizeLabel}
                      </td>
                      <td className="px-4 py-3 text-right font-mono text-xs tabular-nums text-muted-foreground">
                        {new Intl.DateTimeFormat("en", {
                          hour: "2-digit",
                          minute: "2-digit",
                          second: "2-digit",
                        }).format(new Date(log.ts))}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>
      </section>
    </>
  );
}

function CommandCopy({ label, command }: { label: string; command: string }) {
  return (
    <div className="min-w-0 rounded-lg border border-hairline bg-background/80 p-2 pl-4 text-left">
      <div className="mb-1 font-mono text-[10px] uppercase tracking-[.12em] text-muted-foreground">
        {label}
      </div>
      <div className="flex min-w-0 items-center gap-2">
        <span className="font-mono text-xs text-accent">$</span>
        <code className="min-w-0 flex-1 whitespace-normal break-words font-mono text-xs sm:whitespace-nowrap">
          {command}
        </code>
        <CopyButton text={command} className="size-6 shrink-0" />
      </div>
    </div>
  );
}

function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / (1024 * 1024)).toFixed(1)} MB`;
}

function Stat({
  icon: Icon,
  label,
  value,
  detail,
  active = false,
}: {
  icon: typeof Activity;
  label: string;
  value: string;
  detail: string;
  active?: boolean;
}) {
  return (
    <div className="rounded-xl border border-hairline bg-card/75 p-5">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium text-muted-foreground">{label}</span>
        <Icon className={`size-4 ${active ? "text-success" : "text-accent"}`} aria-hidden />
      </div>
      <div className="mt-5 font-display text-4xl font-semibold tracking-[-.05em] tabular-nums">
        {value}
      </div>
      <div className="mt-1 text-xs text-muted-foreground">{detail}</div>
    </div>
  );
}
