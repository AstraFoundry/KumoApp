// Lists message keys the pages use that the catalogs lack, and keys a
// translation lacks. Exits 1 when anything is missing.
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join } from "node:path";

const root = new URL("..", import.meta.url).pathname;
const catalogs = Object.fromEntries(
  readdirSync(join(root, "locales"))
    .filter((f) => f.endsWith(".json"))
    .map((f) => [f.slice(0, -5), JSON.parse(readFileSync(join(root, "locales", f), "utf8"))]),
);
const source = catalogs.en;
const used = new Set();
function walk(dir) {
  for (const name of readdirSync(dir)) {
    const path = join(dir, name);
    if (statSync(path).isDirectory()) walk(path);
    else if (/\.tsx?$/.test(name) && !name.endsWith("locales.ts")) {
      for (const m of readFileSync(path, "utf8").matchAll(/\bt\.(\w+)\(/g)) used.add(m[1]);
    }
  }
}
walk(join(root, "src"));
let failed = false;
const missing = [...used].filter((k) => !(k in source)).sort();
if (missing.length) {
  failed = true;
  console.log(`en lacks ${missing.length}:\n  ${missing.join("\n  ")}`);
}
for (const [lang, catalog] of Object.entries(catalogs)) {
  if (lang === "en") continue;
  const lacking = Object.keys(source).filter((k) => !k.startsWith("@") && !(k in catalog));
  if (lacking.length) {
    failed = true;
    console.log(`${lang} lacks ${lacking.length}:\n  ${lacking.join("\n  ")}`);
  }
}
const unused = Object.keys(source).filter((k) => !k.startsWith("@") && !used.has(k) && !/^(tray|notify|mode)/.test(k));
if (unused.length) console.log(`unused in pages (fine if Go uses them): ${unused.join(", ")}`);
process.exit(failed ? 1 : 0);
