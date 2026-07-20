import type { Metadata } from "next";
import { Check, ShieldCheck } from "lucide-react";
import { LegalSection, LegalShell } from "@/app/components/legal-shell";

export const metadata: Metadata = { title: "Security", description: "Bitrok security practices and vulnerability disclosure policy." };

const measures = ["Encrypted public connections", "Verified GitHub identities", "Non-root relay runtime", "Dependency monitoring", "Content Security Policy headers", "API rate limiting"];

export default function SecurityPage() {
  const email = process.env.SECURITY_CONTACT_EMAIL || "security@bitrok.tech";
  return <LegalShell eyebrow="Trust / Security" title="Secure by default. Open to scrutiny." intro="Bitrok treats every tunnel as a security boundary and every credential as sensitive infrastructure.">
    <div className="mb-8 flex items-center gap-4 rounded-xl border border-success/25 bg-success/8 p-5"><div className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-success/12 text-success"><ShieldCheck className="size-5" aria-hidden /></div><div><strong className="block">Security is part of the protocol</strong><p className="mt-1 text-sm text-muted-foreground">Authentication, encrypted transport, and scoped access are built into the connection flow.</p></div></div>
    <LegalSection number="01" title="Security Measures"><ul className="grid gap-3 sm:grid-cols-2">{measures.map((measure) => <li key={measure} className="flex items-center gap-3 rounded-lg border border-hairline bg-background/55 p-3 text-sm text-foreground"><Check className="size-3.5 shrink-0 text-success" aria-hidden />{measure}</li>)}</ul></LegalSection>
    <LegalSection number="02" title="Responsible Disclosure"><p>If you believe you found a vulnerability, report it to <a href={`mailto:${email}`} className="font-medium text-foreground underline decoration-accent decoration-2 underline-offset-4 hover:text-accent">{email}</a>. Give us reasonable time to investigate before public disclosure, avoid privacy violations or data destruction, and do not exploit the issue beyond what is necessary to demonstrate it.</p></LegalSection>
  </LegalShell>;
}
