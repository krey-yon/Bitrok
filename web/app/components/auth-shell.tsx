import Link from "next/link";
import { Logo } from "@/components/ui/logo";
import { ThemeToggle } from "@/components/ui/theme-toggle";

export function AuthShell({ eyebrow, title, description, children, asideTitle = "One URL. Every session.", asideBody = "Your frontend keeps the same backend endpoint while your local service disconnects, restarts, and moves with you." }: { eyebrow: string; title: string; description: string; children: React.ReactNode; asideTitle?: string; asideBody?: string }) {
  return (
    <div className="relative min-h-full bg-page-gradient">
      <div className="absolute inset-0 auth-grid opacity-45 [mask-image:linear-gradient(to_bottom,#000,transparent_90%)]" aria-hidden />
      <div className="absolute inset-0 bg-noise" aria-hidden />
      <header className="relative z-10 flex h-16 items-center justify-between px-5 sm:px-8">
        <Link href="/" aria-label="Bitrok home" className="rounded-md text-lg"><Logo /></Link>
        <ThemeToggle />
      </header>
      <main id="main-content" className="relative z-10 mx-auto grid min-h-[calc(100svh-4rem)] w-full max-w-5xl items-stretch gap-6 px-5 py-5 lg:h-[calc(100svh-4rem)] lg:min-h-0 lg:grid-cols-[.88fr_1.12fr] lg:py-6">
        <section className="order-2 hidden overflow-hidden rounded-[var(--radius-xl)] border border-hairline bg-foreground p-8 text-background lg:order-1 lg:flex lg:min-h-0 lg:flex-col lg:justify-between">
          <div><div className="font-mono text-[10px] uppercase tracking-[.18em] text-background/50">Bitrok Network / Auth</div><div className="mt-4 flex items-center gap-2 text-xs text-background/70"><span className="size-1.5 animate-pulse-dot rounded-full bg-accent" />Relay available</div></div>
          <div><svg className="mb-8 w-full" viewBox="0 0 420 150" aria-hidden><path d="M18 76H150C190 76 184 28 230 28H400M150 76H260C300 76 294 122 338 122H400" fill="none" stroke="currentColor" strokeOpacity=".18" strokeWidth="1.5" strokeDasharray="5 7" /><circle cx="18" cy="76" r="6" fill="var(--secondary)" /><circle cx="150" cy="76" r="9" fill="var(--accent)" /><circle cx="400" cy="28" r="5" fill="var(--accent)" /><circle cx="400" cy="122" r="5" fill="var(--accent)" /><circle r="3" fill="var(--accent)"><animateMotion dur="3s" repeatCount="indefinite" path="M18 76H150C190 76 184 28 230 28H400" /></circle></svg><h2 className="max-w-sm text-balance text-4xl font-semibold leading-[1.02] tracking-[-.045em]">{asideTitle}</h2><p className="mt-5 max-w-sm text-pretty text-sm leading-6 text-background/60">{asideBody}</p></div>
          <div className="grid grid-cols-3 gap-px overflow-hidden rounded-lg border border-background/15 bg-background/15 text-center"><div className="bg-foreground p-3"><strong className="block font-mono text-xs text-accent">TLS</strong><span className="text-[10px] text-background/45">encrypted</span></div><div className="bg-foreground p-3"><strong className="block font-mono text-xs text-accent">HTTP</strong><span className="text-[10px] text-background/45">requests</span></div><div className="bg-foreground p-3"><strong className="block font-mono text-xs text-accent">DNS</strong><span className="text-[10px] text-background/45">stable URL</span></div></div>
        </section>
        <section className="order-1 flex items-center lg:order-2"><div className="w-full rounded-[var(--radius-xl)] border border-hairline bg-card/85 p-6 shadow-[0_24px_80px_rgb(0_0_0/10%)] backdrop-blur sm:p-10"><div className="signal-label">{eyebrow}</div><h1 className="mt-5 text-balance text-3xl font-semibold tracking-[-.04em] sm:text-4xl">{title}</h1><p className="mt-3 text-pretty text-sm leading-6 text-muted-foreground">{description}</p><div className="mt-8">{children}</div></div></section>
      </main>
    </div>
  );
}
