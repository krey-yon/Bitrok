import { requireAuth } from "@/lib/session";
import Link from "next/link";
import { CliTokenGenerator } from "./cli-token-generator";
import { Logo } from "@/components/ui/logo";
import { Eyebrow } from "@/components/ui/eyebrow";

export default async function CliTokenPage() {
  await requireAuth();

  return (
    <div className="min-h-full flex flex-col">
      <nav className="sticky top-0 z-50 border-b border-hairline bg-background/80 backdrop-blur">
        <div className="max-w-3xl mx-auto px-6 h-12 flex items-center justify-between text-sm">
          <Link href="/dashboard" className="font-mono">
            <Logo />
          </Link>
          <Link
            href="/dashboard"
            className="text-muted-foreground hover:text-foreground transition-colors font-mono text-xs"
          >
            ← dashboard
          </Link>
        </div>
      </nav>

      <main className="flex-1 max-w-3xl mx-auto px-6 py-14 w-full">
        <Eyebrow ornament="·">cli token</Eyebrow>
        <h1 className="mt-3 text-3xl font-semibold tracking-tight mb-2">
          CLI token.
        </h1>
        <p className="text-sm text-muted font-mono mb-12">
          for the bitrok cli · valid 30 days
        </p>

        <CliTokenGenerator />
      </main>
    </div>
  );
}
