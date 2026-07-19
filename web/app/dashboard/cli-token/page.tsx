import Link from "next/link";
import { AtSign, KeyRound, ShieldCheck } from "lucide-react";
import { requireAuth } from "@/lib/session";
import { getUsernameForUser } from "@/lib/username";
import { DashboardHeader } from "@/app/components/dashboard-header";
import { buttonClassName } from "@/components/ui/button";
import { CliTokenGenerator } from "./cli-token-generator";

export default async function CliTokenPage() {
  const session = await requireAuth();
  const username = await getUsernameForUser(session.user.id);
  return (
    <div className="min-h-full bg-page-gradient">
      <DashboardHeader email={session.user.email ?? undefined} username={username} />
      <main id="main-content" className="section-shell py-10 sm:py-14">
        <div className="grid gap-8 lg:grid-cols-[.75fr_1.25fr] lg:gap-14">
          <section>
            <div className="signal-label">CLI access</div>
            <h1 className="mt-5 text-balance text-4xl font-semibold tracking-[-.045em] sm:text-5xl">
              Connect your terminal.
            </h1>
            <p className="mt-4 max-w-md text-pretty leading-7 text-muted-foreground">
              Generate a short-lived credential for the Bitrok CLI. It authorizes tunnel connections
              for your account
              {username ? (
                <>
                  {" "}
                  as <span className="font-mono text-foreground">@{username}</span>
                </>
              ) : null}
              .
            </p>
            {!username && (
              <div className="mt-6 rounded-lg border border-accent/25 bg-accent/8 p-4 text-sm">
                <div className="flex gap-3">
                  <AtSign className="mt-0.5 size-4 shrink-0 text-accent" aria-hidden />
                  <div>
                    <strong>No username yet</strong>
                    <p className="mt-1 text-muted-foreground">
                      Set one first so CLI tokens embed your slug in tunnel hosts.
                    </p>
                    <Link
                      href="/dashboard/settings"
                      className={buttonClassName({ variant: "accent", size: "sm", className: "mt-3" })}
                    >
                      Create username
                    </Link>
                  </div>
                </div>
              </div>
            )}
            <div className="mt-8 space-y-4 text-sm">
              <div className="flex gap-3">
                <KeyRound className="mt-0.5 size-4 shrink-0 text-accent" aria-hidden />
                <div>
                  <strong>30-day lifetime</strong>
                  <p className="mt-1 text-muted-foreground">Create a fresh token whenever you need one.</p>
                </div>
              </div>
              <div className="flex gap-3">
                <ShieldCheck className="mt-0.5 size-4 shrink-0 text-accent" aria-hidden />
                <div>
                  <strong>Shown once</strong>
                  <p className="mt-1 text-muted-foreground">Bitrok never displays the raw token again.</p>
                </div>
              </div>
            </div>
          </section>
          <section className="rounded-[var(--radius-xl)] border border-hairline bg-card/80 p-6 shadow-[0_24px_70px_rgb(0_0_0/8%)] sm:p-8">
            <CliTokenGenerator />
          </section>
        </div>
      </main>
    </div>
  );
}
