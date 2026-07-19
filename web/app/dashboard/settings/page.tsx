import { AtSign, KeyRound, Terminal } from "lucide-react";
import Link from "next/link";
import { requireAuth } from "@/lib/session";
import { getUsernameForUser } from "@/lib/username";
import { DashboardHeader } from "@/app/components/dashboard-header";
import { SignOutButton } from "../sign-out-button";
import { buttonClassName } from "@/components/ui/button";
import { UsernameForm } from "./username-form";

export default async function SettingsPage() {
  const session = await requireAuth();
  const username = await getUsernameForUser(session.user.id);

  return (
    <div className="min-h-full bg-page-gradient">
      <DashboardHeader
        email={session.user.email ?? undefined}
        username={username}
        signOut={<SignOutButton />}
      />
      <main id="main-content" className="section-shell py-10 sm:py-14">
        <div className="grid gap-8 lg:grid-cols-[.75fr_1.25fr] lg:gap-14">
          <section>
            <div className="signal-label">Account</div>
            <h1 className="mt-5 text-balance text-4xl font-semibold tracking-[-.045em] sm:text-5xl">
              Your username.
            </h1>
            <p className="mt-4 max-w-md text-pretty leading-7 text-muted-foreground">
              This slug is baked into every public tunnel URL and into CLI tokens. Pick something
              short you want on the internet.
            </p>
            <div className="mt-8 space-y-4 text-sm">
              <div className="flex gap-3">
                <AtSign className="mt-0.5 size-4 shrink-0 text-accent" aria-hidden />
                <div>
                  <strong>Deterministic hosts</strong>
                  <p className="mt-1 text-muted-foreground">
                    <code className="font-mono text-foreground">bitrok myapp 3000</code> becomes{" "}
                    <code className="font-mono text-foreground">
                      myapp-{username || "you"}.bitrok.tech
                    </code>
                    .
                  </p>
                </div>
              </div>
              <div className="flex gap-3">
                <KeyRound className="mt-0.5 size-4 shrink-0 text-accent" aria-hidden />
                <div>
                  <strong>CLI tokens carry it</strong>
                  <p className="mt-1 text-muted-foreground">
                    After changing username, generate a new CLI token so new tunnels pick up the
                    slug.
                  </p>
                </div>
              </div>
            </div>
            <Link
              href="/dashboard/cli-token"
              className={buttonClassName({ variant: "ghost", className: "mt-8" })}
            >
              <Terminal className="size-4" aria-hidden />
              CLI token
            </Link>
          </section>

          <section className="rounded-[var(--radius-xl)] border border-hairline bg-card/80 p-6 shadow-[0_24px_70px_rgb(0_0_0/8%)] sm:p-8">
            <h2 className="text-xl font-semibold">
              {username ? "Update username" : "Create username"}
            </h2>
            <p className="mt-1 text-sm text-muted-foreground">
              {username
                ? "Changing this does not rename existing tunnel hosts — only new tunnels."
                : "Required for clean tunnel URLs from the CLI."}
            </p>
            <div className="mt-7">
              <UsernameForm initialUsername={username} />
            </div>
          </section>
        </div>
      </main>
    </div>
  );
}
