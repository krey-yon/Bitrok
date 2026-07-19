import Link from "next/link";
import { ArrowLeft, Terminal } from "lucide-react";
import { DashboardHeader } from "@/app/components/dashboard-header";
import { buttonClassName } from "@/components/ui/button";

/**
 * Create-tunnel form removed — tunnels are CLI-only.
 * This route redirects UX to CLI docs / token page.
 */
export default function NewTunnelPage() {
  return (
    <div className="min-h-full bg-page-gradient">
      <DashboardHeader />
      <main id="main-content" className="section-shell py-10 sm:py-14">
        <Link
          href="/dashboard"
          className="inline-flex items-center gap-2 rounded-md text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-4" aria-hidden />
          Back to Overview
        </Link>

        <div className="mx-auto mt-14 max-w-xl text-center">
          <div className="mx-auto flex size-14 items-center justify-center rounded-2xl border border-hairline bg-card text-accent">
            <Terminal className="size-6" aria-hidden />
          </div>
          <div className="signal-label mt-8">CLI only</div>
          <h1 className="mt-4 text-balance text-4xl font-semibold tracking-[-.045em]">
            Create tunnels from the terminal.
          </h1>
          <p className="mt-4 text-pretty leading-7 text-muted-foreground">
            The web form is gone — deterministic URLs are carved by the CLI so the host always matches
            your app name and username.
          </p>

          <div className="mt-8 space-y-2 rounded-xl border border-hairline bg-card/80 p-5 text-left font-mono text-sm">
            <div className="text-xs text-muted-foreground"># log in once</div>
            <div>
              <span className="text-accent">$</span> bitrok login
            </div>
            <div className="pt-3 text-xs text-muted-foreground"># start a tunnel</div>
            <div>
              <span className="text-accent">$</span> bitrok myapp 3000
            </div>
            <div className="pt-1 text-xs text-muted-foreground">
              → https://myapp-&lt;you&gt;.bitrok.tech
            </div>
            <div className="pt-3 text-xs text-muted-foreground"># background</div>
            <div>
              <span className="text-accent">$</span> bitrok myapp 3000 -d
            </div>
            <div className="pt-3 text-xs text-muted-foreground"># multi-tunnel from bitrok.yml</div>
            <div>
              <span className="text-accent">$</span> bitrok up
            </div>
          </div>

          <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <Link href="/dashboard/cli-token" className={buttonClassName({ variant: "accent" })}>
              Get CLI token
            </Link>
            <Link href="/dashboard" className={buttonClassName({ variant: "ghost" })}>
              Back to dashboard
            </Link>
          </div>
        </div>
      </main>
    </div>
  );
}
