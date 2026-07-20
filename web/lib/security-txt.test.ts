import assert from "node:assert/strict";
import test from "node:test";

import { buildSecurityTxt } from "./security-txt.ts";

test("security.txt has a real contact, canonical links, and a fresh expiry", async () => {
  const before = Date.now();
  const body = buildSecurityTxt("https://bitrok.tech", "security@bitrok.tech", before);
  const expiry = body.match(/^Expires: (.+)$/m)?.[1];

  assert.match(body, /^Contact: mailto:security@bitrok\.tech$/m);
  assert.match(body, /^Canonical: https:\/\/bitrok\.tech\/\.well-known\/security\.txt$/m);
  assert.ok(expiry, "security.txt must include an expiry");
  assert.ok(Date.parse(expiry) >= before + 179 * 24 * 60 * 60 * 1000);
  assert.ok(Date.parse(expiry) <= before + 181 * 24 * 60 * 60 * 1000);
});
