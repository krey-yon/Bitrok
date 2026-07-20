import assert from "node:assert/strict";
import test from "node:test";

import {
  fetchVerifiedGithubIdentity,
  selectVerifiedGithubEmail,
} from "./github-identity.ts";

test("prefers the verified primary GitHub email", () => {
  assert.equal(
    selectVerifiedGithubEmail([
      { email: "other@example.com", verified: true },
      { email: "Primary@Example.com", primary: true, verified: true },
    ]),
    "primary@example.com",
  );
});

test("never accepts an unverified GitHub email", () => {
  assert.equal(
    selectVerifiedGithubEmail([{ email: "unverified@example.com", primary: true, verified: false }]),
    null,
  );
});

test("fetches GitHub identity with required API headers", async () => {
  const requests: Array<{ url: string; headers: Headers }> = [];
  const githubFetch = async (
    input: string | URL | Request,
    init?: RequestInit,
  ): Promise<Response> => {
    const url = String(input);
    requests.push({ url, headers: new Headers(init?.headers) });
    const body = url.endsWith("/emails")
      ? [{ email: "Verified@Example.com", primary: true, verified: true }]
      : { id: 42, login: "verified-user", name: null, avatar_url: "https://example.com/avatar" };
    return new Response(JSON.stringify(body), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  };

  const identity = await fetchVerifiedGithubIdentity("oauth-token", githubFetch);

  assert.equal(identity?.user.id, "42");
  assert.equal(identity?.user.email, "verified@example.com");
  assert.equal(identity?.user.name, "verified-user");
  assert.deepEqual(requests.map((request) => request.url), [
    "https://api.github.com/user",
    "https://api.github.com/user/emails",
  ]);
  for (const request of requests) {
    assert.equal(request.headers.get("Authorization"), "Bearer oauth-token");
    assert.equal(request.headers.get("User-Agent"), "Bitrok-Web");
    assert.equal(request.headers.get("X-GitHub-Api-Version"), "2022-11-28");
  }
});

test("preserves a valid profile when no verified email is available", async () => {
  const githubFetch = async (
    input: string | URL | Request,
  ): Promise<Response> => {
    const body = String(input).endsWith("/emails")
      ? [{ email: "unverified@example.com", primary: true, verified: false }]
      : { id: 42, login: "verified-user", name: "User" };
    return new Response(JSON.stringify(body), { status: 200 });
  };

  const identity = await fetchVerifiedGithubIdentity("oauth-token", githubFetch);
  assert.equal(identity?.user.id, "42");
  assert.equal(identity?.user.email, null);
  assert.equal(identity?.user.emailVerified, false);
});
