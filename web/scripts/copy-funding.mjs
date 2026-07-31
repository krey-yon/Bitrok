import { copyFileSync, mkdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..", "..");
const out = join(root, "web", "public", "funding.json");
mkdirSync(dirname(out), { recursive: true });
copyFileSync(join(root, "funding.json"), out);
