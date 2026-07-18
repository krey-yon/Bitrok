import Link from "next/link";
import { Globe, Server, Star, Terminal } from "lucide-react";
import { AuthRedirect } from "@/app/auth-redirect";
import { AuroraBackground } from "@/app/components/aurora-background";
import { NetworkMesh } from "@/app/components/network-mesh";
import { Logo } from "@/components/ui/logo";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Eyebrow } from "@/components/ui/eyebrow";
import { StatusGlyph } from "@/components/ui/status-glyph";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { CopyButton } from "@/components/ui/copy-button";

const INSTALL_CMD = "curl -fsSL bitrok.tech/install | sh";

export default function Home() {
  return (
    <div className="min-h-full flex flex-col bg-page-gradient">
      <AuthRedirect />

      {/* Nav — hairline, blurred, minimal */}
      <nav className="sticky top-0 z-50 border-b border-hairline bg-background/70 backdrop-blur-xl">
        <div className="max-w-5xl mx-auto px-6 h-14 flex items-center justify-between text-sm">
          <Link href="/" className="text-base">
            <Logo />
          </Link>
          <div className="flex items-center gap-2 text-muted-foreground">
            <Link href="https://github.com/krey-yon/Bitrok">
              <Button variant="ghost" size="sm">
                <Star className="size-3.5" aria-hidden />
                Open Source
              </Button>
            </Link>
            <Link
              href="/login"
              className="px-3 py-1.5 rounded-[calc(var(--radius)*0.85)] hover:text-foreground hover:bg-foreground/[0.05] transition-colors"
            >
              Sign in
            </Link>
            <Link href="/register">
              <Button variant="accent" size="sm" arrow>
                Get started
              </Button>
            </Link>
            <ThemeToggle className="ml-1" />
          </div>
        </div>
      </nav>

      <main className="flex-1">
        {/* Hero — gradient + noise + dotted art + animated SVG aurora */}
        <section className="relative overflow-hidden">
          <div className="absolute inset-0 bg-hero-gradient" aria-hidden />
          <AuroraBackground className="absolute inset-0 w-full h-full" />
          <div className="absolute inset-0 bg-noise" aria-hidden />
          <div className="absolute inset-0 bg-dots" aria-hidden />

          <div className="relative max-w-5xl mx-auto px-6 pt-32 pb-20">
            <div className="flex flex-col items-center text-center">
              <Eyebrow className="mb-6 animate-fade-in">
                self-hosted tunnels
              </Eyebrow>
              <h1 className="text-5xl sm:text-6xl lg:text-7xl font-semibold tracking-[-0.03em] leading-[1.04] animate-slide-up">
                localhost,
                <br />
                meet the <span className="text-accent">internet.</span>
              </h1>
              <p className="mt-6 text-lg text-muted-foreground max-w-xl mx-auto leading-relaxed animate-slide-up [animation-delay:80ms]">
                One command gives your local port a permanent public URL.
                Your server. Your domain. Your rules.
              </p>

              <div className="mt-9 flex flex-wrap items-center justify-center gap-4 animate-slide-up [animation-delay:160ms]">
                <Link href="/register">
                  <Button variant="accent" size="lg" arrow>
                    Get started
                  </Button>
                </Link>
                <div className="flex items-center gap-2 rounded-[var(--radius)] border border-hairline bg-card/80 pl-3 pr-1.5 h-12 font-mono text-sm backdrop-blur-sm">
                  <span className="text-muted select-none">$</span>
                  <code className="text-foreground">{INSTALL_CMD}</code>
                  <CopyButton text={INSTALL_CMD} />
                </div>
              </div>

              <p className="mt-5 text-xs text-muted-foreground animate-slide-up [animation-delay:200ms]">
                free · open source · self-hosted
              </p>
            </div>
          </div>
        </section>

        {/* Features */}
        <section className="relative">
          <div className="max-w-5xl mx-auto px-6 py-24">
            <div className="text-center mb-14">
              <Eyebrow className="mb-4">what you get</Eyebrow>
              <h2 className="text-3xl sm:text-4xl font-semibold tracking-[-0.02em]">
                Three commands. One permanent URL.
              </h2>
            </div>
            <div className="grid sm:grid-cols-3 gap-px bg-hairline rounded-[var(--radius-lg)] overflow-hidden border border-hairline">
              {FEATURES.map((f) => (
                <div
                  key={f.title}
                  className="bg-card p-7 flex flex-col gap-4 transition-colors duration-200 hover:bg-background/60"
                >
                  <div className="inline-flex size-10 items-center justify-center rounded-[var(--radius)] border border-hairline text-accent">
                    <f.icon className="size-5" aria-hidden />
                  </div>
                  <div>
                    <h3 className="font-semibold tracking-tight mb-1.5">
                      {f.title}
                    </h3>
                    <p className="text-sm text-muted-foreground leading-relaxed">
                      {f.body}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Stats band — animated relay mesh behind the numbers */}
        <section className="relative overflow-hidden border-y border-hairline">
          <NetworkMesh className="absolute inset-0 w-full h-full opacity-40" />
          <div className="absolute inset-0 bg-noise" aria-hidden />
          <div className="relative max-w-5xl mx-auto px-6 py-20">
            <div className="text-center mb-12">
              <Eyebrow className="mb-4">by the numbers</Eyebrow>
              <h2 className="text-3xl sm:text-4xl font-semibold tracking-[-0.02em]">
                One relay. Every tunnel.
              </h2>
            </div>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-8">
              {STATS.map((s) => (
                <div key={s.label} className="text-center">
                  <div className="text-4xl sm:text-5xl font-semibold tracking-tight text-accent">
                    {s.value}
                  </div>
                  <div className="mt-2 text-xs uppercase tracking-[0.18em] text-muted-foreground">
                    {s.label}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Getting started — three terminal-styled steps */}
        <section className="relative">
          <div className="max-w-5xl mx-auto px-6 py-24">
            <div className="text-center mb-14">
              <Eyebrow className="mb-4">thirty seconds</Eyebrow>
              <h2 className="text-3xl sm:text-4xl font-semibold tracking-[-0.02em]">
                Start to finish.
              </h2>
            </div>
            <div className="grid sm:grid-cols-3 gap-5">
              {STEPS.map((s) => (
                <Card key={s.n} variant="terminal" terminalTitle={`step ${s.n}`}>
                  <div className="px-4 py-4 space-y-3">
                    <div className="flex items-start gap-2.5 font-mono text-sm">
                      <StatusGlyph variant="success" className="mt-0.5 shrink-0" />
                      <span className="text-foreground break-all">{s.cmd}</span>
                    </div>
                    <p className="text-xs text-muted-foreground leading-relaxed pl-6">
                      {s.desc}
                    </p>
                  </div>
                </Card>
              ))}
            </div>
          </div>
        </section>

        {/* Final CTA */}
        <section className="relative overflow-hidden">
          <div className="absolute inset-0 bg-hero-gradient" aria-hidden />
          <AuroraBackground className="absolute inset-0 w-full h-full" />
          <div className="absolute inset-0 bg-noise" aria-hidden />
          <div className="absolute inset-0 bg-dots" aria-hidden />
          <div className="relative max-w-5xl mx-auto px-6 py-28 text-center">
            <Eyebrow className="mb-4">ditch the random urls</Eyebrow>
            <h2 className="text-4xl sm:text-5xl font-semibold tracking-[-0.02em] mb-4">
              Your server. Your domain. Your rules.
            </h2>
            <p className="text-muted-foreground mb-8">
              Free. Open source. Yours.
            </p>
            <Link href="/register">
              <Button variant="accent" size="lg" arrow>
                Get started
              </Button>
            </Link>
          </div>
        </section>
      </main>

      <footer className="border-t border-hairline">
        <div className="max-w-5xl mx-auto px-6 py-8 flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-muted-foreground font-mono">
          <span>© 2026 bitrok</span>
          <div className="flex items-center gap-6">
            <Link
              href="https://github.com"
              className="hover:text-foreground transition-colors"
            >
              github
            </Link>
            <Link href="/privacy" className="hover:text-foreground transition-colors">
              privacy
            </Link>
            <Link href="/security" className="hover:text-foreground transition-colors">
              security
            </Link>
          </div>
        </div>
      </footer>
    </div>
  );
}

const FEATURES = [
  {
    title: "Permanent subdomains",
    body: "Claim myapp.bitrok.tech once and keep it forever. No random URLs on every restart.",
    icon: Globe,
  },
  {
    title: "Self-hosted",
    body: "The relay is a single Go binary with SQLite. Run it on your own hardware — your traffic never touches us.",
    icon: Server,
  },
  {
    title: "Minimal CLI",
    body: "bitrok login, bitrok create, bitrok up. Three commands and you're receiving traffic.",
    icon: Terminal,
  },
] as const;

const STATS = [
  { value: "1", label: "binary to deploy" },
  { value: "0", label: "keys exposed" },
  { value: "∞", label: "permanent subdomains" },
  { value: "<30s", label: "to first tunnel" },
] as const;

const STEPS = [
  { n: "1", cmd: "bitrok login", desc: "Authenticate via your browser." },
  { n: "2", cmd: "bitrok create myapp --port 3000", desc: "Reserve your subdomain." },
  { n: "3", cmd: "bitrok up myapp", desc: "Traffic flows to localhost." },
] as const;
