import assert from "node:assert/strict";
import test from "node:test";

import { buildGithubAppAuthorizationURL } from "./github-authorization.ts";

test("builds a GitHub App authorization URL without OAuth scopes", () => {
  const source = new URLSearchParams({
    response_type: "code",
    client_id: "untrusted-client-id",
    state: "state-value",
    scope: "",
    redirect_uri: "https://attacker.example/callback",
    code_challenge_method: "S256",
    code_challenge: "challenge-value",
  });

  const result = buildGithubAppAuthorizationURL(
    source,
    "configured-client-id",
    "https://www.bitrok.tech/api/auth/callback/github",
  );

  assert.equal(result.origin, "https://github.com");
  assert.equal(result.pathname, "/login/oauth/authorize");
  assert.equal(result.searchParams.get("client_id"), "configured-client-id");
  assert.equal(
    result.searchParams.get("redirect_uri"),
    "https://www.bitrok.tech/api/auth/callback/github",
  );
  assert.equal(result.searchParams.get("state"), "state-value");
  assert.equal(result.searchParams.get("code_challenge"), "challenge-value");
  assert.equal(result.searchParams.has("scope"), false);
});
