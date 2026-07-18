import Link from "next/link";
import { KeyRound, LayoutDashboard, Plus } from "lucide-react";
import { Logo } from "@/components/ui/logo";
import { ThemeToggle } from "@/components/ui/theme-toggle";

export function DashboardHeader({ email, signOut }: { email?: string; signOut?: React.ReactNode }) {
  return (
    <header className="sticky top-0 z-50 border-b border-hairline bg-background/82 backdrop-blur-xl">
      <div className="section-shell flex h-16 items-center justify-between gap-4">
        <Link href="/dashboard" aria-label="Bitrok dashboard" className="shrink-0 rounded-md text-lg"><Logo /></Link>
        <nav aria-label="Dashboard navigation" className="flex min-w-0 items-center gap-1">
          <Link href="/dashboard" className="hidden items-center gap-2 rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-foreground/[0.05] hover:text-foreground md:flex"><LayoutDashboard className="size-3.5" aria-hidden />Overview</Link>
          <Link href="/dashboard/cli-token" className="hidden items-center gap-2 rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-foreground/[0.05] hover:text-foreground sm:flex"><KeyRound className="size-3.5" aria-hidden />CLI Token</Link>
          <Link href="/dashboard/tunnels/new" className="inline-flex h-9 items-center gap-1.5 rounded-md bg-accent px-3 text-xs font-semibold text-accent-foreground transition-colors hover:bg-accent-light"><Plus className="size-3.5" aria-hidden />New Tunnel</Link>
          <ThemeToggle />
          {email && <span className="hidden max-w-40 truncate border-l border-hairline pl-3 text-xs text-muted-foreground lg:inline" title={email}>{email}</span>}
          {signOut}
        </nav>
      </div>
    </header>
  );
}
