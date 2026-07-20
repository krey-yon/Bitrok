import { NextResponse } from "next/server";
import { buildSecurityTxt } from "@/lib/security-txt";

export async function GET() {
  const baseUrl = (process.env.NEXT_PUBLIC_APP_URL || "https://bitrok.tech").replace(/\/+$/, "");
  const contact = process.env.SECURITY_CONTACT_EMAIL || "security@bitrok.tech";

  return new NextResponse(buildSecurityTxt(baseUrl, contact), {
    headers: {
      "Cache-Control": "public, max-age=86400",
      "Content-Type": "text/plain; charset=utf-8",
    },
  });
}
