import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import test from "node:test";

import { GET as getPowerShellInstaller } from "../app/install.ps1/route.ts";
import { GET as getShellInstaller } from "../app/install/route.ts";

test("served shell installer has valid syntax and checksum verification", async () => {
  const script = await (await getShellInstaller()).text();

  assert.match(script, /ARCHIVE_NAME="bitrok_\$\{OS\}_\$\{ARCH\}\.tar\.gz"/);
  assert.match(script, /download "\$TMP_DIR\/checksums\.txt"/);
  execFileSync("sh", ["-n"], { input: script });
});

test("served PowerShell installer preserves security-critical regexes", async () => {
  const script = await (await getPowerShellInstaller()).text();

  assert.ok(script.includes("'^v\\d+\\.\\d+\\.\\d+(?:[-+][0-9A-Za-z.-]+)?$'"));
  assert.ok(script.includes('"^[0-9a-fA-F]{64}\\s+$([regex]::Escape($archiveName))$"'));
  assert.ok(script.includes("-split '\\s+'"));
});
