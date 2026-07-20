import Link from "next/link";
import { ArrowRight, Check, Cloud, Code2, ExternalLink, Radio, Server, Sparkles } from "lucide-react";
import { AuthRedirect } from "@/app/auth-redirect";
import { TrafficDiagram } from "@/app/components/traffic-diagram";
import { Logo } from "@/components/ui/logo";
import { buttonClassName } from "@/components/ui/button";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { CopyButton } from "@/components/ui/copy-button";
import { GitHubMark } from "@/components/ui/github-mark";

const INSTALL_CMD = "curl -fsSL bitrok.tech/install | sh";
const INSTALL_WINDOWS_CMD = "irm https://bitrok.tech/install.ps1 | iex";
const GITHUB_URL = "https://github.com/krey-yon/Bitrok";
const X_URL = "https://x.com/Krey_yon";
const LINKEDIN_URL = "https://www.linkedin.com/in/vikas387/";

export default function Home() {
  return (
    <div className="min-h-full bg-page-gradient">
      <AuthRedirect />
      <header className="sticky top-0 z-50 border-b border-hairline bg-background/80 backdrop-blur-xl">
        <div className="section-shell flex h-16 min-w-0 items-center justify-between gap-3">
          <Link href="/" aria-label="Bitrok home" className="shrink-0 rounded-md text-lg"><Logo /></Link>
          <nav aria-label="Main navigation" className="flex min-w-0 items-center gap-1 sm:gap-2">
            <Link href="#how-it-works" className="hidden rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:text-foreground md:inline-flex">How It Works</Link>
            <Link href="#use-cases" className="hidden rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:text-foreground md:inline-flex">Use Cases</Link>
            <Link href={GITHUB_URL} aria-label="Bitrok creator krey-yon on GitHub" className="hidden items-center gap-2 rounded-md px-2.5 py-2 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.05] hover:text-foreground sm:inline-flex"><GitHubMark /><span className="font-mono">krey-yon</span></Link>
            <ThemeToggle />
            <Link href="/register" className={buttonClassName({ variant: "accent", size: "sm", className: "shrink-0 px-3 sm:px-3.5" })}>Get Started <ArrowRight className="size-3.5" aria-hidden /></Link>
          </nav>
        </div>
      </header>

      <main id="main-content">
        <section className="relative isolate overflow-hidden border-b border-hairline">
          <div className="absolute inset-0 bg-hero-gradient" aria-hidden />
          <div className="absolute inset-0 bg-grid [mask-image:linear-gradient(to_bottom,#000,transparent_88%)]" aria-hidden />
          <div className="absolute inset-0 bg-noise" aria-hidden />
          <div className="hero-stage section-shell relative grid min-h-[calc(100svh-4rem)] min-w-0 items-center gap-10 py-12 sm:py-14 lg:grid-cols-[1.08fr_.92fr] lg:py-16">
            <div className="min-w-0 max-w-2xl">
              <div className="signal-label mb-7 animate-fade-in">Deterministic tunnels for local services</div>
              <h1 className="hero-title max-w-full text-balance text-[clamp(2.7rem,12vw,5.7rem)] font-semibold leading-[.93] tracking-[-.058em] animate-slide-up sm:text-[clamp(3.15rem,6vw,5.7rem)] sm:leading-[.91] sm:tracking-[-.06em]">
                Your backend can live <span className="relative inline-block text-accent"><span className="relative z-10">anywhere.</span><svg className="absolute -bottom-2 left-0 h-3 w-full" viewBox="0 0 400 14" preserveAspectRatio="none" aria-hidden><path d="M3 10C92 2 267 2 397 8" fill="none" stroke="var(--secondary)" strokeWidth="4" strokeLinecap="round" /></svg></span>
              </h1>
              <p className="hero-copy mt-6 max-w-xl text-pretty text-base leading-7 text-muted-foreground animate-slide-up [animation-delay:80ms] sm:text-lg sm:leading-8">
                Give a local API, webhook receiver, or demo backend one permanent public URL. Set it in Vercel once, then run your service from your laptop whenever you need it.
              </p>
              <div className="hero-actions mt-7 max-w-2xl animate-slide-up [animation-delay:140ms]">
                <InstallCommand label="macOS / Linux" command={INSTALL_CMD} prompt="$" featured />
              </div>
              <div className="hero-benefits mt-6 flex flex-wrap gap-x-6 gap-y-2 text-sm text-muted-foreground">
                {['Stable across restarts', 'HTTP requests', 'Self-hostable'].map((item) => <span key={item} className="inline-flex items-center gap-2"><Check className="size-3.5 text-accent" aria-hidden />{item}</span>)}
              </div>
            </div>

            <div className="hero-visual relative mx-auto hidden min-w-0 w-full max-w-2xl overflow-hidden lg:ml-auto lg:block lg:overflow-visible animate-slide-up [animation-delay:180ms]">
              <div className="absolute -inset-8 rounded-full bg-accent/10 blur-3xl" aria-hidden />
              <div className="relative overflow-hidden rounded-[1.35rem] border border-border bg-card shadow-[0_28px_90px_rgb(0_0_0/18%)]">
                <div className="flex h-11 items-center justify-between border-b border-hairline px-4">
                  <div className="flex items-center gap-1.5" aria-hidden><span className="size-2 rounded-full bg-danger" /><span className="size-2 rounded-full bg-warning" /><span className="size-2 rounded-full bg-success" /></div>
                  <div className="font-mono text-[10px] uppercase tracking-[.17em] text-muted-foreground">live route · tls</div>
                  <div className="flex items-center gap-1.5 text-[10px] font-medium text-success"><span className="size-1.5 animate-pulse-dot rounded-full bg-success" />ONLINE</div>
                </div>
                <div className="relative p-4 sm:p-7">
                  <div className="absolute inset-0 bg-grid-fine opacity-35" aria-hidden />
                  <div className="relative min-w-0 overflow-hidden rounded-xl border border-hairline bg-background/60 p-2">
                    <TrafficDiagram className="block h-auto max-w-full" />
                  </div>
                  <div className="relative mt-4 grid gap-3 sm:grid-cols-[1fr_auto]">
                    <div className="min-w-0 rounded-lg border border-hairline bg-card px-4 py-3">
                      <div className="mb-1 font-mono text-[10px] uppercase tracking-[.15em] text-muted-foreground">Permanent endpoint</div>
                      <div className="truncate font-mono text-xs text-accent sm:text-sm">https://excali-kreyon.bitrok.tech</div>
                    </div>
                    <div className="flex items-center gap-3 rounded-lg border border-hairline bg-card px-4 py-3 font-mono text-xs"><span className="text-muted-foreground">p95</span><strong className="tabular-nums">42ms</strong></div>
                  </div>
                </div>
              </div>
              <div className="absolute -bottom-5 -left-4 hidden rounded-lg border border-hairline bg-card px-4 py-3 shadow-xl sm:block">
                <div className="font-mono text-[10px] uppercase tracking-[.14em] text-muted-foreground">Frontend env</div>
                <code className="mt-1 block text-xs">BACKEND_URL=<span className="text-accent">your stable URL</span></code>
              </div>
            </div>
          </div>
        </section>

        <section className="border-b border-hairline bg-card/35">
          <div className="section-shell grid grid-cols-2 divide-x divide-y divide-hairline border-x border-hairline md:grid-cols-4 md:divide-y-0">
            {TRUST_ITEMS.map(({ label, value }) => <div key={label} className="px-5 py-6 sm:px-8"><div className="font-display text-xl font-semibold tracking-tight sm:text-2xl">{value}</div><div className="mt-1 text-xs text-muted-foreground">{label}</div></div>)}
          </div>
        </section>

        <section id="how-it-works" className="section-shell py-24 sm:py-32">
          <div className="grid gap-12 lg:grid-cols-[.7fr_1.3fr]">
            <div className="lg:sticky lg:top-28 lg:self-start">
              <div className="signal-label">How It Works</div>
              <h2 className="mt-5 max-w-md text-balance text-4xl font-semibold leading-[1.03] tracking-[-.045em] sm:text-5xl">A fixed address for a moving backend.</h2>
              <p className="mt-5 max-w-md text-pretty leading-7 text-muted-foreground">The URL belongs to your project—not to a temporary session. Disconnect, restart, switch networks, and reconnect without touching frontend configuration.</p>
            </div>
            <div className="space-y-4">
              {STEPS.map((step, index) => <article key={step.title} className="group grid gap-6 rounded-[var(--radius-lg)] border border-hairline bg-card/70 p-6 transition-[border-color,transform,background-color] duration-300 hover:-translate-y-0.5 hover:border-foreground/25 hover:bg-card sm:grid-cols-[4rem_1fr_auto] sm:items-center sm:p-8">
                <div className="font-mono text-xs text-muted-foreground">0{index + 1}</div>
                <div><h3 className="text-xl font-semibold tracking-tight">{step.title}</h3><p className="mt-2 max-w-xl text-sm leading-6 text-muted-foreground">{step.body}</p></div>
                <code className="w-fit rounded-md border border-hairline bg-background px-3 py-2 font-mono text-xs text-accent">{step.code}</code>
              </article>)}
            </div>
          </div>
        </section>

        <section id="use-cases" className="border-y border-hairline bg-foreground text-background">
          <div className="section-shell py-24 sm:py-32">
            <div className="grid gap-8 lg:grid-cols-2 lg:items-end">
              <div><div className="signal-label !text-background/60">Built for demos that need to stay configured</div><h2 className="mt-5 max-w-2xl text-balance text-4xl font-semibold leading-[1.03] tracking-[-.045em] sm:text-5xl">Ship the frontend. Keep the backend close.</h2></div>
              <p className="max-w-xl text-pretty leading-7 text-background/65 lg:justify-self-end">A serious public endpoint without paying to keep every experimental service running on cloud infrastructure 24/7.</p>
            </div>
            <div className="mt-14 grid gap-px overflow-hidden rounded-[var(--radius-lg)] border border-background/15 bg-background/15 md:grid-cols-2 lg:grid-cols-4">
              {USE_CASES.map(({ icon: Icon, title, body }) => <article key={title} className="bg-foreground p-7 transition-colors hover:bg-background/[0.06]"><Icon className="size-6 text-accent" aria-hidden /><h3 className="mt-12 text-lg font-semibold">{title}</h3><p className="mt-3 text-sm leading-6 text-background/60">{body}</p></article>)}
            </div>
          </div>
        </section>

        <section className="section-shell py-24 sm:py-32">
          <div className="grid overflow-hidden rounded-[var(--radius-xl)] border border-hairline bg-card lg:grid-cols-2">
            <div className="p-7 sm:p-12"><div className="signal-label">Minimal by design</div><h2 className="mt-5 text-balance text-3xl font-semibold tracking-[-.04em] sm:text-4xl">From zero to public in 3 commands.</h2><p className="mt-4 max-w-lg leading-7 text-muted-foreground">No DNS edits per project. No random address to paste into Vercel. No cloud VM for every backend.</p><div className="mt-8 flex flex-wrap gap-3">{['macOS', 'Linux', 'Windows'].map((os) => <span key={os} className="rounded-full border border-hairline px-3 py-1.5 text-xs text-muted-foreground">{os}</span>)}</div></div>
            <div className="border-t border-hairline bg-background p-4 sm:p-8 lg:border-l lg:border-t-0">
              <div className="overflow-hidden rounded-xl border border-border bg-[#0c0f0a] text-[#eff5e5] shadow-2xl">
                <div className="flex h-10 items-center justify-between border-b border-white/10 px-4"><span className="font-mono text-[10px] text-white/45">~/projects/excali</span><span className="size-1.5 rounded-full bg-[#b8f34a] shadow-[0_0_10px_#b8f34a]" /></div>
                <div className="space-y-4 p-5 font-mono text-xs sm:text-sm"><p><span className="text-[#b8f34a]">$</span> bitrok login</p><p><span className="text-[#b8f34a]">$</span> bitrok create excali --port 3000</p><p><span className="text-[#b8f34a]">$</span> bitrok up excali</p><div className="border-t border-white/10 pt-4 text-white/55"><p><span className="text-[#b8f34a]">●</span> tunnel connected</p><p className="mt-1 break-all text-[#b8f34a]">https://excali-kreyon.bitrok.tech → localhost:3000</p></div></div>
              </div>
            </div>
          </div>
        </section>

        <section className="relative overflow-hidden border-t border-hairline">
          <div className="absolute inset-0 bg-grid opacity-60" aria-hidden /><div className="absolute inset-0 bg-noise" aria-hidden />
          <div className="section-shell relative py-24 text-center sm:py-32"><Sparkles className="mx-auto size-7 text-secondary" aria-hidden /><h2 className="mx-auto mt-6 max-w-3xl text-balance text-4xl font-semibold leading-[1] tracking-[-.05em] sm:text-6xl">Make your local backend feel permanent.</h2><p className="mx-auto mt-6 max-w-xl text-pretty text-lg text-muted-foreground">Claim the endpoint your frontend can depend on.</p><div className="mt-9 flex justify-center"><Link href="/register" className={buttonClassName({ variant: "accent", size: "lg" })}>Get Started <ArrowRight className="size-4" aria-hidden /></Link></div><div className="mx-auto mt-8 grid max-w-2xl gap-2 sm:grid-cols-2"><InstallCommand label="macOS / Linux" command={INSTALL_CMD} prompt="$" /><InstallCommand label="Windows PowerShell" command={INSTALL_WINDOWS_CMD} prompt=">" /></div></div>
        </section>
      </main>

      <footer className="border-t border-hairline bg-card/40"><div className="section-shell flex flex-col gap-8 py-10 sm:flex-row sm:items-end sm:justify-between"><div><Logo className="text-lg" /><p className="mt-3 max-w-sm text-sm text-muted-foreground">Deterministic tunnels for backends that live wherever you do.</p></div><div className="flex flex-wrap items-center gap-x-6 gap-y-3 text-sm text-muted-foreground"><Link href={GITHUB_URL} className="inline-flex items-center gap-1.5 hover:text-foreground"><GitHubMark className="size-3.5" />krey-yon</Link><a href={X_URL} target="_blank" rel="noopener noreferrer" className="hover:text-foreground">X / Twitter</a><a href={LINKEDIN_URL} target="_blank" rel="noopener noreferrer" className="inline-flex items-center gap-1.5 hover:text-foreground">LinkedIn<ExternalLink className="size-3" aria-hidden /></a><Link href="/privacy" className="hover:text-foreground">Privacy</Link><Link href="/security" className="hover:text-foreground">Security</Link><span>© 2026 Bitrok</span></div></div></footer>
    </div>
  );
}

