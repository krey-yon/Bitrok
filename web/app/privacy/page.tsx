import type { Metadata } from "next";
import Link from "next/link";
import { Logo } from "@/components/ui/logo";
import { Eyebrow } from "@/components/ui/eyebrow";
import { Divider } from "@/components/ui/divider";

export const metadata: Metadata = {
  title: "Privacy Policy",
  description: "Bitrok privacy policy and data handling practices.",
};

const sectionLabel =
  "font-mono text-xs uppercase tracking-[0.22em] text-muted-foreground mt-10 mb-3";
const body = "text-sm text-muted-foreground leading-relaxed mb-4";

export default function PrivacyPage() {
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
        <Eyebrow ornament="·">legal</Eyebrow>
        <h1 className="mt-3 text-3xl font-semibold tracking-tight mb-2">
          Privacy policy.
        </h1>
        <p className="text-sm text-muted font-mono mb-4">
          last updated:{" "}
          {new Date().toLocaleDateString("en-US", {
            year: "numeric",
            month: "long",
            day: "numeric",
          })}
        </p>

        <h2 className={sectionLabel}>01 · information we collect</h2>
        <p className={body}>
          Information you provide when you create an account — name, email,
          authentication credentials — plus information about your tunnels and
          usage patterns.
        </p>

        <Divider />

        <h2 className={sectionLabel}>02 · how we use it</h2>
        <p className={body}>
          To provide, maintain, and improve the service, to communicate with
          you, and to detect and prevent fraud and abuse.
        </p>

        <Divider />

        <h2 className={sectionLabel}>03 · data security</h2>
        <p className={body}>
          We implement appropriate technical and organizational measures to
          protect your personal data against unauthorized access, alteration,
          disclosure, or destruction.
        </p>

        <Divider />

        <h2 className={sectionLabel}>04 · data retention</h2>
        <p className={body}>
          We retain personal data only as long as necessary for the purposes it
          was collected for, including legal, accounting, or reporting
          requirements.
        </p>

        <Divider />

        <h2 className={sectionLabel}>05 · your rights</h2>
        <p className={body}>
          You may access, correct, or delete your personal data, and may
          restrict or object to certain processing of it.
        </p>

        <Divider />

        <h2 className={sectionLabel}>06 · contact</h2>
        <p className={body}>
          Questions?{" "}
          <a
            href={`mailto:${process.env.PRIVACY_CONTACT_EMAIL || "privacy@example.com"}`}
            className="text-accent hover:underline"
          >
            {process.env.PRIVACY_CONTACT_EMAIL || "privacy@example.com"}
          </a>
        </p>
      </main>
    </div>
  );
}
