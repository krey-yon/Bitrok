import type { Metadata } from "next";
import { LegalSection, LegalShell } from "@/app/components/legal-shell";

export const metadata: Metadata = { title: "Privacy Policy", description: "Bitrok privacy policy and data handling practices." };

export default function PrivacyPage() {
  const email = process.env.PRIVACY_CONTACT_EMAIL || "privacy@example.com";
  return <LegalShell eyebrow="Legal / Privacy" title="Privacy, without the fog." intro="A plain-language overview of what Bitrok collects, why it is needed, and the control you keep. Last revised July 19, 2026.">
    <LegalSection number="01" title="Information We Collect"><p>We collect information you provide when creating an account—such as your name, email, and authentication credentials—plus tunnel configuration and service usage needed to operate Bitrok.</p></LegalSection>
    <LegalSection number="02" title="How We Use It"><p>We use this information to provide, maintain, secure, and improve the service; communicate with you; and detect fraud or abuse.</p></LegalSection>
    <LegalSection number="03" title="Data Security"><p>We apply technical and organizational safeguards designed to protect personal data from unauthorized access, alteration, disclosure, or destruction.</p></LegalSection>
    <LegalSection number="04" title="Data Retention"><p>We retain personal data only for as long as necessary to provide the service and meet legal, accounting, or reporting obligations.</p></LegalSection>
    <LegalSection number="05" title="Your Rights"><p>You may request access to, correction of, or deletion of your personal data, and may restrict or object to certain processing where applicable.</p></LegalSection>
    <LegalSection number="06" title="Contact"><p>Questions about your data? Email <a href={`mailto:${email}`} className="font-medium text-foreground underline decoration-accent decoration-2 underline-offset-4 hover:text-accent">{email}</a>.</p></LegalSection>
  </LegalShell>;
}