function InstallCommand({ label, command, prompt, featured = false }: { label: string; command: string; prompt: string; featured?: boolean }) {
  return <div className={featured ? "min-w-0 rounded-lg border border-border bg-card p-2 pl-4 text-left shadow-sm" : "min-w-0 rounded-lg border border-hairline bg-card/80 p-1.5 pl-4 text-left"}><div className="mb-1 font-mono text-[10px] uppercase tracking-[.12em] text-muted-foreground">{label}</div><div className="flex min-w-0 items-center gap-2"><span className="font-mono text-xs text-muted-foreground">{prompt}</span><code className={featured ? "min-w-0 flex-1 whitespace-normal break-words font-mono text-xs sm:whitespace-nowrap" : "min-w-0 flex-1 truncate font-mono text-xs"}>{command}</code><CopyButton text={command} className={featured ? "size-6 shrink-0 rounded-md text-muted-foreground/80 hover:bg-accent/[0.06]" : undefined} /></div></div>;
}

const TRUST_ITEMS = [
  { value: "1 URL", label: "across every restart" },
  { value: "30 sec", label: "from install to public" },
  { value: "HTTP", label: "request traffic" },
  { value: "Your relay", label: "self-host when you want" },
] as const;

const STEPS = [
  { title: "Reserve a name", body: "Choose a project slug once. Bitrok combines it with your username to create a URL nobody else can claim.", code: "excali-kreyon" },
  { title: "Point it at localhost", body: "Map the endpoint to any local port. Your service stays private until the CLI opens an authenticated tunnel.", code: "localhost:3000" },
  { title: "Configure the frontend once", body: "Use the permanent endpoint in Vercel, Netlify, a mobile app, or a game client. Reconnect without changing it.", code: "bitrok up excali" },
] as const;

const USE_CASES = [
  { icon: Code2, title: "APIs & webhooks", body: "Test OAuth callbacks, payments, and third-party webhooks against local code." },
  { icon: Radio, title: "Webhook testing", body: "Expose OAuth callbacks, payment hooks, and event receivers against local code." },
  { icon: Cloud, title: "Vercel frontends", body: "Keep one backend environment variable while development moves with you." },
  { icon: Server, title: "Internal tools", body: "Share dashboards, Kafka UIs, and admin services without deploying each one." },
] as const;
