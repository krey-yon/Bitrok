import { requireAuth } from "@/lib/session";
import { getServerTunnels, getServerLogs } from "@/lib/server-api";
import Link from "next/link";
import { SignOutButton } from "./sign-out-button";
import { Logo } from "@/components/ui/logo";
import { Eyebrow } from "@/components/ui/eyebrow";
import { Button } from "@/components/ui/button";
import { StatusGlyph } from "@/components/ui/status-glyph";
import { TerminalPanel } from "@/components/ui/terminal";

export default async function DashboardPage() {
  const session = await requireAuth();

  const [tunnelsResult, logsResult] = await Promise.allSettled([
    getServerTunnels(session.user.id),
    getServerLogs(session.user.id, 10),
  ]);

  const tunnels =
    tunnelsResult.status === "fulfilled" ? tunnelsResult.value : [];
  const logs =
    logsResult.status === "fulfilled" ? logsResult.value.logs : [];
  const totalRequests =
    logsResult.status === "fulfilled" ? logsResult.value.total : 0;

  if (tunnelsResult.status === "rejected") {
    console.error("dashboard: failed to load tunnels:", tunnelsResult.reason);
  }
  if (logsResult.status === "rejected") {
    console.error("dashboard: failed to load logs:", logsResult.reason);
  }

  const totalTunnels = tunnels.length;
  const activeTunnels = tunnels.filter((t) => t.active).length;

  const statusVariant = (status: number) =>
    status >= 400 ? "danger" : status >= 300 ? "warning" : "success";
  const methodColor = (status: number) =>
    status >= 400
      ? "text-danger"
      : status >= 300
        ? "text-accent"
        : "text-success";

  return (
    <div className="min-h-full flex flex-col">
      <nav className="sticky top-0 z-50 border-b border-hairline bg-background/80 backdrop-blur">
        <div className="max-w-3xl mx-auto px-6 h-12 flex items-center justify-between text-sm">
          <Link href="/dashboard" className="font-mono">
            <Logo />
          </Link>
          <div className="flex items-center gap-5 text-muted-foreground font-mono text-xs">
            <Link
              href="/dashboard/cli-token"
              className="hover:text-foreground transition-colors"
            >
              cli token
            </Link>
            <span className="hidden sm:inline truncate max-w-[180px]">
              {session.user.email}
            </span>
            <SignOutButton />
          </div>
        </div>
      </nav>

      <main className="flex-1 max-w-3xl mx-auto px-6 py-14 w-full">
        <Eyebrow ornament="·">tunnels</Eyebrow>
        <h1 className="mt-3 text-3xl font-semibold tracking-tight mb-2">
          Tunnels.
        </h1>
        <p className="text-sm text-muted font-mono mb-12">
          {activeTunnels} of {totalTunnels} active · {totalRequests} requests
          served
        </p>

        {/* Tunnel list — hairline rows with status glyphs */}
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-xs uppercase tracking-[0.22em] text-muted-foreground font-mono">
            your tunnels
          </h2>
          <Link
            href="/dashboard/tunnels/new"
            className="text-sm text-accent font-mono hover:underline"
          >
            › new tunnel
          </Link>
        </div>

        {tunnels.length === 0 ? (
          <TerminalPanel title="no tunnels yet" className="text-left">
            <div className="space-y-4">
              <p className="text-sm text-muted font-mono">
                {"// no tunnels yet. create one, then run bitrok up"}
              </p>
              <Link href="/dashboard/tunnels/new">
                <Button arrow>Create tunnel</Button>
              </Link>
            </div>
          </TerminalPanel>
        ) : (
          <div className="border-t border-hairline">
            {tunnels.map((tunnel) => (
              <div
                key={tunnel.id}
                className="flex items-center justify-between gap-4 py-4 border-b border-hairline"
              >
                <div className="flex items-center gap-3 min-w-0">
                  <StatusGlyph
                    variant={tunnel.active ? "active" : "idle"}
                    pulse={tunnel.active}
                  />
                  <div className="min-w-0">
                    <div className="font-medium truncate">{tunnel.name}</div>
                    <div className="text-xs text-muted font-mono truncate">
                      {tunnel.host} → :{tunnel.port}
                    </div>
                  </div>
                </div>
                <span
                  className={`text-xs font-mono shrink-0 ${tunnel.active ? "text-success" : "text-muted-foreground"}`}
                >
                  {tunnel.active ? "active" : "inactive"}
                </span>
              </div>
            ))}
          </div>
        )}

        {/* Recent activity — live relay log */}
        {logs.length > 0 && (
          <>
            <h2 className="text-xs uppercase tracking-[0.22em] text-muted-foreground font-mono mt-14 mb-3">
              recent activity
            </h2>
            <TerminalPanel
              title="relay log · last 10"
              className="text-left"
              bodyClassName="px-3.5 py-2.5"
            >
              <div className="space-y-0.5">
                {logs.map((log) => (
                  <div
                    key={log.id}
                    className="flex items-center gap-3 py-1"
                  >
                    <span className="w-4 shrink-0 text-center">
                      <StatusGlyph variant={statusVariant(log.status)} />
                    </span>
                    <span
                      className={`font-mono text-xs shrink-0 w-12 ${methodColor(log.status)}`}
                    >
                      {log.method}
                    </span>
                    <span className="font-mono text-xs text-foreground truncate flex-1 min-w-0">
                      {log.path}
                    </span>
                    <span className="font-mono text-xs text-muted shrink-0 tabular-nums">
                      {log.status}
                    </span>
                    <span className="font-mono text-xs text-muted shrink-0 tabular-nums hidden sm:inline">
                      {log.latency_ms}ms
                    </span>
                    <span className="font-mono text-xs text-muted shrink-0 hidden md:inline">
                      {new Date(log.ts).toLocaleTimeString()}
                    </span>
                  </div>
                ))}
              </div>
            </TerminalPanel>
          </>
        )}
      </main>
    </div>
  );
}
