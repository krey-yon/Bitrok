import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { Logo } from "@/components/ui/logo";
import { ThemeToggle } from "@/components/ui/theme-toggle";

export function LegalShell({ eyebrow, title, intro, children }: { eyebrow: string; title: string; intro: string; children: React.ReactNode }) {
  return <div className="min-h-full bg-page-gradient"><header className="sticky top-0 z-50 border-b border-hairline bg-background/82 backdrop-blur-xl"><div className="section-shell flex h-16 items-center justify-between"><Link href="/" aria-label="Bitrok home" className="text-lg"><Logo /></Link><div className="flex items-center gap-1"><Link href="/" className="inline-flex items-center gap-2 rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-foreground/[0.05] hover:text-foreground"><ArrowLeft className="size-4" aria-hidden />Home</Link><ThemeToggle /></div></div></header><main id="main-content" className="section-shell py-14 sm:py-20"><div className="grid gap-12 lg:grid-cols-[.65fr_1.35fr]"><header className="lg:sticky lg:top-28 lg:self-start"><div className="signal-label">{eyebrow}</div><h1 className="mt-5 text-balance text-4xl font-semibold tracking-[-.045em] sm:text-5xl">{title}</h1><p className="mt-5 max-w-md text-pretty leading-7 text-muted-foreground">{intro}</p></header><article className="rounded-[var(--radius-xl)] border border-hairline bg-card/80 p-6 shadow-[0_20px_60px_rgb(0_0_0/7%)] sm:p-10">{children}</article></div></main></div>;
}

export function LegalSection({ number, title, children }: { number: string; title: string; children: React.ReactNode }) {
  return <section className="border-b border-hairline py-8 first:pt-0 last:border-0 last:pb-0"><div className="font-mono text-[10px] uppercase tracking-[.15em] text-accent">{number}</div><h2 className="mt-3 text-xl font-semibold tracking-tight">{title}</h2><div className="mt-3 text-[15px] leading-7 text-muted-foreground">{children}</div></section>;
}
