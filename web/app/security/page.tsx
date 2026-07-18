import type { Metadata } from "next";
import Link from "next/link";
import { Logo } from "@/components/ui/logo";
import { Eyebrow } from "@/components/ui/eyebrow";
import { Divider } from "@/components/ui/divider";
import { StatusGlyph } from "@/components/ui/status-glyph";

export const metadata: Metadata = {
  title: "Security",
  description: "Bitrok security practices and vulnerability disclosure policy.",
};

const sectionLabel =
  "font-mono text-xs uppercase tracking-[0.22em] text-muted-foreground mt-10 mb-3";
const body = "text-sm text-muted-foreground leading-relaxed mb-4";

const measures = [
  "All connections use HTTPS with TLS 1.3",
  "Passwords hashed with bcrypt",
  "GitHub OAuth sign-in",
  "Regular dependency updates",
  "Content Security Policy headers",
  "Rate limiting on API endpoints",
];

export default function SecurityPage() {
  return (
    <div className="min-h-full flex flex-col">
      <nav className="sticky top-0 z-50 border-b border-hairline bg-background/80 backdrop-blur">
        <div className="max-w-3xl mx-auto px-6 h-12 flex items-center justify-between text-sm">
          <Link href="/" className="font-mono">
            <Logo />
          </Link>
          <Link
            href="/"
            className="text-muted-foreground hover:text-foreground transition-colors font-mono text-xs"
          >
            ← home
          </Link>
        </div>
      </nav>

      <main className="flex-1 max-w-3xl mx-auto px-6 py-14 w-full">
        <Eyebrow ornament="✦">security</Eyebrow>
        <h1 className="mt-3 text-3xl font-semibold tracking-tight mb-2">
          Security.
        </h1>
        <p className={body}>
          All data is encrypted in transit using TLS 1.3, and we follow
          industry best practices for authentication and authorization.
        </p>

        <h2 className={sectionLabel}>measures</h2>
        <div className="border-t border-hairline">
          {measures.map((m) => (
            <div
              key={m}
              className="flex items-center gap-3 py-2.5 border-b border-hairline text-sm text-muted-foreground"
            >
              <StatusGlyph variant="success" />
              <span>{m}</span>
            </div>
          ))}
        </div>

        <Divider />

        <h2 className={sectionLabel}>responsible disclosure</h2>
        <p className={body}>
          If you believe you have found a vulnerability, report it to{" "}
          <a
            href={`mailto:${process.env.SECURITY_CONTACT_EMAIL || "security@example.com"}`}
            className="text-accent hover:underline"
          >
            {process.env.SECURITY_CONTACT_EMAIL || "security@example.com"}
          </a>
          . We ask that you give us reasonable time to fix the issue before
          public disclosure, avoid privacy violations and data destruction, and
          not exploit it beyond what is needed to demonstrate the issue.
        </p>
      </main>
    </div>
  );
}
