import assert from "node:assert/strict";
import test from "node:test";

import { decideUsernameClaim } from "./username-claim.ts";

test("allows the first username claim", () => {
  assert.deepEqual(decideUsernameClaim(null, "vikas"), { action: "claim" });
});

test("treats the same username claim as idempotent", () => {
  assert.deepEqual(decideUsernameClaim("vikas", "vikas"), {
    action: "unchanged",
    username: "vikas",
  });
});

test("rejects changing an existing username", () => {
  assert.deepEqual(decideUsernameClaim("vikas", "someone-else"), { action: "reject" });
});
