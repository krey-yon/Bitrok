import { NextResponse } from "next/server";

export async function GET() {
  const baseUrl = process.env.NEXT_PUBLIC_APP_URL || "";
  const contact = process.env.SECURITY_CONTACT_EMAIL || "security@example.com";

  const lines = [
    `Contact: mailto:${contact}`,
    `Expires: 2026-12-31T23:59:59Z`,
  ];

  if (baseUrl) {
    lines.push(`Acknowledgments: ${baseUrl}/security`);
    lines.push(`Canonical: ${baseUrl}/.well-known/security.txt`);
    lines.push(`Policy: ${baseUrl}/security`);
  }

  lines.push("Preferred-Languages: en");

  return new NextResponse(lines.join("\n"), {
    headers: {
      "Content-Type": "text/plain; charset=utf-8",
    },
  });
}
