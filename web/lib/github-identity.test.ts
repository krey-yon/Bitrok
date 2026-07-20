import assert from "node:assert/strict";
import test from "node:test";

import { selectVerifiedGithubEmail } from "./github-identity.ts";

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
