import { NextRequest, NextResponse } from "next/server";
import { getAuthBaseURL } from "@/lib/app-url";
import { buildGithubAppAuthorizationURL } from "@/lib/github-authorization";

export function GET(request: NextRequest) {
  const clientId = process.env.GITHUB_CLIENT_ID;
  if (!clientId) {
    return NextResponse.json({ error: "GitHub sign-in is unavailable" }, { status: 503 });
  }

  const state = request.nextUrl.searchParams.get("state");
  if (!state) {
    return NextResponse.json({ error: "Missing OAuth state" }, { status: 400 });
  }

  const callbackURL = `${getAuthBaseURL()}/api/auth/callback/github`;
  const authorizationURL = buildGithubAppAuthorizationURL(
    request.nextUrl.searchParams,
    clientId,
    callbackURL,
  );

  const response = NextResponse.redirect(authorizationURL);
  response.headers.set("Cache-Control", "no-store");
  return response;
}
